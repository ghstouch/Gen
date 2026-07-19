package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ghstouch/Gen/internal/config"
)

func TestNewRouter(t *testing.T) {
	providers := []config.Provider{
		{Name: "test1", BaseURL: "http://localhost:1", Priority: 2, Enabled: true},
		{Name: "test2", BaseURL: "http://localhost:2", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	if r == nil {
		t.Fatal("NewRouter() returned nil")
	}

	// Check providers are sorted by priority
	got := r.GetProviders()
	if len(got) != 2 {
		t.Fatalf("GetProviders count = %d, want 2", len(got))
	}
	if got[0].Name != "test2" {
		t.Errorf("First provider = %q, want %q (lower priority)", got[0].Name, "test2")
	}
}

func TestGetProviders(t *testing.T) {
	providers := []config.Provider{
		{Name: "a", Enabled: true},
		{Name: "b", Enabled: false},
	}

	r := NewRouter(providers)
	got := r.GetProviders()

	if len(got) != 2 {
		t.Errorf("GetProviders count = %d, want 2", len(got))
	}
}

func TestGetModels(t *testing.T) {
	providers := []config.Provider{
		{Name: "openai", Models: []string{"gpt-4"}, Enabled: true, Priority: 1},
	}

	r := NewRouter(providers)
	models := r.GetModels()

	if len(models) == 0 {
		t.Error("GetModels() returned empty")
	}
}

func TestRouteChatCompletion(t *testing.T) {
	// Create a mock provider server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "test-123",
			"model": "gpt-4",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "Hello"}},
			},
		})
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"Hi"}]}`)

	resp, err := r.RouteChatCompletion(body, http.Header{})
	if err != nil {
		t.Fatalf("RouteChatCompletion failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouteChatCompletionFallback(t *testing.T) {
	// First server fails
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer failServer.Close()

	// Second server succeeds
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "fallback-123",
			"model": "gpt-4",
		})
	}))
	defer successServer.Close()

	providers := []config.Provider{
		{Name: "fail", BaseURL: failServer.URL, Priority: 1, Enabled: true},
		{Name: "success", BaseURL: successServer.URL, Priority: 2, Enabled: true},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	resp, err := r.RouteChatCompletion(body, http.Header{})
	if err != nil {
		t.Fatalf("RouteChatCompletion with fallback failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify it used the fallback provider
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["id"] != "fallback-123" {
		t.Errorf("Expected fallback response, got %v", result["id"])
	}
}

func TestRouteChatStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"id\":\"stream-test\"}\n\n"))
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[],"stream":true}`)

	resp, providerName, err := r.RouteChatStream(body, http.Header{})
	if err != nil {
		t.Fatalf("RouteChatStream failed: %v", err)
	}
	defer resp.Body.Close()

	if providerName != "test" {
		t.Errorf("ProviderName = %q, want %q", providerName, "test")
	}
}

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	results := r.HealthCheck()

	if results["test"] != "healthy" {
		t.Errorf("HealthCheck[test] = %q, want %q", results["test"], "healthy")
	}
}

