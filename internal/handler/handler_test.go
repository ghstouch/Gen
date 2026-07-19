package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/models"
	"github.com/ghstouch/Gen/internal/router"
)

func setupTestHandler() *Handler {
	providers := []config.Provider{
		{Name: "test", BaseURL: "http://localhost:1", Priority: 1, Enabled: true},
	}
	r := router.NewRouter(providers)
	return NewHandler(r)
}

func TestNewHandler(t *testing.T) {
	h := setupTestHandler()
	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if h.router == nil {
		t.Fatal("router is nil")
	}
	if h.started.IsZero() {
		t.Fatal("started time is zero")
	}
}

func TestChatCompletionsMethodNotAllowed(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest("GET", "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestChatCompletionsInvalidJSON(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString("invalid"))
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletionsMissingModel(t *testing.T) {
	h := setupTestHandler()

	body := `{"messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletionsSuccess(t *testing.T) {
	// Create mock provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ChatResponse{
			ID:    "test-123",
			Model: "gpt-4",
			Choices: []models.ChatChoice{
				{Index: 0, Message: models.ChatMessage{Role: "assistant", Content: "Hello!"}},
			},
		})
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}
	r := router.NewRouter(providers)
	h := NewHandler(r)

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}

	var resp models.ChatResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != "test-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "test-123")
	}
}

func TestChatCompletionsStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"id\":\"stream-123\"}\n\n"))
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}
	r := router.NewRouter(providers)
	h := NewHandler(r)

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), "text/event-stream")
	}
}

func TestModelsEndpoint(t *testing.T) {
	providers := []config.Provider{
		{Name: "openai", Models: []string{"gpt-4"}, Enabled: true, Priority: 1},
	}
	r := router.NewRouter(providers)
	h := NewHandler(r)

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	h.Models(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}

	var resp models.ModelList
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Object != "list" {
		t.Errorf("Object = %q, want %q", resp.Object, "list")
	}
}

func TestHealthEndpoint(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}

	var resp models.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
	if resp.Port != config.Port {
		t.Errorf("Port = %d, want %d", resp.Port, config.Port)
	}
}

func TestStatsEndpoint(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	h.Stats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), "application/json")
	}
}

func TestCorsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// corsMiddleware is in main.go, test CORS via the actual handler
	wrapped := CorsMiddleware(handler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS Allow-Origin = %q, want %q", w.Header().Get("Access-Control-Allow-Origin"), "*")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS Allow-Methods should not be empty")
	}
	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCorsMiddlewarePassThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// corsMiddleware is in main.go, test CORS via the actual handler
	wrapped := CorsMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "ok" {
		t.Errorf("Body = %q, want %q", w.Body.String(), "ok")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), "application/json")
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["key"] != "value" {
		t.Errorf("key = %q, want %q", result["key"], "value")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp models.ChatResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Message != "test error" {
		t.Errorf("ErrorMessage = %q, want %q", resp.Error.Message, "test error")
	}
}

func TestHandlerStartTime(t *testing.T) {
	before := time.Now()
	h := setupTestHandler()
	after := time.Now()

	if h.started.Before(before) || h.started.After(after) {
		t.Error("started time should be between before and after creation")
	}
}

func TestUptimeInHealth(t *testing.T) {
	h := setupTestHandler()

	// Small delay to ensure uptime > 0
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	var resp models.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Uptime == "" {
		t.Error("Uptime should not be empty")
	}
}

func TestProviderCountInHealth(t *testing.T) {
	providers := []config.Provider{
		{Name: "a", Enabled: true, Priority: 1},
		{Name: "b", Enabled: false, Priority: 2},
		{Name: "c", Enabled: true, Priority: 3},
	}
	r := router.NewRouter(providers)
	h := NewHandler(r)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	var resp models.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Only enabled providers should be counted
	if resp.Providers != 2 {
		t.Errorf("Providers = %d, want 2", resp.Providers)
	}
}

func TestX_RoutedByHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ChatResponse{
			ID:    "test",
			Model: "gpt-4",
		})
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}
	r := router.NewRouter(providers)
	h := NewHandler(r)

	body := `{"model":"gpt-4","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ChatCompletions(w, req)

	if w.Header().Get("X-Routed-By") == "" {
		t.Error("X-Routed-By header should be set")
	}
}
