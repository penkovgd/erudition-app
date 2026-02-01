package closer

import (
	"context"
	"io"
	"log/slog"
)

func CloseOrLog(log *slog.Logger, closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Error("failed to close", "error", err.Error())
	}
}
func CloseOrPanic(closer io.Closer) {
	if err := closer.Close(); err != nil {
		panic("close error: " + err.Error())
	}
}

type CloserContext interface {
	Close(context.Context) error
}

func CloseOrLogContext(ctx context.Context, log *slog.Logger, closer CloserContext) {
	if err := closer.Close(ctx); err != nil {
		log.Error("failed to close", "error", err.Error())
	}
}
func CloseOrPanicContext(ctx context.Context, closer CloserContext) {
	if err := closer.Close(ctx); err != nil {
		panic("close error: " + err.Error())
	}
}
