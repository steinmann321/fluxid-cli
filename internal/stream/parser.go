package stream

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ParserState tracks the current state of the stream parser.
type ParserState struct {
	inTextBlock         bool
	inToolBlock         bool
	textBlockHasContent bool
	currentToolName     string
}

// Parser parses Claude's stream-json output and formats it for display.
type Parser struct {
	reader    *bufio.Reader
	formatter *Formatter
	state     ParserState
}

// NewStreamParser creates a new stream parser.
//
// IMPORTANT: This parser can handle arbitrarily large JSON lines without crashing.
// Unlike bufio.Scanner (which has fixed buffer limits), this implementation uses
// bufio.Reader with a growing bytes.Buffer to accumulate lines of any size.
//
// This prevents "token too long" errors and process crashes when Claude returns
// large responses (e.g., reading large files, extensive code generation).
func NewStreamParser(reader io.Reader, writer io.Writer) *Parser {
	// Use bufio.Reader with large buffer for efficient reading
	// This is much more flexible than Scanner for handling arbitrarily large lines
	const bufferSize = 256 * 1024 // 256KB read buffer
	bufferedReader := bufio.NewReaderSize(reader, bufferSize)

	return &Parser{
		reader:    bufferedReader,
		formatter: NewFormatter(writer),
		state: ParserState{
			inTextBlock:         false,
			inToolBlock:         false,
			textBlockHasContent: false,
			currentToolName:     "",
		},
	}
}

// Parse reads and parses the JSON stream line by line.
// Uses a growing buffer approach to handle arbitrarily large lines without crashes.
func (p *Parser) Parse() error {
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

// parseLine parses a single JSON line and dispatches to the appropriate handler.
// If the line is not valid JSON, it passes it through as plain text (for test stubs).
func (p *Parser) parseLine(line string) {
	var event Event
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// Not valid JSON - pass through as plain text (for test stubs and non-JSON output)
		p.formatter.WriteTextDelta(line)
		p.formatter.WriteTextBlockEnd()
		return
	}

	switch event.Type {
	case "system":
		p.handleSystemEvent(&event)
	case "stream_event":
		p.handleEvent(&event)
	case "assistant":
		p.handleAssistantEvent(&event)
	case "tool_result":
		p.handleToolResultEvent(&event)
	case "result":
		p.handleResultEvent(&event)
	}
}

// handleSystemEvent processes system initialization events.
func (p *Parser) handleSystemEvent(event *Event) {
	if event.Subtype == "init" && event.Model != "" {
		p.formatter.WriteModel(event.Model)
	}
}

// handleEvent processes streaming API events (content blocks, deltas, etc).
func (p *Parser) handleEvent(event *Event) {
	if event.Event == nil {
		return
	}

	switch event.Event.Type {
	case "message_start":
		p.state.textBlockHasContent = false
	case "content_block_start":
		p.handleContentBlockStart(event)
	case "content_block_delta":
		p.handleContentBlockDelta(event)
	case "content_block_stop":
		p.handleContentBlockStop()
	}
}

func (p *Parser) handleContentBlockStart(event *Event) {
	if event.Event.ContentBlock == nil {
		return
	}
	switch event.Event.ContentBlock.Type {
	case "text":
		p.state.inTextBlock = true
		p.state.textBlockHasContent = false
	case "tool_use":
		p.state.inToolBlock = true
		p.state.currentToolName = event.Event.ContentBlock.Name
	}
}

func (p *Parser) handleContentBlockDelta(event *Event) {
	if event.Event.Delta == nil {
		return
	}
	if event.Event.Delta.Type == "text_delta" && event.Event.Delta.Text != "" {
		p.formatter.WriteTextDelta(event.Event.Delta.Text)
		p.state.textBlockHasContent = true
	}
}

func (p *Parser) handleContentBlockStop() {
	if p.state.inTextBlock {
		if p.state.textBlockHasContent {
			p.formatter.WriteTextBlockEnd()
		}
		p.state.inTextBlock = false
		p.state.textBlockHasContent = false
	} else if p.state.inToolBlock {
		p.state.inToolBlock = false
	}
}

// handleAssistantEvent processes assistant messages (tool call announcements and text content).
func (p *Parser) handleAssistantEvent(event *Event) {
	// Handle tool call announcements
	if event.ToolCall != nil && event.ToolCall.Name != "" {
		p.formatter.WriteToolStart(event.ToolCall.Name, event.ToolCall.Description)
		return
	}

	// Handle text content in assistant messages (non-streaming case)
	if event.Message != nil && len(event.Message.Content) > 0 {
		for _, item := range event.Message.Content {
			if item.Type == "text" && item.Text != "" {
				p.formatter.WriteTextDelta(item.Text)
				p.formatter.WriteTextBlockEnd()
			}
		}
	}
}

// handleToolResultEvent processes tool completion events.
func (p *Parser) handleToolResultEvent(event *Event) {
	if event.ToolCall != nil && event.ToolCall.Name != "" {
		p.formatter.WriteToolEnd(event.ToolCall.Name)
	}
}

// handleResultEvent processes final result events (success/error).
func (p *Parser) handleResultEvent(event *Event) {
	if event.Subtype == "success" {
		p.formatter.WriteSuccess(event.DurationMS, event.NumTurns, event.TotalCostUSD)
	} else {
		errorMsg := event.Error
		if errorMsg == "" {
			errorMsg = "Unknown error"
		}
		p.formatter.WriteError(errorMsg)
	}
}
