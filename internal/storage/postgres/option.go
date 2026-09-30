package postgres

import "log/slog"

type options struct {
	logger      *slog.Logger
	pingRetries int
}

func defaultOptions() options {
	return options{
		logger:      slog.Default(),
		pingRetries: 0,
	}
}

type Options func(*options)

func WithLogger(log *slog.Logger) Options {
	return func(o *options) { o.logger = log }
}

func WithPingRetries(n int) Options {
	return func(o *options) { o.pingRetries = n }
}
