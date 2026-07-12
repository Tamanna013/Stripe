// Package logging provides FlowGuard's canonical structured logging wrapper
// around log/slog, enforcing the shared JSON log schema documented in the
// FlowGuard architecture spec (service, level, timestamp, message, attributes,
// and optional trace_id/span_id for request-scoped logs).
package logging

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey string

const (
	traceIDKey ctxKey = "trace_id"
	spanIDKey  ctxKey = "span_id"
)

// Logger wraps slog.Logger to enforce FlowGuard's canonical log schema.
type Logger struct {
	base        *slog.Logger
	serviceName string
}

// New creates a Logger for the given service name, writing JSON to stdout.
func New(serviceName string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{
		base:        slog.New(handler),
		serviceName: serviceName,
	}
}

// WithTraceContext returns a context carrying trace/span IDs so subsequent
// log calls using this context automatically include them.
func WithTraceContext(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, spanIDKey, spanID)
	return ctx
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, attrs map[string]any) {
	args := []any{"service", l.serviceName}
	if attrs == nil {
		attrs = map[string]any{}
	}
	if tid, ok := ctx.Value(traceIDKey).(string); ok && tid != "" {
		args = append(args, "trace_id", tid)
	}
	if sid, ok := ctx.Value(spanIDKey).(string); ok && sid != "" {
		args = append(args, "span_id", sid)
	}
	args = append(args, "attributes", attrs)
	l.base.Log(ctx, level, msg, args...)
}

// Info logs at info level with the given message and structured attributes.
func (l *Logger) Info(ctx context.Context, msg string, attrs map[string]any) {
	l.log(ctx, slog.LevelInfo, msg, attrs)
}

// Warn logs at warn level with the given message and structured attributes.
func (l *Logger) Warn(ctx context.Context, msg string, attrs map[string]any) {
	l.log(ctx, slog.LevelWarn, msg, attrs)
}

// Error logs at error level with the given message, error, and structured attributes.
func (l *Logger) Error(ctx context.Context, msg string, err error, attrs map[string]any) {
	if attrs == nil {
		attrs = map[string]any{}
	}
	if err != nil {
		attrs["error"] = err.Error()
	}
	l.log(ctx, slog.LevelError, msg, attrs)
}

// Debug logs at debug level with the given message and structured attributes.
func (l *Logger) Debug(ctx context.Context, msg string, attrs map[string]any) {
	l.log(ctx, slog.LevelDebug, msg, attrs)
}
