package draft

type ClientOpt func(*clientOpts)

type clientOpts struct {
	logLimit int64
}

func defaultClientOpts() *clientOpts {
	return &clientOpts{
		logLimit: 0,
	}
}

func WithLogsLimit(limit int64) ClientOpt {
	return func(opts *clientOpts) {
		if limit > 0 {
			opts.logLimit = limit
		}
	}
}
