package models

// ChatRequest - OpenAI-compatible chat completion request
type ChatRequest struct {
	Model            string         `json:"model"`
	Messages         []ChatMessage  `json:"messages"`
	Stream           bool           `json:"stream,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	N                int            `json:"n,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	User             string         `json:"user,omitempty"`
}

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse - OpenAI-compatible chat completion response
type ChatResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []ChatChoice       `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"`
	Error   *ErrorDetail       `json:"error,omitempty"`
}

// ChatChoice represents a single completion choice
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message,omitempty"`
	Delta        *ChatDelta   `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
}

// ChatDelta for streaming responses
type ChatDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorDetail for API errors
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ModelList - list of available models
type ModelList struct {
	Object string   `json:"object"`
	Data   []Model  `json:"data"`
}

// Model represents a single model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// HealthResponse for health check
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	Providers int    `json:"providers"`
	Port      int    `json:"port"`
}