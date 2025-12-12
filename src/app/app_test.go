package app

import (
	"bytes"
	"testing"
)

func TestHelloReturnsGreeting(t *testing.T) {
	t.Parallel()
	if got := Hello(); got != "hello world" {
		t.Fatalf("Hello() = %q, want %q", got, "hello world")
	}
}

func TestPrintHelloWritesGreeting(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	SetupLogger(&buf)
	PrintHello()
	if got := buf.String(); got != "hello world\n" {
		t.Fatalf("PrintHello() wrote %q, want %q", got, "hello world\n")
	}
}
