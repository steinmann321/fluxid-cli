package stream

import (
	"bytes"
	"strings"
	"testing"
)

func TestCodexParserAgentMessage(t *testing.T) {
	t.Parallel()
	input := `{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"Hello from Codex"}}`

	var output bytes.Buffer
	parser := NewCodexParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello from Codex\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestCodexParserReasoningItem(t *testing.T) {
	t.Parallel()
	input := `{"type":"item.completed","item":{"id":"item_0","type":"reasoning","text":"Thinking..."}}`

	var output bytes.Buffer
	parser := NewCodexParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Reasoning items should not produce output
	if output.String() != "" {
		t.Errorf("Expected empty output for reasoning item, got %q", output.String())
	}
}

func TestCodexParserTurnCompleted(t *testing.T) {
	t.Parallel()
	input := `{"type":"turn.completed","usage":{"input_tokens":100,"output_tokens":50}}`

	var output bytes.Buffer
	parser := NewCodexParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should show success indicator
	if !strings.Contains(output.String(), "SUCCESS") {
		t.Errorf("Expected success indicator in output, got %q", output.String())
	}
}

func TestCodexParserInvalidJSON(t *testing.T) {
	t.Parallel()
	input := `not valid json`

	var output bytes.Buffer
	parser := NewCodexParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Invalid JSON should be passed through
	expected := "not valid json\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}
