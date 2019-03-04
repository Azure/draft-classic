package cmdline

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/draft/pkg/builder"
	"github.com/fatih/color"
	"golang.org/x/net/context"
)

var (
	yellow = color.New(color.FgHiYellow, color.BgBlack, color.Bold).SprintFunc()
	green  = color.New(color.FgHiGreen, color.BgBlack, color.Bold).SprintFunc()
	blue   = color.New(color.FgHiBlue, color.BgBlack, color.Underline).SprintFunc()
	cyan   = color.New(color.FgCyan, color.BgBlack).SprintFunc()
	red    = color.New(color.FgHiRed, color.BgBlack).Add(color.Italic).SprintFunc()
)

// cmdline provides a basic cli ui/ux for draft client operations. It handles
// the draft state machine and displays a measure of progress for each draft
// client api invocation.
type cmdline struct {
	ctx  context.Context
	opts options
	done chan struct{}
	once sync.Once
	err  error
}

// Init initializes the cmdline interface.
func (cli *cmdline) Init(rootctx context.Context, opts ...Option) {
	DefaultOpts()(&cli.opts)
	for _, opt := range opts {
		opt(&cli.opts)
	}
	if out := cli.opts.stdout; isTerminal(out) {
		initTerminal(out)
	} else {
		NoColor()(&cli.opts)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cli.ctx = ctx
	cli.done = make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-rootctx.Done():
		case <-cli.Done():
		}
		cancel()
		wg.Done()
	}()
	go func() {
		wg.Wait()
		cli.Stop()
	}()
}

// Done returns a channel that signals the cmdline interface is finished.
func (cli *cmdline) Done() <-chan struct{} { return cli.done }

// Stop notify the cmdline interface internals to finish and performs the necessary cleanup.
func (cli *cmdline) Stop() error {
	cli.once.Do(func() {
		close(cli.done)
	})
	return cli.err
}

// Display provides a UI for the draft client. When performing a draft 'up'
// Display will output a measure of progress for each summary yielded by the
// draft state machine.
func Display(ctx context.Context, app string, summaries <-chan *builder.Summary, opts ...Option) {
	var cli cmdline
	cli.Init(ctx, opts...)

	fmt.Fprintf(cli.opts.stdout, "%s: '%s': %s\n",
		blue("Draft Up Started"),
		cyan(app),
		yellow(cli.opts.buildID),
	)
	ongoing := make(map[string]chan builder.SummaryStatusCode)
	var (
		wg     sync.WaitGroup
		id     string
		failed bool
	)
	defer func() {
		for _, c := range ongoing {
			close(c)
		}
		cli.Stop()
		wg.Wait()

		logText := fmt.Sprintf("%s `%s`\n", blue("Inspect the logs with"), yellow("draft logs ", id))

		if failed {
			fmt.Fprintf(cli.opts.stderr, logText)
		} else {
			fmt.Fprintf(cli.opts.stdout, logText)
		}
	}()
	for {
		select {
		case summary, ok := <-summaries:
			if !ok {
				return
			}
			if id == "" {
				id = summary.BuildID
			}
			if summary.StatusCode == builder.SummaryFailure {
				failed = true
			}
			if ch, ok := ongoing[summary.StageDesc]; !ok {
				ch = make(chan builder.SummaryStatusCode, 1)
				ongoing[summary.StageDesc] = ch
				wg.Add(1)
				go func(desc string, ch chan builder.SummaryStatusCode, wg *sync.WaitGroup) {
					progress(&cli, app, desc, ch)
					delete(ongoing, desc)
					wg.Done()
				}(summary.StageDesc, ch, &wg)
			} else {
				ch <- summary.StatusCode
			}
		case <-cli.Done():
			return
		}
	}
}

func progress(cli *cmdline, app, desc string, codes <-chan builder.SummaryStatusCode) {
	start := time.Now()
	done := make(chan builder.SummaryStatusCode, 1)
	go func() {
		defer close(done)
		for code := range codes {
			if code == builder.SummarySuccess || code == builder.SummaryFailure {
				done <- code
			}
		}
	}()
	m := fmt.Sprintf("%s: %s", cyan(app), yellow(desc))
	s := `-\|/-`
	i := 0
	for {
		select {
		case code := <-done:
			switch code {
			case builder.SummarySuccess:
				fmt.Fprintf(cli.opts.stdout, "\r%s: %s  (%.4fs)\n", cyan(app), passStr(desc, cli.opts.displayEmoji), time.Since(start).Seconds())
				return
			case builder.SummaryFailure:
				fmt.Fprintf(cli.opts.stderr, "\r%s: %s  (%.4fs)\n", cyan(app), failStr(desc, cli.opts.displayEmoji), time.Since(start).Seconds())
				return
			}
		default:
			fmt.Fprintf(cli.opts.stdout, "\r%s %c", m, s[i%len(s)])
			time.Sleep(50 * time.Millisecond)
			i++
		}
	}
}

func passStr(msg string, displayEmoji bool) string {
	return fmt.Sprintf("%s: %s", green(msg), concatStrAndEmoji("SUCCESS", " ⚓ ", displayEmoji))
}

func failStr(msg string, displayEmoji bool) string {
	return fmt.Sprintf("%s: %s", red(msg), concatStrAndEmoji("FAIL", " ❌ ", displayEmoji))
}

func concatStrAndEmoji(text string, emoji string, displayEmoji bool) string {
	var concatStr strings.Builder
	concatStr.WriteString(text)
	if displayEmoji {
		concatStr.WriteString(emoji)
	}
	return concatStr.String()
}
