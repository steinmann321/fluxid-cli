package stream

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestParseSystemEvent(t *testing.T) {
	t.Parallel()
	input := `{"type":"system","subtype":"init","model":"claude-sonnet-4.5"}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Model: claude-sonnet-4.5\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseAssistantTextMessage(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello world\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseToolCall(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant","tool_call":{"name":"Bash","description":"Run command"}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "[TOOL] Bash\n  Run command\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseToolResult(t *testing.T) {
	t.Parallel()
	input := `{"type":"tool_result","tool_call":{"name":"Bash"}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "[DONE] Bash\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseSuccessResult(t *testing.T) {
	t.Parallel()
	input := `{"type":"result","subtype":"success","duration_ms":1234,"num_turns":5,"total_cost_usd":0.023}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "\nSUCCESS: Duration 1234ms, Turns 5, Cost $0.023\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseErrorResult(t *testing.T) {
	t.Parallel()
	input := `{"type":"result","subtype":"error","error":"Something went wrong"}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "\nERROR: Something went wrong\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParseMalformedJSON(t *testing.T) {
	t.Parallel()
	input := `{invalid json}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	// Should not fail - non-JSON lines are passed through as plain text
	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse should not fail on malformed JSON: %v", err)
	}

	// Output should contain the line passed through as plain text
	expected := "{invalid json}\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestParsePlainText(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"Plain text line 1",
		"Plain text line 2",
		"PROMPT: Enter your name:",
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := output.String()

	// Verify all plain text lines are passed through
	if !strings.Contains(result, "Plain text line 1") {
		t.Error("Missing plain text line 1")
	}
	if !strings.Contains(result, "Plain text line 2") {
		t.Error("Missing plain text line 2")
	}
	if !strings.Contains(result, "PROMPT: Enter your name:") {
		t.Error("Missing prompt line")
	}
}

func TestParseMultipleEvents(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"system","subtype":"init","model":"claude-sonnet-4.5"}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Starting work"}]}}`,
		`{"type":"assistant","tool_call":{"name":"Bash","description":"Run tests"}}`,
		`{"type":"tool_result","tool_call":{"name":"Bash"}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Tests passed"}]}}`,
		`{"type":"result","subtype":"success","duration_ms":5000,"num_turns":3,"total_cost_usd":0.05}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := output.String()

	// Verify key elements are present
	if !strings.Contains(result, "Model: claude-sonnet-4.5") {
		t.Error("Missing model name")
	}
	if !strings.Contains(result, "Starting work") {
		t.Error("Missing first text message")
	}
	if !strings.Contains(result, "[TOOL] Bash") {
		t.Error("Missing tool start")
	}
	if !strings.Contains(result, "[DONE] Bash") {
		t.Error("Missing tool end")
	}
	if !strings.Contains(result, "Tests passed") {
		t.Error("Missing second text message")
	}
	if !strings.Contains(result, "SUCCESS") {
		t.Error("Missing success message")
	}
}

func TestParseMixedJSONAndPlainText(t *testing.T) {
	t.Parallel()
	// Simulate a mix of JSON (from real Claude) and plain text (from stub or other sources)
	input := strings.Join([]string{
		`{"type":"system","subtype":"init","model":"claude-sonnet-4.5"}`,
		"Plain text from stub",
		`{"type":"assistant","message":{"content":[{"type":"text","text":"JSON message"}]}}`,
		"PROMPT: Enter input:",
		`{"type":"result","subtype":"success","duration_ms":1000,"num_turns":1,"total_cost_usd":0.01}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result := output.String()

	// Verify JSON events are parsed
	if !strings.Contains(result, "Model: claude-sonnet-4.5") {
		t.Error("Missing model name from JSON")
	}
	if !strings.Contains(result, "JSON message") {
		t.Error("Missing text from JSON message")
	}
	if !strings.Contains(result, "SUCCESS") {
		t.Error("Missing success from JSON result")
	}

	// Verify plain text is passed through
	if !strings.Contains(result, "Plain text from stub") {
		t.Error("Missing plain text line")
	}
	if !strings.Contains(result, "PROMPT: Enter input:") {
		t.Error("Missing prompt line")
	}
}

func TestParseVeryLargeJSONLine(t *testing.T) {
	t.Parallel()
	// Create a JSON message with very large text content (> 1MB)
	largeText := strings.Repeat("A", 2*1024*1024) // 2MB of text
	input := fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":"%s"}]}}`, largeText)

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed on large line: %v", err)
	}

	result := output.String()

	// Verify the large text was processed (check beginning and end)
	if !strings.Contains(result, "AAAAAAA") {
		t.Error("Large text not in output")
	}

	// Verify we got the full text
	if !strings.HasSuffix(strings.TrimSpace(result), "AAAAAAA") {
		t.Error("Large text appears truncated")
	}
}

func TestParseMultipleLargeLines(t *testing.T) {
	t.Parallel()
	// Create multiple large JSON lines to test buffer reuse
	lines := make([]string, 5)
	for i := 0; i < 5; i++ {
		largeText := strings.Repeat(fmt.Sprintf("Line%d", i), 500*1024) // ~3MB per line
		lines[i] = fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":"%s"}]}}`, largeText)
	}
	input := strings.Join(lines, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed on multiple large lines: %v", err)
	}

	result := output.String()

	// Verify all lines were processed
	for i := 0; i < 5; i++ {
		marker := fmt.Sprintf("Line%d", i)
		if !strings.Contains(result, marker) {
			t.Errorf("Missing output from line %d", i)
		}
	}
}
