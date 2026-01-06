package errors_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	fluxiderr "fluxid-cli/internal/errors"
)

//nolint:paralleltest // Uses shared error output buffer
func TestNewConfigError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "simple error",
			description: "missing required field",
			want:        "error: config: missing required field",
		},
		{
			name:        "detailed error",
			description: "invalid agent value: must be one of [claude, codex, opencode]",
			want:        "error: config: invalid agent value: must be one of [claude, codex, opencode]",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fluxiderr.NewConfigError(testCase.description)
			if err.Error() != testCase.want {
				t.Errorf("fluxiderr.NewConfigError() = %q, want %q", err.Error(), testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestNewArgsError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "simple error",
			description: "invalid flag format",
			want:        "error: args: invalid flag format",
		},
		{
			name:        "detailed error",
			description: "--flag value syntax not supported, use --flag=value",
			want:        "error: args: --flag value syntax not supported, use --flag=value",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fluxiderr.NewArgsError(testCase.description)
			if err.Error() != testCase.want {
				t.Errorf("fluxiderr.NewArgsError() = %q, want %q", err.Error(), testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestNewWorkflowError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "simple error",
			description: "execution failed",
			want:        "error: workflow: execution failed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fluxiderr.NewWorkflowError(testCase.description)
			if err.Error() != testCase.want {
				t.Errorf("fluxiderr.NewWorkflowError() = %q, want %q", err.Error(), testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestNewIPCError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "simple error",
			description: "invalid report format",
			want:        "error: ipc: invalid report format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fluxiderr.NewIPCError(testCase.description)
			if err.Error() != testCase.want {
				t.Errorf("fluxiderr.NewIPCError() = %q, want %q", err.Error(), testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestComponentError(t *testing.T) {
	tests := []struct {
		name      string
		component string
		desc      string
		want      string
	}{
		{
			name:      "config component",
			component: "config",
			desc:      "missing required field",
			want:      "error: config: missing required field",
		},
		{
			name:      "args component",
			component: "args",
			desc:      "invalid flag",
			want:      "error: args: invalid flag",
		},
		{
			name:      "workflow component",
			component: "workflow",
			desc:      "execution failed",
			want:      "error: workflow: execution failed",
		},
		{
			name:      "ipc component",
			component: "ipc",
			desc:      "invalid format",
			want:      "error: ipc: invalid format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := &fluxiderr.ComponentError{
				Component:   testCase.component,
				Description: testCase.desc,
			}
			if err.Error() != testCase.want {
				t.Errorf("fluxiderr.ComponentError.Error() = %q, want %q", err.Error(), testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestLogError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ComponentError",
			err:  fluxiderr.NewConfigError("test error"),
			want: "error: config: test error\n",
		},
		{
			name: "standard error",
			err:  errors.New("standard error"), //nolint:err113 // Test code using simple error
			want: "standard error\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Capture stderr
			old := os.Stderr
			reader, writer, _ := os.Pipe()
			os.Stderr = writer

			fluxiderr.LogError(testCase.err)

			_ = writer.Close()
			os.Stderr = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)

			got := buf.String()
			if got != testCase.want {
				t.Errorf("LogError() output = %q, want %q", got, testCase.want)
			}
		})
	}
}

//nolint:paralleltest // Uses shared error output buffer
func TestWrappedError(t *testing.T) {
	//nolint:err113 // Test code using simple error for verification
	baseErr := errors.New("base error")
	configErr := fluxiderr.NewConfigError("wrapped: " + baseErr.Error())

	want := "error: config: wrapped: base error"
	if configErr.Error() != want {
		t.Errorf("wrapped error = %q, want %q", configErr.Error(), want)
	}
}
