package models

import (
	"encoding/json"
	"testing"
)

func TestChatRequestJSON(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ChatRequest: %v", err)
	}

	var result ChatRequest
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ChatRequest: %v", err)
	}

	if result.Model != req.Model {
		t.Errorf("Model = %q, want %q", result.Model, req.Model)
	}
	if len(result.Messages) != len(req.Messages) {
		t.Errorf("Messages count = %d, want %d", len(result.Messages), len(req.Messages))
	}
	if result.Messages[0].Role != "user" {
		t.Errorf("Message role = %q, want %q", result.Messages[0].Role, "user")
	}
	if !result.Stream {
		t.Error("Stream should be true")
	}
}

func TestChatMessageJSON(t *testing.T) {
	msg := ChatMessage{
		Role:    "assistant",
		Content: "Hello! How can I help?",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal ChatMessage: %v", err)
	}

	var result ChatMessage
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ChatMessage: %v", err)
	}

	if result.Role != msg.Role {
		t.Errorf("Role = %q, want %q", result.Role, msg.Role)
	}
	if result.Content != msg.Content {
		t.Errorf("Content = %q, want %q", result.Content, msg.Content)
	}
}

func TestChatResponseJSON(t *testing.T) {
	resp := ChatResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1700000000,
		Model:   "gpt-4",
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: &Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ChatResponse: %v", err)
	}

	var result ChatResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ChatResponse: %v", err)
	}

	if result.ID != resp.ID {
		t.Errorf("ID = %q, want %q", result.ID, resp.ID)
	}
	if len(result.Choices) != 1 {
		t.Errorf("Choices count = %d, want 1", len(result.Choices))
	}
	if result.Choices[0].FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", result.Choices[0].FinishReason, "stop")
	}
	if result.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %d, want 15", result.Usage.TotalTokens)
	}
}

func TestChatDeltaJSON(t *testing.T) {
	delta := ChatDelta{
		Role:    "assistant",
		Content: "Hello",
	}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("Failed to marshal ChatDelta: %v", err)
	}

	var result ChatDelta
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ChatDelta: %v", err)
	}

	if result.Role != delta.Role {
		t.Errorf("Role = %q, want %q", result.Role, delta.Role)
	}
	if result.Content != delta.Content {
		t.Errorf("Content = %q, want %q", result.Content, delta.Content)
	}
}

func TestModelListJSON(t *testing.T) {
	modelList := ModelList{
		Object: "list",
		Data: []Model{
			{ID: "gpt-4", Object: "model", Created: 1700000000, OwnedBy: "openai"},
			{ID: "gpt-3.5-turbo", Object: "model", Created: 1700000000, OwnedBy: "openai"},
		},
	}

	data, err := json.Marshal(modelList)
	if err != nil {
		t.Fatalf("Failed to marshal ModelList: %v", err)
	}

	var result ModelList
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ModelList: %v", err)
	}

	if result.Object != "list" {
		t.Errorf("Object = %q, want %q", result.Object, "list")
	}
	if len(result.Data) != 2 {
		t.Errorf("Data count = %d, want 2", len(result.Data))
	}
}

func TestHealthResponseJSON(t *testing.T) {
	health := HealthResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Uptime:    "5m30s",
		Providers: 3,
		Port:      2500,
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("Failed to marshal HealthResponse: %v", err)
	}

	var result HealthResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal HealthResponse: %v", err)
	}

	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q", result.Status, "ok")
	}
	if result.Port != 2500 {
		t.Errorf("Port = %d, want 2500", result.Port)
	}
}

func TestErrorDetailJSON(t *testing.T) {
	errDetail := ErrorDetail{
		Message: "Invalid API key",
		Type:    "authentication_error",
		Code:    "401",
	}

	data, err := json.Marshal(errDetail)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorDetail: %v", err)
	}

	var result ErrorDetail
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ErrorDetail: %v", err)
	}

	if result.Message != errDetail.Message {
		t.Errorf("Message = %q, want %q", result.Message, errDetail.Message)
	}
	if result.Type != errDetail.Type {
		t.Errorf("Type = %q, want %q", result.Type, errDetail.Type)
	}
}
