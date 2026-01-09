package stream

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewParserClaude tests that NewParser returns ClaudeParser for "claude" agent.
func TestNewParserClaude(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "claude")

	if _, ok := parser.(*ClaudeParser); !ok {
		t.Errorf("Expected *ClaudeParser, got %T", parser)
	}
}

// TestNewParserCodex tests that NewParser returns CodexParser for "codex" agent.
func TestNewParserCodex(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "codex")

	if _, ok := parser.(*CodexParser); !ok {
		t.Errorf("Expected *CodexParser, got %T", parser)
	}
}

// TestNewParserOpencode tests that NewParser returns OpencodeParser for "opencode" agent.
func TestNewParserOpencode(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "opencode")

	if _, ok := parser.(*OpencodeParser); !ok {
		t.Errorf("Expected *OpencodeParser, got %T", parser)
	}
}

// TestNewParserGemini tests that NewParser returns GeminiParser for "gemini" agent.
func TestNewParserGemini(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "gemini")

	if _, ok := parser.(*GeminiParser); !ok {
		t.Errorf("Expected *GeminiParser, got %T", parser)
	}
}

// TestNewParserDefault tests that NewParser returns ClaudeParser for unknown agent.
func TestNewParserDefault(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "unknown-agent")

	if _, ok := parser.(*ClaudeParser); !ok {
		t.Errorf("Expected *ClaudeParser (default), got %T", parser)
	}
}

// TestNewParserEmptyAgent tests that NewParser returns ClaudeParser for empty agent string.
func TestNewParserEmptyAgent(t *testing.T) {
	t.Parallel()
	input := `{}`
	var output bytes.Buffer
	parser := NewParser(strings.NewReader(input), &output, "")

	if _, ok := parser.(*ClaudeParser); !ok {
		t.Errorf("Expected *ClaudeParser (default), got %T", parser)
	}
}
