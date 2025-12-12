//nolint:paralleltest // CLI tests manipulate os.Args
package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHandleIPCCommand_NoSubcommand(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := handleIPCCommand([]string{})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "'ipc' requires a subcommand") {
		t.Errorf("Expected error message about missing subcommand, got: %s", output)
	}
}

func TestHandleIPCCommand_UnknownSubcommand(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := handleIPCCommand([]string{"unknown-command"})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "unknown ipc subcommand") {
		t.Errorf("Expected error message about unknown subcommand, got: %s", output)
	}
}

func TestHandleIPCCommand_GetReportSchema(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleIPCCommand([]string{"get-report-schema"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains expected schema fields
	expectedFields := []string{"command:", "artifact:", "timestamp:", "status:", "issues:"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing expected field %s", field)
		}
	}
}

func TestHandleGetReportSchema_Help(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleGetReportSchema([]string{"--help"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify help output contains usage information
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Help output missing Usage section")
	}

	if !strings.Contains(output, "Description:") {
		t.Errorf("Help output missing Description section")
	}

	if !strings.Contains(output, "Example:") {
		t.Errorf("Help output missing Example section")
	}
}

func TestHandleGetReportSchema_ShortHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleGetReportSchema([]string{"-h"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify help output
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Help output missing Usage section")
	}
}

func TestPrintHelp(t *testing.T) {
	var buf bytes.Buffer
	testText := "Test help message"
	printHelp(&buf, testText)

	output := buf.String()
	if output != testText {
		t.Errorf("printHelp() output = %q, want %q", output, testText)
	}
}

func TestParseSessionFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		envValue    string
		wantSession string
		wantErr     bool
	}{
		{
			name:        "session from --session flag",
			args:        []string{"--session", "test-123"},
			envValue:    "",
			wantSession: "test-123",
			wantErr:     false,
		},
		{
			name:        "session from environment variable",
			args:        []string{},
			envValue:    "env-session-456",
			wantSession: "env-session-456",
			wantErr:     false,
		},
		{
			name:        "missing session flag value",
			args:        []string{"--session"},
			envValue:    "",
			wantSession: "",
			wantErr:     true,
		},
		{
			name:        "no session provided",
			args:        []string{},
			envValue:    "",
			wantSession: "",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment
			if tc.envValue != "" {
				t.Setenv("FLUXID_SESSION_ID", tc.envValue)
			} else {
				_ = os.Unsetenv("FLUXID_SESSION_ID")
			}

			// Parse session flag
			sessionID, err := parseSessionFlag(tc.args)

			// Check error
			if (err != nil) != tc.wantErr {
				t.Errorf("parseSessionFlag() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Check session ID
			if sessionID != tc.wantSession {
				t.Errorf("parseSessionFlag() sessionID = %q, want %q", sessionID, tc.wantSession)
			}
		})
	}
}
