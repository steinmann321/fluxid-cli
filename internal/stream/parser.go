// Package stream provides JSON stream parsing for agent outputs.
package stream

const (
	typeText = "text"
)

// ParserState tracks the current state of the stream parser.
// Used by Claude parser which needs to track content blocks.
type ParserState struct {
	inTextBlock         bool
	inToolBlock         bool
	textBlockHasContent bool
	currentToolName     string
}
