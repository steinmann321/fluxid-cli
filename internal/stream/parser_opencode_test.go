package stream

import (
	"bytes"
	"strings"
	"testing"
)

func TestOpencodeParserTextEvent(t *testing.T) {
	t.Parallel()
	input := `{"type":"text","part":{"type":"text","text":"Hello from Opencode"}}`

	var output bytes.Buffer
	parser := NewOpencodeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello from Opencode\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestOpencodeParserStepFinish(t *testing.T) {
	t.Parallel()
	input := `{"type":"step_finish","part":{"type":"step-finish","reason":"stop"}}`

	var output bytes.Buffer
	parser := NewOpencodeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should show success indicator
	if !strings.Contains(output.String(), "SUCCESS") {
		t.Errorf("Expected success indicator in output, got %q", output.String())
	}
}

func TestOpencodeParserWithoutPart(t *testing.T) {
	t.Parallel()
	input := `{"type":"text"}`

	var output bytes.Buffer
	parser := NewOpencodeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without crashing
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

func TestOpencodeParserInvalidJSON(t *testing.T) {
	t.Parallel()
	input := `invalid json here`

	var output bytes.Buffer
	parser := NewOpencodeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Invalid JSON should be passed through
	expected := "invalid json here\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}
