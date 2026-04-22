package graceful

import (
	"context"
	"os/signal"
	"syscall"
)

// RunWithGracefulShutdown executes the main runtime function and registers
// shutdown hooks that run when a termination signal is received.
func RunWithGracefulShutdown(execution func(executionContext context.Context), shutdown func()) {
	// Build a cancellable context that is closed on OS termination signals.
	sysContext, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// Register graceful shutdown callbacks.
	context.AfterFunc(sysContext, func() {
		shutdown()
	})

	// Start the main execution flow.
	execution(sysContext)
}
