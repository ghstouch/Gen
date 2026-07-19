package config

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	// Port - hardcoded server port
	Port = 2500

	// Server settings
	ReadTimeout  = 120
	WriteTimeout = 300
	IdleTimeout  = 120
	MaxBodySize  = 10 << 20 // 10MB
)

// Provider represents an AI provider configuration
type Provider struct {
	Name     string   `json:"name"`
	BaseURL  string   `json:"base_url"`
	APIKey   string   `json:"api_key"`
	Models   []string `json:"models"`
	Priority int      `json:"priority"` // lower = higher priority
	Enabled  bool     `json:"enabled"`
}

// DefaultProviders returns the default fallback provider list
func DefaultProviders() []Provider {
	return []Provider{
		{
			Name:     "openai",
			BaseURL:  "https://api.openai.com/v1",
			Models:   []string{"gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"},
			Priority: 1,
			Enabled:  true,
		},
		{
			Name:     "groq",
			BaseURL:  "https://api.groq.com/openai/v1",
			Models:   []string{"llama-3.3-70b-versatile", "mixtral-8x7b-32768"},
			Priority: 2,
			Enabled:  true,
		},
		{
			Name:     "together",
			BaseURL:  "https://api.together.xyz/v1",
			Models:   []string{"meta-llama/Llama-3.3-70B-Instruct-Turbo"},
			Priority: 3,
			Enabled:  true,
		},
		{
			Name:     "openrouter",
			BaseURL:  "https://openrouter.ai/api/v1",
			Models:   []string{"openai/gpt-4o", "anthropic/claude-3.5-sonnet"},
			Priority: 4,
			Enabled:  true,
		},
		{
			Name:     "gemini",
			BaseURL:  "https://generativelanguage.googleapis.com/v1beta/openai",
			Models:   []string{"gemini-2.0-flash", "gemini-1.5-pro"},
			Priority: 5,
			Enabled:  true,
		},
	}
}

// LoadProviders from env vars or use defaults
// Example env vars:
//
//	GEN_OPENAI_KEY=sk-xxx
//	GEN_GROQ_KEY=gsk_xxx
//	GEN_TOGETHER_KEY=xxx
//	GEN_OPENROUTER_KEY=sk-or-xxx
//	GEN_GEMINI_KEY=xxx
func LoadProviders() []Provider {
	providers := DefaultProviders()

	// Sort by priority
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	for i := range providers {
		envKey := fmt.Sprintf("GEN_%s_KEY", strings.ToUpper(providers[i].Name))
		if key := os.Getenv(envKey); key != "" {
			providers[i].APIKey = key
			log.Printf("[Config] Using API key for %s from %s", providers[i].Name, envKey)
		}

		// Check if provider is explicitly disabled
		disableKey := fmt.Sprintf("GEN_%s_DISABLE", strings.ToUpper(providers[i].Name))
		if os.Getenv(disableKey) == "1" || os.Getenv(disableKey) == "true" {
			providers[i].Enabled = false
			log.Printf("[Config] Provider %s disabled via %s", providers[i].Name, disableKey)
		}
	}

	return providers
}
