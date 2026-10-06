package asynq

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

type asynqLogger struct {
	log *slog.Logger
}

func newAsynqLogger(log *slog.Logger) asynq.Logger {
	return &asynqLogger{log: log}
}

func (l *asynqLogger) Debug(args ...any) { l.log.Debug("asynq", "args", args) }
func (l *asynqLogger) Info(args ...any)  { l.log.Info("asynq", "args", args) }
func (l *asynqLogger) Warn(args ...any)  { l.log.Warn("asynq", "args", args) }
func (l *asynqLogger) Error(args ...any) { l.log.Error("asynq", "args", args) }
func (l *asynqLogger) Fatal(args ...any) { l.log.Error("asynq fatal", "args", args) }
