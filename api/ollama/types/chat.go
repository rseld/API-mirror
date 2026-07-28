package types

import "encoding/json"

type ChatRequest struct {
	Model     string          `json:"model"`
	Messages  []Message       `json:"messages"`
	Stream    *bool           `json:"stream,omitempty"`
	Format    json.RawMessage `json:"format,omitempty"`
	KeepAlive string          `json:"keep_alive,omitempty"`
	Options   map[string]any  `json:"options,omitempty"`
}

type ChatResponse struct {
	Model           string  `json:"model"`
	CreatedAt       string  `json:"created_at"`
	Message         Message `json:"message"`
	Done            bool    `json:"done"`
	DoneReason      string  `json:"done_reason,omitempty"`
	TotalDuration   int64   `json:"total_duration,omitempty"`
	LoadDuration    int64   `json:"load_duration,omitempty"`
	PromptEvalCount int64   `json:"prompt_eval_count,omitempty"`
	EvalCount       int64   `json:"eval_count,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	Images  []ImageData `json:"images,omitempty"`
}

type ImageData []byte
