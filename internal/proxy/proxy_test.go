package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
	if client.httpClient.Timeout != time.Duration(config.WriteTimeout)*time.Second {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, time.Duration(config.WriteTimeout)*time.Second)
	}
}

func TestProxyRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/json")
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ChatResponse{
			ID:    "test-123",
			Model: "gpt-4",
			Choices: []models.ChatChoice{
				{Index: 0, Message: models.ChatMessage{Role: "assistant", Content: "Hello"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient()
	provider := config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Enabled: true,
	}

	body := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}`)
	resp, err := client.ProxyRequest(provider, body, http.Header{})
	if err != nil {
		t.Fatalf("ProxyRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result models.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if result.ID != "test-123" {
		t.Errorf("ID = %q, want %q", result.ID, "test-123")
	}
}

func TestProxyRequestWithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-api-key")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.ChatResponse{ID: "ok"})
	}))
	defer server.Close()

	client := NewClient()
	provider := config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Enabled: true,
	}

	body := []byte(`{"model":"gpt-4","messages":[]}`)
	resp, err := client.ProxyRequest(provider, body, http.Header{})
	if err != nil {
		t.Fatalf("ProxyRequest failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestProxyStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Accept = %q, want %q", r.Header.Get("Accept"), "text/event-stream")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"id\":\"test\"}\n\n"))
	}))
	defer server.Close()

	client := NewClient()
	provider := config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Enabled: true,
	}

	body := []byte(`{"model":"gpt-4","messages":[],"stream":true}`)
	resp, err := client.ProxyStream(provider, body, http.Header{})
	if err != nil {
		t.Fatalf("ProxyStream failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("Path = %q, want %q", r.URL.Path, "/models")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient()
	provider := config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Enabled: true,
	}

	err := client.HealthCheck(provider)
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheckFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient()
	provider := config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Enabled: true,
	}

	err := client.HealthCheck(provider)
	if err == nil {
		t.Error("HealthCheck should fail for 500 response")
	}
}

func TestReadResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test body content"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	body, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("ReadResponseBody failed: %v", err)
	}

	if string(body) != "test body content" {
		t.Errorf("Body = %q, want %q", string(body), "test body content")
	}
}

func TestFilterModels(t *testing.T) {
	providers := []config.Provider{
		{
			Name:    "openai",
			Models:  []string{"gpt-4", "gpt-3.5-turbo"},
			Enabled: true,
		},
		{
			Name:    "groq",
			Models:  []string{"llama-3", "gpt-4"}, // duplicate
			Enabled: true,
		},
		{
			Name:    "disabled",
			Models:  []string{"model-x"},
			Enabled: false,
		},
	}

	models := FilterModels(providers)

	// Should have 3 unique models (gpt-4, gpt-3.5-turbo, llama-3)
	if len(models) != 3 {
		t.Errorf("FilterModels count = %d, want 3", len(models))
	}

	// Check no duplicates
	seen := make(map[string]bool)
	for _, m := range models {
		if seen[m.ID] {
			t.Errorf("Duplicate model: %s", m.ID)
		}
		seen[m.ID] = true
	}

	// Check disabled provider's models are excluded
	for _, m := range models {
		if m.ID == "model-x" {
			t.Error("Disabled provider's model should be excluded")
		}
	}
}
