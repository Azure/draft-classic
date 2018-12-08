package tasks

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	// ErrNoTaskFile is the error message shown when no Task file is found
	ErrNoTaskFile = errors.New(".draft-tasks.toml not found")
)

const (
	// PreUp are the kind of tasks to be executed in preparation to an up command
	PreUp = "PreUp"
	// PostUp are the kind of tasks to be executed after an up command and post-deploy tasks have executed
	PostUp = "PostUp"
	// PostDeploy are the kind of tasks to be executed after a deploy command
	PostDeploy = "PostDeploy"
	// PostDelete are the kind of tasks to be executed after a delete command
	PostDelete = "PostDelete"
)

var (
	// reEnvironmentVariable matches environment variables embedded in
	// strings. Only simple expressions ($FOO) are supported. Variables
	// may escaped to avoid interpolation, in the form $$FOO or \$FOO
	reEnvironmentVariable = regexp.MustCompile(`([\\\$]?\$[a-zA-Z_][a-zA-Z0-9_]*)`)
)

// Runner runs the given command. An alternative to DefaultRunner can
// be used in tests.
type Runner func(c *exec.Cmd) error

// DefaultRunner runs the given command
var DefaultRunner = func(c *exec.Cmd) error { return c.Run() }

// Tasks represents the different kinds of tasks read from Tasks' file
type Tasks struct {
	PreUp      map[string]string `toml:"pre-up"`
	PostUp     map[string]string `toml:"post-up"`
	PostDeploy map[string]string `toml:"post-deploy"`
	PostDelete map[string]string `toml:"cleanup"`
}

// Result represents the result of a Task's execution
type Result struct {
	Kind    string
	Command []string
	Pass    bool
	Message string
}

// Load takes a path to file where tasks are defined and loads them in tasks
func Load(path string) (*Tasks, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoTaskFile
		}
		return nil, err
	}

	t := Tasks{}
	if _, err := toml.DecodeFile(path, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

// Run executes a series of tasks of a given kind and returns the list of results
func (t *Tasks) Run(runner Runner, kind, podName string) ([]Result, error) {
	results := []Result{}

	switch kind {
	case PreUp:
		for _, task := range t.PreUp {
			result := executeTask(runner, task, kind)
			results = append(results, result)
		}
	case PostUp:
		for _, task := range t.PostUp {
			result := executeTask(runner, task, kind)
			results = append(results, result)
		}
	case PostDeploy:
		for _, task := range t.PostDeploy {
			cmd := preparePostDeployTask(evaluateArgs(task), podName)
			result := runTask(runner, cmd, kind)
			results = append(results, result)
		}
	case PostDelete:
		for _, task := range t.PostDelete {
			result := executeTask(runner, task, kind)
			results = append(results, result)
		}
	default:
		return results, fmt.Errorf("Task kind: %s not supported", kind)
	}

	return results, nil
}

func executeTask(runner Runner, task, kind string) Result {
	args := evaluateArgs(task)
	cmd := prepareTask(args)
	return runTask(runner, cmd, kind)
}

func runTask(runner Runner, cmd *exec.Cmd, kind string) Result {
	result := Result{Kind: kind, Pass: false}
	result.Command = append([]string{cmd.Path}, cmd.Args[0:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := runner(cmd)
	if err != nil {
		result.Pass = false
		result.Message = err.Error()
		return result
	}
	result.Pass = true

	return result
}

func prepareTask(args []string) *exec.Cmd {
	var cmd *exec.Cmd
	if len(args) < 2 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	return cmd
}

func preparePostDeployTask(args []string, podName string) *exec.Cmd {
	args = append([]string{"exec", podName, "--"}, args[0:]...)
	return exec.Command("kubectl", args[0:]...)
}

func evaluateArgs(task string) []string {
	args := strings.Split(task, " ")
	for i, arg := range args {
		args[i] = reEnvironmentVariable.ReplaceAllStringFunc(arg, func(expr string) string {
			// $$FOO and \$FOO are kept as-is
			if strings.HasPrefix(expr, "$$") || strings.HasPrefix(expr, "\\$") {
				return expr[1:]
			}

			return os.Getenv(expr[1:])
		})
	}
	return args
}
