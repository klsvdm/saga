package saga

import "time"

type Option func(opts *options)

type retry struct {
	Attempts int
	Delay    time.Duration
}

type options struct {
	Retry retry
}

func WithRetry(attemts int, delay time.Duration) Option {
	return func(opts *options) {
		opts.Retry = retry{
			Attempts: attemts,
			Delay:    delay,
		}
	}
}

var defaultOptions = options{
	Retry: retry{},
}

func applyOptions(opts ...Option) *options {
	os := defaultOptions

	for _, opt := range opts {
		opt(&os)
	}

	return &os
}
