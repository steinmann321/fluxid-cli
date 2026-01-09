package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
)

// OpencodeParser parses Opencode's JSONL output format.
type OpencodeParser struct {
	reader    *bufio.Reader
	formatter *Formatter
}

// NewOpencodeParser creates a new Opencode parser.
func NewOpencodeParser(reader io.Reader, writer io.Writer, bufferSize int) *OpencodeParser {
	return &OpencodeParser{
		reader:    bufio.NewReaderSize(reader, bufferSize),
		formatter: NewFormatter(writer),
	}
}

// Parse reads and parses the Opencode JSON stream line by line.
func (p *OpencodeParser) Parse() error {
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

// parseLine parses a single JSON line from Opencode using gjson.
func (p *OpencodeParser) parseLine(line string) {
	if !gjson.Valid(line) {
		// Not valid JSON - pass through as plain text
		p.formatter.WriteTextDelta(line)
		p.formatter.WriteTextBlockEnd()
		return
	}

	result := gjson.Parse(line)
	eventType := result.Get("type").String()

	switch eventType {
	case "text":
		p.handleText(result)
	case "step_finish":
		p.handleStepFinish(result)
	}
}

// handleText processes text events from Opencode.
func (p *OpencodeParser) handleText(result gjson.Result) {
	text := result.Get("part.text").String()
	if text != "" {
		p.formatter.WriteTextDelta(text)
		p.formatter.WriteTextBlockEnd()
	}
}

// handleStepFinish processes step_finish events from Opencode.
func (p *OpencodeParser) handleStepFinish(_ gjson.Result) {
	// Step completed - show completion info
	// Opencode doesn't provide duration or cost in the finish event
	p.formatter.WriteSuccess(0, 1, 0)
}
