package stream

import (
	"fmt"
	"io"
)

// Formatter handles formatted output of parsed stream events.
type Formatter struct {
	writer io.Writer
}

// NewFormatter creates a new formatter that writes to the given writer.
func NewFormatter(writer io.Writer) *Formatter {
	return &Formatter{writer: writer}
}

// WriteModel outputs the model name at the start of a session.
func (f *Formatter) WriteModel(model string) {
	_, _ = fmt.Fprintf(f.writer, "Model: %s\n", model)
}

// WriteTextDelta outputs a text delta without a newline (for streaming).
func (f *Formatter) WriteTextDelta(text string) {
	_, _ = fmt.Fprint(f.writer, text)
}

// WriteTextBlockEnd outputs a newline to end a text block.
func (f *Formatter) WriteTextBlockEnd() {
	_, _ = fmt.Fprintln(f.writer)
}

// WriteToolStart outputs the start of a tool execution.
func (f *Formatter) WriteToolStart(name, description string) {
	_, _ = fmt.Fprintf(f.writer, "[TOOL] %s\n", name)
	if description != "" {
		_, _ = fmt.Fprintf(f.writer, "  %s\n", description)
	}
}

// WriteToolEnd outputs the completion of a tool execution.
func (f *Formatter) WriteToolEnd(name string) {
	_, _ = fmt.Fprintf(f.writer, "[DONE] %s\n", name)
}

// WriteSuccess outputs the final success statistics.
func (f *Formatter) WriteSuccess(duration, turns int, cost float64) {
	_, _ = fmt.Fprintln(f.writer)
	_, _ = fmt.Fprintf(f.writer, "SUCCESS: Duration %dms, Turns %d, Cost $%.3f\n", duration, turns, cost)
}

// WriteError outputs the final error message.
func (f *Formatter) WriteError(errorMsg string) {
	_, _ = fmt.Fprintln(f.writer)
	_, _ = fmt.Fprintf(f.writer, "ERROR: %s\n", errorMsg)
}
