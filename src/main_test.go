package main

import (
	"bytes"
	"fluxid-loop/src/app"
	"os"
	"testing"
)

// TestMainPrintsHello captures stdout and ensures main prints the greeting.
func TestMainPrintsHello(t *testing.T) {
	t.Parallel()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	// Configure logger to write to stdout and run the program logic
	app.SetupLogger(os.Stdout)
	main()

	// Restore stdout and collect output
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}

	if got := buf.String(); got != "hello world\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}
