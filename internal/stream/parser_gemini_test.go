package stream

import (
	"bytes"
	"strings"
	"testing"
)

func TestGeminiParserAssistantMessage(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"Hello","delta":true}`,
		`{"type":"message","role":"assistant","content":" from Gemini","delta":true}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewGeminiParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello from Gemini"
	if !strings.Contains(output.String(), expected) {
		t.Errorf("Expected output to contain %q, got %q", expected, output.String())
	}
}

func TestGeminiParserUserMessage(t *testing.T) {
	t.Parallel()
	input := `{"type":"message","role":"user","content":"Say hello"}`

	var output bytes.Buffer
	parser := NewGeminiParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// User messages should not produce output
	if output.String() != "" {
		t.Errorf("Expected empty output for user message, got %q", output.String())
	}
}

func TestGeminiParserResult(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"Done","delta":true}`,
		`{"type":"result","status":"success"}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewGeminiParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should show success indicator
	if !strings.Contains(output.String(), "SUCCESS") {
		t.Errorf("Expected success indicator in output, got %q", output.String())
	}
}

func TestGeminiParserInvalidJSON(t *testing.T) {
	t.Parallel()
	input := `not valid json at all`

	var output bytes.Buffer
	parser := NewGeminiParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Invalid JSON should be passed through
	expected := "not valid json at all\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}
