// Package closer provides wrappers around the close function
// to check for errors it returns (to make linter happy)
package closer

import (
	"context"
	"io"
	"log/slog"
)

// CloseOrLog closes the object or, if it fails, logs an error
func CloseOrLog(log *slog.Logger, closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Error("failed to close", "error", err.Error())
	}
}

// CloseOrPanic closes the object or, if it fails, panics
func CloseOrPanic(closer io.Closer) {
	if err := closer.Close(); err != nil {
		panic("close error: " + err.Error())
	}
}

// closerContext is an analog of io.Closer that also accepts a context
type closerContext interface {
	Close(context.Context) error
}

// CloseOrLogContext - the same as CloseOrLog but for closerContext
func CloseOrLogContext(ctx context.Context, log *slog.Logger, closer closerContext) {
	if err := closer.Close(ctx); err != nil {
		log.Error("failed to close", "error", err.Error())
	}
}

// CloseOrPanicContext - the same as CloseOrPanic but for closerContext
func CloseOrPanicContext(ctx context.Context, closer closerContext) {
	if err := closer.Close(ctx); err != nil {
		panic("close error: " + err.Error())
	}
}
