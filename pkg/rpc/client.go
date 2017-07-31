package rpc

import (
	"fmt"
	"io"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/Azure/draft/pkg/version"
)

type clientImpl struct {
	opts clientOpts
}

func newClientImpl(opts ...ClientOpt) Client {
	var c clientImpl
	for _, opt := range opts {
		opt(&c.opts)
	}
	return &c
}

// Version implementes rpc.Client.Version
func (c *clientImpl) Version(ctx context.Context) (v *version.Version, err error) {
	err = c.Connect(func(ctx context.Context, rpc DraftClient) error {
		var r *Version
		if r, err = rpc.GetVersion(ctx, &empty.Empty{}); err != nil {
			return fmt.Errorf("error getting server version: %v", err)
		}
		v = &version.Version{SemVer: r.SemVer, GitCommit: r.GitCommit, GitTreeState: r.GitTreeState}
		return nil
	})
	return
}

// UpBuild implementes rpc.Client.UpBuild
func (c *clientImpl) UpBuild(ctx context.Context, msg *UpRequest) (<-chan *UpSummary, error) {
	ret := make(chan *UpSummary)
	err := c.Connect(func(ctx context.Context, rpc DraftClient) (err error) {
		recv := make(chan *UpMessage)
		errc := make(chan error)
		go func() {
			if err = up_build(ctx, rpc, msg, recv); err != nil {
				errc<- err
			}
			close(recv)
			close(errc)
		}()
		go func() {
			for resp := range recv {
				fmt.Printf("summary:\n\tstage_name: %s\n\tstatus_text: %s\n\tstatus_code: %d\n", 
					resp.GetUpSummary().GetStageName(),
					resp.GetUpSummary().GetStatusText(),
					resp.GetUpSummary().GetStatusCode())
				if summary := resp.GetUpSummary(); summary != nil {
					ret<- summary
				}
			}
			close(ret)
		}()
		return <-errc
	})
	return ret, err
}

// UpStream implementes rpc.Client.UpStream
func (c *clientImpl) UpStream(apictx context.Context, msgc <-chan *UpRequest) (<-chan *UpSummary, error)  {
	ret := make(chan *UpSummary)
	err := c.Connect(func(ctx context.Context, rpc DraftClient) error {
		recv := make(chan *UpMessage)
		go func() {
			for resp := range recv {
				fmt.Printf("summary:\n\tstage_name: %s\n\tstatus_text: %s\n\tstatus_code: %d\n", 
					resp.GetUpSummary().GetStageName(),
					resp.GetUpSummary().GetStatusText(),
					resp.GetUpSummary().GetStatusCode())
				if summary := resp.GetUpSummary(); summary != nil {
					ret<- summary
				}
			}
			close(ret)
		}()
		return up_stream(ctx, rpc, msgc, recv)
	})
	return ret, err
}

// Connect connects the DraftClient to the DraftServer.
func (c *clientImpl) Connect(fn func(context.Context, DraftClient) error) (err error) {
	var conn *grpc.ClientConn
	if conn, err = grpc.Dial(c.opts.addr, grpc.WithInsecure()); err != nil {
		return fmt.Errorf("failed to dial %q: %v", c.opts.addr, err)
	}
	defer conn.Close()
	client := NewDraftClient(conn)
	rpcctx := context.Background()
	return fn(rpcctx, client)
}

func up_build(ctx context.Context, rpc DraftClient, msg *UpRequest, recv chan<- *UpMessage) error {
	s, err := rpc.UpBuild(ctx, &UpMessage{&UpMessage_UpRequest{msg}})
	if err != nil {
		return fmt.Errorf("rpc error handling up_build: %v", err)
	}
	for {
		f, err := s.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("rpc error handling up_build recv: %v", err)
		}
		select {
			case <-ctx.Done():
				return nil
			case recv<- f:
				// nothing to do
			default:
		}
	}
}

func up_stream(ctx context.Context, rpc DraftClient, send <-chan *UpRequest, recv chan<- *UpMessage) error {
	stream, err := rpc.UpStream(ctx)
	if err != nil {
		return fmt.Errorf("rpc error handling up_stream: %v", err)
	}
	done := make(chan struct{})
	errc := make(chan error)
	defer func() {
		stream.CloseSend()
		<-done
		close(recv)
		close(errc)
	}()
	go func() {
		for {
			m, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				errc <- fmt.Errorf("failed to receive a summary: %v", err)
				return
			}
			recv<- m
		}
	}()
	for {
		select {
			case msg, ok := <-send:
				if !ok { return nil }		
				fmt.Printf("client: sending: %v\n", msg)
				req := &UpMessage{&UpMessage_UpRequest{msg}}
				if err := stream.Send(req); err != nil {
					return fmt.Errorf("failed to send an up message: %v", err)
				}
			case err := <-errc:
				return err
		}
	}
}