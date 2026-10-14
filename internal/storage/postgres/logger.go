package postgres

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
)

type pgxLogger struct {
	log *slog.Logger
}

func newPgxLogger(log *slog.Logger) *pgxLogger {
	return &pgxLogger{log: log}
}

func (l *pgxLogger) Log(
	ctx context.Context,
	level tracelog.LogLevel,
	msg string,
	data map[string]any) {
	if sql, ok := data["sql"].(string); ok {
		data["sql"] = normalizeSQL(sql)
	}
	args := make([]any, 0, len(data)*2)
	for k, v := range data {
		args = append(args, k, v)
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		l.log.Debug(msg, args...)
	case tracelog.LogLevelInfo:
		l.log.Info(msg, args...)
	case tracelog.LogLevelWarn:
		l.log.Warn(msg, args...)
	case tracelog.LogLevelError:
		l.log.Error(msg, args...)
	default:
		l.log.Debug(msg, args...)
	}
}

func normalizeSQL(sql string) string {
	return strings.Join(strings.Fields(sql), " ")

	// sql = strings.ReplaceAll(sql, "\t", " ")
	// sql = strings.ReplaceAll(sql, "\n", " ")
	// for strings.Contains(sql, "  ") {
	// 	sql = strings.ReplaceAll(sql, "  ", " ")
	// }

	// return strings.TrimSpace(sql)
}
