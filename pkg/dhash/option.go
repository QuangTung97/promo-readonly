package dhash

import "time"

type sessionOptions struct {
	waitLeaseDurations []time.Duration
}

func defaultSessionOptions() sessionOptions {
	return sessionOptions{
		waitLeaseDurations: []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			50 * time.Millisecond,
		},
	}
}

func newSessionOptions(options ...SessionOption) sessionOptions {
	opts := defaultSessionOptions()
	for _, fn := range options {
		fn(&opts)
	}
	return opts
}

// SessionOption ...
type SessionOption func(opts *sessionOptions)

// WithWaitLeaseDurations ...
func WithWaitLeaseDurations(durations []time.Duration) SessionOption {
	return func(opts *sessionOptions) {
		opts.waitLeaseDurations = durations
	}
}
