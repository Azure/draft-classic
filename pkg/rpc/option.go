package rpc

// clientOpts specifies the union of all configurable
// options an rpc.Client accepts to communicate with
// the draft rpc.Server.
type clientOpts struct {
	addr string
	host string
}

// WithServerAddr sets the draft server address
// the client should dial when invoking an rpc.
func WithServerAddr(addr string) ClientOpt {
	return func(opts *clientOpts) {
		opts.addr = addr
	}
}

// WithServerHost sets the draft server host
// the client should use when invoking an rpc.
func WithServerHost(host string) ClientOpt {
	return func(opts *clientOpts) {
		opts.host = host
	}
}

// serverOpts specifies the union of all configurable
// options an rpc.Server accepts, e.g. TLS config.
type serverOpts struct{}
