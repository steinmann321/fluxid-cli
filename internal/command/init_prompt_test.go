package command

import (
	"os"
	"sync"
	"testing"
)

// Tests for promptOverwrite function
// Note: These tests cannot run in parallel as they modify global state (os.Stdin)

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ConfirmWithY(t *testing.T) {
	// Mock stdin with "y"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("y\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if !result {
		t.Error("Expected true for 'y' input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ConfirmWithYes(t *testing.T) {
	// Mock stdin with "yes"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("yes\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if !result {
		t.Error("Expected true for 'yes' input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ConfirmWithUppercase(t *testing.T) {
	// Mock stdin with "YES"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("YES\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if !result {
		t.Error("Expected true for 'YES' input (case-insensitive)")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ConfirmWithMixedCase(t *testing.T) {
	// Mock stdin with "YeS"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("YeS\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if !result {
		t.Error("Expected true for 'YeS' input (case-insensitive)")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_DeclineWithN(t *testing.T) {
	// Mock stdin with "n"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("n\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if result {
		t.Error("Expected false for 'n' input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_DeclineWithNo(t *testing.T) {
	// Mock stdin with "no"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("no\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if result {
		t.Error("Expected false for 'no' input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_DeclineWithInvalidInput(t *testing.T) {
	// Mock stdin with invalid input
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("maybe\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if result {
		t.Error("Expected false for invalid input 'maybe'")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_DeclineWithEmptyInput(t *testing.T) {
	// Mock stdin with empty input
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("\n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if result {
		t.Error("Expected false for empty input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_DeclineWithWhitespace(t *testing.T) {
	// Mock stdin with whitespace
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("   \n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if result {
		t.Error("Expected false for whitespace input")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_EOFError(t *testing.T) {
	// Mock stdin with EOF (immediate close)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	_ = w.Close() // Close immediately to trigger EOF

	result := promptOverwrite("/tmp/test")
	if result {
		t.Error("Expected false for EOF error")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ReadError(t *testing.T) {
	// Mock stdin with a broken pipe to trigger read error
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	// Use a pipe and close both ends to simulate error
	r, w, _ := os.Pipe()
	_ = w.Close()
	_ = r.Close()
	os.Stdin = r

	result := promptOverwrite("/tmp/test")
	if result {
		t.Error("Expected false for read error")
	}
}

//nolint:paralleltest // Tests modify global os.Stdin and cannot run concurrently
func TestPromptOverwrite_ConfirmWithExtraWhitespace(t *testing.T) {
	// Mock stdin with "  yes  " (extra whitespace should be trimmed)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe() //nolint:varnamelen // Standard os.Pipe() return names
	os.Stdin = r

	var wg sync.WaitGroup //nolint:varnamelen // Standard abbreviation for WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.WriteString("  yes  \n")
		_ = w.Close()
	}()

	result := promptOverwrite("/tmp/test")
	wg.Wait()

	if !result {
		t.Error("Expected true for '  yes  ' input (whitespace trimmed)")
	}
}
