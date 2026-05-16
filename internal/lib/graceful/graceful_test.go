package graceful

import (
	"context"
	"testing"
	"time"
)

func TestRunWithGracefulShutdown_ShutdownCalledAfterExecution(t *testing.T) {
	execCalled := make(chan struct{}, 1)
	shutdownCalled := make(chan struct{}, 1)
	execCtxCancelled := make(chan bool, 1)

	execution := func(ctx context.Context) {
		execCalled <- struct{}{}

		// During execution the context should not yet be cancelled by the
		// deferred cancel in RunWithGracefulShutdown.
		select {
		case <-ctx.Done():
			execCtxCancelled <- true
		default:
			execCtxCancelled <- false
		}
	}

	shutdown := func() {
		shutdownCalled <- struct{}{}
	}

	// Call the function under test. It will call execution synchronously,
	// then return which triggers the deferred cancel and should cause the
	// registered AfterFunc to invoke shutdown.
	RunWithGracefulShutdown(execution, shutdown)

	// Verify execution was called.
	select {
	case <-execCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("execution was not called")
	}

	// Verify shutdown was called.
	select {
	case <-shutdownCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("shutdown was not called")
	}

	// Verify the context passed to execution was not cancelled during execution.
	if <-execCtxCancelled {
		t.Fatal("expected context NOT to be cancelled during execution")
	}
}

func TestRunWithGracefulShutdown_ShutdownCalledOnPanic(t *testing.T) {
	shutdownCalled := make(chan struct{}, 1)

	execution := func(ctx context.Context) {
		panic("simulated panic")
	}

	shutdown := func() {
		shutdownCalled <- struct{}{}
	}

	recover := func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from execution")
		}
	}

	// Run in a separate function so we can recover the panic and let the
	// deferred cancel in RunWithGracefulShutdown run.
	func() {
		defer recover()

		RunWithGracefulShutdown(execution, shutdown)
	}()

	select {
	case <-shutdownCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("shutdown was not called after panic")
	}
}

func TestRunWithGracefulShutdown_ConcurrentInvocations(t *testing.T) {
	const n = 2
	execCalled := make([]chan struct{}, n)
	shutdownCalled := make([]chan struct{}, n)

	for i := 0; i < n; i++ {
		execCalled[i] = make(chan struct{}, 1)
		shutdownCalled[i] = make(chan struct{}, 1)
		idx := i

		go func() {
			RunWithGracefulShutdown(func(ctx context.Context) {
				execCalled[idx] <- struct{}{}
			}, func() {
				shutdownCalled[idx] <- struct{}{}
			})
		}()
	}

	// Verify each invocation called execution and shutdown.
	for i := 0; i < n; i++ {
		select {
		case <-execCalled[i]:
		case <-time.After(1 * time.Second):
			t.Fatalf("execution %d was not called", i)
		}

		select {
		case <-shutdownCalled[i]:
		case <-time.After(1 * time.Second):
			t.Fatalf("shutdown %d was not called", i)
		}
	}
}

func TestRunWithGracefulShutdown_NilExecutionPanics(t *testing.T) {
	// If execution is nil, calling it should panic in the current
	// goroutine; recover here and verify shutdown still runs.
	shutdownCalled := make(chan struct{}, 1)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic when execution is nil")
			}
		}()

		RunWithGracefulShutdown(nil, func() {
			shutdownCalled <- struct{}{}
		})
	}()

	select {
	case <-shutdownCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("shutdown was not called after nil execution panic")
	}
}
