package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
)

// GeminiParser parses Gemini's stream-json output format.
type GeminiParser struct {
	reader    *bufio.Reader
	formatter *Formatter
}

// NewGeminiParser creates a new Gemini parser.
func NewGeminiParser(reader io.Reader, writer io.Writer, bufferSize int) *GeminiParser {
	return &GeminiParser{
		reader:    bufio.NewReaderSize(reader, bufferSize),
		formatter: NewFormatter(writer),
	}
}

// Parse reads and parses the Gemini JSON stream line by line.
func (p *GeminiParser) Parse() error {
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

// parseLine parses a single JSON line from Gemini using gjson.
func (p *GeminiParser) parseLine(line string) {
	if !gjson.Valid(line) {
		// Not valid JSON - pass through as plain text
		p.formatter.WriteTextDelta(line)
		p.formatter.WriteTextBlockEnd()
		return
	}

	result := gjson.Parse(line)
	eventType := result.Get("type").String()

	switch eventType {
	case "message":
		p.handleMessage(result)
	case "result":
		p.handleResult(result)
	}
}

// handleMessage processes message events from Gemini.
func (p *GeminiParser) handleMessage(result gjson.Result) {
	role := result.Get("role").String()

	// Only process assistant messages
	if role == "assistant" {
		content := result.Get("content").String()
		if content != "" {
			// Gemini sends delta chunks
			p.formatter.WriteTextDelta(content)
		}
	}
}

// handleResult processes result events from Gemini.
func (p *GeminiParser) handleResult(result gjson.Result) {
	// End the text block first
	p.formatter.WriteTextBlockEnd()

	status := result.Get("status").String()
	if status == "success" {
		// Gemini doesn't provide duration or cost in result
		p.formatter.WriteSuccess(0, 1, 0)
	}
}
