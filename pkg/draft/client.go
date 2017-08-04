package draft

import (
	"fmt"
	"github.com/Azure/draft/pkg/build"
	"github.com/Azure/draft/pkg/rpc"
	"github.com/Azure/draft/pkg/version"
	"golang.org/x/net/context"
	"io"
)

type ClientConfig struct {
	ServerAddr string
	ServerHost string
	Stdout     io.Writer
	Stderr     io.Writer
}

type Client struct {
	cfg *ClientConfig
	rpc rpc.Client
}

func NewClient(cfg *ClientConfig) *Client {
	opts := []rpc.ClientOpt{rpc.WithServerAddr(cfg.ServerAddr)}
	return &Client{cfg: cfg, rpc: rpc.NewClient(opts...)}
}

func (c *Client) Version(ctx context.Context) (*version.Version, error) {
	return c.rpc.Version(ctx)
}

func (c *Client) Up(ctx context.Context, app *build.Context) error {
	req := &rpc.UpRequest{
		Namespace:  app.Env.Namespace,
		AppName:    app.Env.Name,
		Chart:      app.Chart,
		Values:     app.Values,
		AppArchive: &rpc.AppArchive{app.SrcName, app.Archive},
	}
	msgc := make(chan *rpc.UpSummary)
	errc := make(chan error)
	go func() {
		if err := c.rpc.UpBuild(ctx, req, msgc); err != nil {
			errc <- err
		}
		close(errc)
	}()
	for msgc != nil || errc != nil {
		select {
		case msg, ok := <-msgc:
			if !ok {
				msgc = nil
				continue
			}
			fmt.Fprintf(c.cfg.Stdout, "\r%s: %s\n", msg.StageDesc, msg.StatusText)
		case err, ok := <-errc:
			if !ok {
				errc = nil
				continue
			}
			return fmt.Errorf("error running draft up: %v", err)
		}
	}
	return nil
}
