package postgres

import "log/slog"

type options struct {
	logger      *slog.Logger
	pingRetries int
	metrics     bool
}

func defaultOptions() options {
	return options{
		logger:      slog.Default(),
		pingRetries: 0,
		metrics:     false,
	}
}

type Options func(*options)

func WithLogger(log *slog.Logger) Options {
	return func(o *options) { o.logger = log }
}

func WithPingRetries(n int) Options {
	return func(o *options) { o.pingRetries = n }
}

func WithMetrics() Options {
	return func(o *options) { o.metrics = true }
}
