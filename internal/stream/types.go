// Package stream provides JSON stream parsing for Claude's stream-json output format.
package stream

// Event represents the top-level JSON event from Claude's stream-json output.
type Event struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`

	// System event fields
	Model string `json:"model,omitempty"`

	// Stream event fields
	Event *APIEvent `json:"event,omitempty"`

	// Assistant event fields
	Message *AssistantMessage `json:"message,omitempty"`

	// Tool call fields
	ToolCall *ToolCallInfo `json:"tool_call,omitempty"`

	// Result event fields
	DurationMS   int     `json:"duration_ms,omitempty"`
	NumTurns     int     `json:"num_turns,omitempty"`
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// APIEvent represents streaming API events (content blocks, deltas, etc).
type APIEvent struct {
	Type         string        `json:"type"`
	Index        int           `json:"index,omitempty"`
	ContentBlock *ContentBlock `json:"content_block,omitempty"`
	Delta        *Delta        `json:"delta,omitempty"`
}

// ContentBlock represents a content block in the stream.
type ContentBlock struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Delta represents incremental content updates.
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AssistantMessage represents a message from the assistant.
type AssistantMessage struct {
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
	Usage   *UsageInfo    `json:"usage,omitempty"`
}

// ContentItem represents an item in the message content array.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"` // For tool_use
}

// UsageInfo represents token usage information.
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToolCallInfo represents information about a tool call.
type ToolCallInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
