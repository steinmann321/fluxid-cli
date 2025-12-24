package main

// exitCode stores the exit code passed to os.Exit during tests.
var exitCode int //nolint:gochecknoglobals // Test helper global

// mockExit is a test double for os.Exit.
func mockExit(code int) {
	exitCode = code
}
