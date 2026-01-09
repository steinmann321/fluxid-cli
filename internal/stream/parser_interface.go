package stream

import (
	"io"
)

// AgentParser is the interface that all agent-specific parsers implement.
type AgentParser interface {
	// Parse reads and parses the JSON stream line by line.
	Parse() error
}

// NewParser creates the appropriate agent-specific parser.
//
//nolint:ireturn // Factory pattern intentionally returns interface.
func NewParser(reader io.Reader, writer io.Writer, agent string) AgentParser {
	const bufferSize = 256 * 1024 // 256KB read buffer

	switch agent {
	case "codex":
		return NewCodexParser(reader, writer, bufferSize)
	case "opencode":
		return NewOpencodeParser(reader, writer, bufferSize)
	case "gemini":
		return NewGeminiParser(reader, writer, bufferSize)
	default:
		// Default to Claude
		return NewClaudeParser(reader, writer, bufferSize)
	}
}
