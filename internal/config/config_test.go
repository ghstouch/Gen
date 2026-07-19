package config

import (
	"os"
	"testing"
)

func TestDefaultProviders(t *testing.T) {
	providers := DefaultProviders()
	if len(providers) == 0 {
		t.Error("DefaultProviders() returned empty slice")
	}

	// Check each provider has required fields
	for i, p := range providers {
		if p.Name == "" {
			t.Errorf("Provider[%d].Name is empty", i)
		}
		if p.BaseURL == "" {
			t.Errorf("Provider[%d].BaseURL is empty", i)
		}
		if p.Priority < 0 {
			t.Errorf("Provider[%d].Priority is negative: %d", i, p.Priority)
		}
	}

	// Check priority ordering
	for i := 1; i < len(providers); i++ {
		if providers[i].Priority < providers[i-1].Priority {
			t.Errorf("Providers not sorted by priority: [%d]=%d < [%d]=%d",
				i, providers[i].Priority, i-1, providers[i-1].Priority)
		}
	}
}

func TestLoadProvidersWithEnv(t *testing.T) {
	// Set test environment variables
	os.Setenv("GEN_OPENAI_KEY", "test-key-123")
	defer os.Unsetenv("GEN_OPENAI_KEY")

	providers := LoadProviders()

	// Find OpenAI provider
	var openai *Provider
	for i := range providers {
		if providers[i].Name == "openai" {
			openai = &providers[i]
			break
		}
	}

	if openai == nil {
		t.Fatal("OpenAI provider not found")
	}
	if openai.APIKey != "test-key-123" {
		t.Errorf("OpenAI API key = %q, want %q", openai.APIKey, "test-key-123")
	}
}

func TestLoadProvidersWithDisable(t *testing.T) {
	os.Setenv("GEN_GROQ_DISABLE", "1")
	defer os.Unsetenv("GEN_GROQ_DISABLE")

	providers := LoadProviders()

	for _, p := range providers {
		if p.Name == "groq" && p.Enabled {
			t.Error("Groq should be disabled when GEN_GROQ_DISABLE=1")
		}
	}
}

func TestProviderEnabled(t *testing.T) {
	providers := DefaultProviders()
	for i, p := range providers {
		if !p.Enabled {
			t.Errorf("Provider[%d] (%s) should be enabled by default", i, p.Name)
		}
	}
}

func TestProviderModels(t *testing.T) {
	providers := DefaultProviders()
	for i, p := range providers {
		if len(p.Models) == 0 {
			t.Errorf("Provider[%d] (%s) has no models", i, p.Name)
		}
	}
}

func TestPortConstant(t *testing.T) {
	if Port != 2500 {
		t.Errorf("Port = %d, want 2500", Port)
	}
}

func TestTimeoutConstants(t *testing.T) {
	if ReadTimeout <= 0 {
		t.Error("ReadTimeout should be positive")
	}
	if WriteTimeout <= 0 {
		t.Error("WriteTimeout should be positive")
	}
	if IdleTimeout <= 0 {
		t.Error("IdleTimeout should be positive")
	}
	if MaxBodySize <= 0 {
		t.Error("MaxBodySize should be positive")
	}
}
