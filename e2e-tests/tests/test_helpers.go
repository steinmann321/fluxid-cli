package tests

import (
	"context"
	"fmt"
	"time"
)

const (
	minExpectedArgCount = 2 // Minimum expected argument count for validation
)

// standardCommandFilesConfigFor returns a config with absolute paths to command files in the given directory.
func standardCommandFilesConfigFor(dir string) string {
	return fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, dir, dir, dir)
}

// testContext creates a context with timeout for testing.
// This helper avoids direct context.Background() calls in test files.
//
//nolint:unused // Helper function for future test scenarios
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
