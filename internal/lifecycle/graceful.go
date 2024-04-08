package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func GetGracefulContext(ctx context.Context, onShutdown func(context.Context, error)) (context.Context, context.CancelFunc) {
	// Graceful Shutdown
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sCtx, stop := context.WithCancelCause(ctx)
	go func(gCtx context.Context, stop context.CancelCauseFunc) {
		select {
		case <-gCtx.Done():
			cause := context.Cause(gCtx)
			onShutdown(ctx, cause)
			stop(cause)
		}
	}(gCtx, stop)

	return sCtx, cancel
}
