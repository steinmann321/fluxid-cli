// Package main implements the fluxid CLI entry point.
package main

import (
	"fluxid-cli/internal/command"
	"os"
)

// osExit is a variable holding os.Exit to allow test stubbing.
var osExit = os.Exit //nolint:gochecknoglobals // Required for test stubbing os.Exit

func main() {
	exitCode := command.Execute()
	osExit(exitCode)
}
