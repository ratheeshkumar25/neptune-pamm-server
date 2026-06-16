// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/logger/logger.go
// Role: Structured logging
// Description: A thin wrapper over the standard library log/slog. Produces structured
// JSON logs (text in dev), with a configurable level and an optional source location.
// No third-party dependency.

package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Format selects the output encoding.
type Format string

const (
	FormatJSON Format = "json" // structured JSON — use in production
	FormatText Format = "text" // human-readable — use in local dev
)

// Options configures a logger.
type Options struct {
	Level     string // debug | info | warn | error (default: info)
	Format    Format // json | text (default: json)
	AddSource bool   // include file:line of the call site
}

// New builds a *slog.Logger from the given options.
func New(opts Options) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{
		Level:     parseLevel(opts.Level),
		AddSource: opts.AddSource,
	}

	var handler slog.Handler
	switch strings.ToLower(string(opts.Format)) {
	case string(FormatText):
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}

	return slog.New(handler)
}

// Init builds a logger from opts and installs it as the slog default, so the
// package-level slog.Info/Warn/Error helpers route through it. Returns the logger.
func Init(opts Options) *slog.Logger {
	l := New(opts)
	slog.SetDefault(l)
	return l
}

// Default returns the current slog default logger.
func Default() *slog.Logger {
	return slog.Default()
}

// parseLevel maps a level string to slog.Level, defaulting to Info.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// --- request-scoped helpers ----------------------------------------------------

// ctxKey is an unexported type to avoid context key collisions.
type ctxKey struct{}

// WithContext stores a logger in the context (e.g. one enriched with a trace id).
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext retrieves a logger previously stored with WithContext, or the
// default logger if none is present.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