func TestStatsJSON(t *testing.T) {
	providers := []config.Provider{
		{Name: "test", BaseURL: "http://localhost:1", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	stats := r.StatsJSON()

	if len(stats) == 0 {
		t.Error("StatsJSON() returned empty")
	}

	var result []map[string]interface{}
	err := json.Unmarshal(stats, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Stats count = %d, want 1", len(result))
	}
}

func TestCooldownMechanism(t *testing.T) {
	// Create a server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	providers := []config.Provider{
		{Name: "test", BaseURL: server.URL, Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	// Make requests until cooldown triggers
	for i := 0; i < maxConsecutiveFailures; i++ {
		r.RouteChatCompletion(body, http.Header{})
	}

	// The provider should now be in cooldown
	// Get available providers should be empty
	available := r.getAvailableProviders()
	if len(available) != 0 {
		t.Errorf("Expected 0 available providers during cooldown, got %d", len(available))
	}

	// Stats should show cooldown
	stats := r.StatsJSON()
	var result []map[string]interface{}
	json.Unmarshal(stats, &result)

	if len(result) > 0 && !result[0]["in_cooldown"].(bool) {
		t.Error("Provider should be in cooldown")
	}
}

func TestRecordSuccessResetsFailure(t *testing.T) {
	providers := []config.Provider{
		{Name: "test", BaseURL: "http://localhost:1", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)

	// Record some failures
	r.recordFailure("test")
	r.recordFailure("test")

	// Then success
	r.recordSuccess("test")

	// Check failures are reset
	r.mu.RLock()
	failures := r.failures["test"]
	r.mu.RUnlock()

	if failures != 0 {
		t.Errorf("Failures = %d, want 0 after success", failures)
	}
}

func TestNoAvailableProviders(t *testing.T) {
	providers := []config.Provider{
		{Name: "disabled", BaseURL: "http://localhost:1", Priority: 1, Enabled: false},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	_, err := r.RouteChatCompletion(body, http.Header{})
	if err == nil {
		t.Error("Expected error when no providers available")
	}
}

func TestProviderSorting(t *testing.T) {
	providers := []config.Provider{
		{Name: "c", Priority: 3, Enabled: true},
		{Name: "a", Priority: 1, Enabled: true},
		{Name: "b", Priority: 2, Enabled: true},
	}

	r := NewRouter(providers)
	got := r.GetProviders()

	expected := []string{"a", "b", "c"}
	for i, name := range expected {
		if got[i].Name != name {
			t.Errorf("Provider[%d] = %q, want %q", i, got[i].Name, name)
		}
	}
}

func TestEmptyProviderList(t *testing.T) {
	r := NewRouter([]config.Provider{})

	if len(r.GetProviders()) != 0 {
		t.Error("Expected empty provider list")
	}

	if len(r.GetModels()) != 0 {
		t.Error("Expected empty model list")
	}
}

func TestProviderWithEmptyBaseURL(t *testing.T) {
	providers := []config.Provider{
		{Name: "empty", BaseURL: "", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	_, err := r.RouteChatCompletion(body, http.Header{})
	if err == nil {
		t.Error("Expected error for provider with empty BaseURL")
	}
}

func TestMultipleProvidersSamePriority(t *testing.T) {
	providers := []config.Provider{
		{Name: "a", BaseURL: "http://localhost:1", Priority: 1, Enabled: true},
		{Name: "b", BaseURL: "http://localhost:2", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)

	if len(r.GetProviders()) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(r.GetProviders()))
	}
}

func TestProviderDisableEnable(t *testing.T) {
	providers := []config.Provider{
		{Name: "test", BaseURL: "http://localhost:1", Priority: 1, Enabled: false},
	}

	r := NewRouter(providers)

	// Initially disabled
	available := r.getAvailableProviders()
	if len(available) != 0 {
		t.Error("Provider should be disabled")
	}

	// Enable it
	r.mu.Lock()
	r.providers[0].Enabled = true
	r.mu.Unlock()

	available = r.getAvailableProviders()
	if len(available) != 1 {
		t.Error("Provider should be enabled")
	}
}

func TestConcurrentAccess(t *testing.T) {
	providers := []config.Provider{
		{Name: "test", BaseURL: "http://localhost:1", Priority: 1, Enabled: true},
	}

	r := NewRouter(providers)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = r.GetProviders()
			_ = r.GetModels()
			_ = r.HealthCheck()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCooldownDuration(t *testing.T) {
	if cooldownDuration != 60*time.Second {
		t.Errorf("cooldownDuration = %v, want %v", cooldownDuration, 60*time.Second)
	}
}

func TestMaxConsecutiveFailures(t *testing.T) {
	if maxConsecutiveFailures != 3 {
		t.Errorf("maxConsecutiveFailures = %d, want 3", maxConsecutiveFailures)
	}
}
