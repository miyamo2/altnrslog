package altnrslog

import (
	"context"
	"errors"
	"log/slog"
)

type loggerKey struct{}

var (
	// ErrInvalidHandler is returned when the handler is not a TransactionalHandler.
	ErrInvalidHandler = errors.New("invalid handler")
	// ErrNotStored is returned when the logger is not stored in context.Context.
	ErrNotStored = errors.New("logger not stored")
)

// FromContext returns *slog.Logger with *TransactionalHandler, stored in the context.Context.
// If it does not exist, return ErrNotStored.
func FromContext(ctx context.Context) (*slog.Logger, error) {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		return nil, ErrNotStored
	}
	_, ok = logger.Handler().(*TransactionalHandler)
	if !ok {
		return nil, ErrNotStored
	}
	return logger, nil
}

// StoreToContext stores the *slog.Logger in context.Context.
// Logger must be set to *TransactionalHandler in the Handler
func StoreToContext(ctx context.Context, logger *slog.Logger) (context.Context, error) {
	_, ok := logger.Handler().(*TransactionalHandler)
	if !ok {
		return ctx, ErrInvalidHandler
	}
	return context.WithValue(ctx, loggerKey{}, logger), nil
}
