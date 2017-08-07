package rpc

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/Azure/draft/pkg/version"
)

type serverImpl struct {
	opts serverOpts
	h    Handler
	lis  net.Listener
	srv  *grpc.Server
	done int32
}

func newServerImpl(opts ...ServerOpt) *serverImpl {
	var s serverImpl
	for _, opt := range opts {
		opt(&s.opts)
	}
	return &s
}

// Server implements rpc.Server.Serve
func (s *serverImpl) Serve(lis net.Listener, h Handler) error {
	var opts []grpc.ServerOption
	// TODO: If TLS load keys and such
	s.h = h
	s.lis = lis
	s.srv = grpc.NewServer(opts...)
	RegisterDraftServer(s.srv, s)
	return s.srv.Serve(s.lis)
}

// Stop stops the server. Stop returns true if the server was already stopped, false otherwise.
//
// Stop implements rpc.Server.Stop
func (s *serverImpl) Stop() bool {
	if atomic.CompareAndSwapInt32(&s.done, 0, 1) {
		s.srv.GracefulStop()
		s.lis.Close()
		return false
	}
	return true
}

// GetVersion returns the current version of the server.
func (s *serverImpl) GetVersion(ctx context.Context, _ *empty.Empty) (*Version, error) {
	v := version.New()
	return &Version{SemVer: v.SemVer, GitCommit: v.GitCommit, GitTreeState: v.GitTreeState}, nil
}

// UpStream accepts a stream of UpMessages each representing a separate draft up.
// This is the rpc invoked by the draft client when doing a draft up with watch
// enabled.
//
// UpStream implements DraftServer.UpStream
func (s *serverImpl) UpStream(stream Draft_UpStreamServer) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error)

	var wg sync.WaitGroup
	var msg *UpMessage

	// cancel and wait for goroutines to finish.
	defer func() {
		cancel()
		close(errc)
		wg.Wait()
	}()

	// wait for either a cancel or err.
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case err = <-errc:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()

	for {
		switch msg, err = stream.Recv(); {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		}
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, req *UpRequest) {
			defer wg.Done()
			for summary := range s.h.Up(ctx, req) {
				resp := &UpMessage{&UpMessage_UpSummary{summary}}
				if err := stream.Send(resp); err != nil {
					errc <- fmt.Errorf("server: failed to send response: %v", err)
				}
			}
		}(ctx, &wg, msg.GetUpRequest())
	}
	return
}

// UpBuild returns a stream of the summaries within for a given draft upload.
//
// UpBuild implements DraftServer.UpBuild
func (s *serverImpl) UpBuild(msg *UpMessage, stream Draft_UpBuildServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for summary := range s.h.Up(ctx, msg.GetUpRequest()) {
		resp := &UpMessage{&UpMessage_UpSummary{summary}}
		if err := stream.Send(resp); err != nil {
			return fmt.Errorf("server: failed to send response: %v", err)
		}
	}
	return nil
}
