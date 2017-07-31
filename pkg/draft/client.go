package draft

import (
	"fmt"
	"io"
	"golang.org/x/net/context"
	"github.com/Azure/draft/pkg/version"
	"github.com/Azure/draft/pkg/build"
	"github.com/Azure/draft/pkg/rpc"
	"github.com/golang/protobuf/ptypes/any"
)

type ClientConfig struct {
	ServerAddr string
	ServerHost string
	Stdout 	   io.Writer
	Stderr 	   io.Writer
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
		Namespace: app.Env.Namespace,
		Chart:     app.Chart,
		Values:    app.Values,
		Files:     []*any.Any{{app.Source.Name, app.Source.File}},
	}
	if app.Env.Watch {
		fmt.Println("TODO: watch")	
	}

	results, err := c.rpc.UpBuild(ctx, req)
	if err != nil {
		return fmt.Errorf("error running draft up: %v", err)
	}
	for summary := range results {
		fmt.Fprintf(c.cfg.Stdout, "%s: %q\n", summary.StageName, summary.StatusText)
	}
	return nil
}