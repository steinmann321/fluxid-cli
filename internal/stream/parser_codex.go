package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
)

// CodexParser parses Codex's JSONL output format.
type CodexParser struct {
	reader    *bufio.Reader
	formatter *Formatter
}

// NewCodexParser creates a new Codex parser.
func NewCodexParser(reader io.Reader, writer io.Writer, bufferSize int) *CodexParser {
	return &CodexParser{
		reader:    bufio.NewReaderSize(reader, bufferSize),
		formatter: NewFormatter(writer),
	}
}

// Parse reads and parses the Codex JSON stream line by line.
func (p *CodexParser) Parse() error {
	var lineBuffer bytes.Buffer

	for {
		// Read until newline, accumulating in buffer
		chunk, err := p.reader.ReadBytes('\n')

		if len(chunk) > 0 {
			// Append chunk to line buffer
			lineBuffer.Write(chunk)

			// If we got a newline, process the complete line
			if chunk[len(chunk)-1] == '\n' {
				line := lineBuffer.String()
				// Remove trailing newline
				line = line[:len(line)-1]

				// Process the line
				p.parseLine(line)

				// Reset buffer for next line
				lineBuffer.Reset()
			}
		}

		// Check for end of stream
		if err == io.EOF {
			// Process any remaining data in buffer
			if lineBuffer.Len() > 0 {
				p.parseLine(lineBuffer.String())
			}
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to read stream: %w", err)
		}
	}
}

// parseLine parses a single JSON line from Codex using gjson.
func (p *CodexParser) parseLine(line string) {
	if !gjson.Valid(line) {
		// Not valid JSON - pass through as plain text
		p.formatter.WriteTextDelta(line)
		p.formatter.WriteTextBlockEnd()
		return
	}

	result := gjson.Parse(line)
	eventType := result.Get("type").String()

	switch eventType {
	case "item.completed":
		p.handleItemCompleted(result)
	case "turn.completed":
		p.handleTurnCompleted(result)
	}
}

// handleItemCompleted processes item.completed events from Codex.
func (p *CodexParser) handleItemCompleted(result gjson.Result) {
	itemType := result.Get("item.type").String()

	if itemType == "agent_message" {
		text := result.Get("item.text").String()
		if text != "" {
			p.formatter.WriteTextDelta(text)
			p.formatter.WriteTextBlockEnd()
		}
	}
	// Ignore other item types like "reasoning"
}

// handleTurnCompleted processes turn.completed events from Codex.
func (p *CodexParser) handleTurnCompleted(result gjson.Result) {
	// Check if usage information exists
	if result.Get("usage.input_tokens").Exists() && result.Get("usage.output_tokens").Exists() {
		// Show completion (Codex doesn't provide cost or duration)
		p.formatter.WriteSuccess(0, 1, 0)
	}
}
