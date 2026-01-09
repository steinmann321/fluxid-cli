package stream

import (
	"bytes"
	"strings"
	"testing"
)

// TestParseStreamingTextBlocks tests streaming text content (content_block events).
func TestParseStreamingTextBlocks(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":0,` +
			`"delta":{"type":"text_delta","text":"Hello "}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":0,` +
			`"delta":{"type":"text_delta","text":"world"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":0}}`,
	}, "\n")

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

// TestParseStreamingToolBlock tests tool_use content blocks.
func TestParseStreamingToolBlock(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":0,` +
			`"content_block":{"type":"tool_use","name":"Read"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":0}}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Tool blocks don't produce output during streaming (only at start/end announcements)
	if output.String() != "" {
		t.Errorf("Expected empty output for tool block, got %q", output.String())
	}
}

// TestParseEmptyTextBlock tests a text block with no content.
func TestParseEmptyTextBlock(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":0}}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Empty text blocks should not produce newline
	if output.String() != "" {
		t.Errorf("Expected empty output for empty text block, got %q", output.String())
	}
}

// TestParseEventWithoutEventField tests handling events with nil event field.
func TestParseEventWithoutEventField(t *testing.T) {
	t.Parallel()
	input := `{"type":"stream_event"}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without crashing
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseContentBlockStartWithoutContentBlock tests nil content_block field.
func TestParseContentBlockStartWithoutContentBlock(t *testing.T) {
	t.Parallel()
	input := `{"type":"stream_event","event":{"type":"content_block_start","index":0}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without crashing
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseContentBlockDeltaWithoutDelta tests nil delta field.
func TestParseContentBlockDeltaWithoutDelta(t *testing.T) {
	t.Parallel()
	input := `{"type":"stream_event","event":{"type":"content_block_delta","index":0}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without crashing
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseErrorResultWithEmptyError tests error result with no error message.
func TestParseErrorResultWithEmptyError(t *testing.T) {
	t.Parallel()
	input := `{"type":"result","subtype":"error","error":""}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "\nERROR: Unknown error\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

// TestParseSystemEventWithoutModel tests system event with no model field.
func TestParseSystemEventWithoutModel(t *testing.T) {
	t.Parallel()
	input := `{"type":"system","subtype":"init","model":""}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should not output anything when model is empty
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseToolCallWithoutDescription tests tool call with empty description.
func TestParseToolCallWithoutDescription(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant","tool_call":{"name":"Bash","description":""}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "[TOOL] Bash\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

// TestParseAssistantEventWithoutToolCall tests assistant event with nil tool_call.
func TestParseAssistantEventWithoutToolCall(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant"}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without output
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseToolResultWithoutToolCall tests tool_result event with nil tool_call.
func TestParseToolResultWithoutToolCall(t *testing.T) {
	t.Parallel()
	input := `{"type":"tool_result"}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without output
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseMultipleTextBlocks tests multiple text blocks in sequence.
func TestParseMultipleTextBlocks(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":0,` +
			`"delta":{"type":"text_delta","text":"First"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":0}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":1,"content_block":{"type":"text"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":1,` +
			`"delta":{"type":"text_delta","text":"Second"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":1}}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "First\nSecond\n"
	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

// TestParseAssistantMessageWithEmptyContent tests assistant message with empty content array.
func TestParseAssistantMessageWithEmptyContent(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant","message":{"content":[]}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should handle gracefully without output
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseAssistantMessageWithNonTextContent tests assistant message with non-text content.
func TestParseAssistantMessageWithNonTextContent(t *testing.T) {
	t.Parallel()
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash"}]}}`

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Non-text content should be skipped
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}

// TestParseEmptyTextDelta tests text delta with empty text.
func TestParseEmptyTextDelta(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":""}}}`,
		`{"type":"stream_event","event":{"type":"content_block_stop","index":0}}`,
	}, "\n")

	var output bytes.Buffer
	parser := NewClaudeParser(strings.NewReader(input), &output, 256*1024)

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Empty delta should not count as content
	if output.String() != "" {
		t.Errorf("Expected empty output, got %q", output.String())
	}
}
