package workflow

import (
	"os"
	"testing"
)

// TestMain sets up the test environment before running any tests.
func TestMain(m *testing.M) {
	// Run all tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
