package httpapi

import (
	"log/slog"

	"github.com/merloot/market-data/internal/httpapi/debug"
	v1 "github.com/merloot/market-data/internal/httpapi/v1"
)

type Deps struct {
	Debug *debug.Deps
	V1    *v1.Deps
	Log   *slog.Logger
}
