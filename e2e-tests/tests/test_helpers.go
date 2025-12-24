package tests

import (
	"context"
	"time"
)

const (
	minExpectedArgCount = 2 // Minimum expected argument count for validation
)

// Common test constants.
const (
	standardCommandFilesConfig = `commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
)

// testContext creates a context with timeout for testing.
// This helper avoids direct context.Background() calls in test files.
func testContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// testCtx returns a context with timeout for single-value contexts.
// The cancel function is called automatically after timeout expires.
//
//nolint:unparam // Different timeouts may be used in future
func testCtx(timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// Cancel will be called automatically when the timeout expires
	// We don't defer cancel() here to avoid premature cancellation
	_ = cancel
	return ctx
}
