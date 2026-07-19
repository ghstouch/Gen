package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/models"
	"github.com/ghstouch/Gen/internal/proxy"
)

// Router handles provider selection and fallback logic
type Router struct {
	providers []config.Provider
	client    *proxy.Client
	mu        sync.RWMutex
	cooldowns map[string]time.Time // provider -> cooldown until
	failures  map[string]int       // provider -> consecutive failures
}

const (
	maxConsecutiveFailures = 3
	cooldownDuration       = 60 * time.Second
)

// NewRouter creates a new router with the given providers
func NewRouter(providers []config.Provider) *Router {
	// Sort by priority
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	return &Router{
		providers: providers,
		client:    proxy.NewClient(),
		cooldowns: make(map[string]time.Time),
		failures:  make(map[string]int),
	}
}

// GetProviders returns the current provider list
func (r *Router) GetProviders() []config.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers
}

// GetModels returns all available models
func (r *Router) GetModels() []models.Model {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return proxy.FilterModels(r.providers)
}

// getAvailableProviders returns providers sorted by priority, skipping cooldown
func (r *Router) getAvailableProviders() []config.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var available []config.Provider
	now := time.Now()

	for _, p := range r.providers {
		if !p.Enabled {
			continue
		}
		if cd, ok := r.cooldowns[p.Name]; ok && now.Before(cd) {
			continue
		}
		available = append(available, p)
	}

	return available
}

// recordSuccess resets failure count for a provider
func (r *Router) recordSuccess(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, name)
	delete(r.cooldowns, name)
}

// recordFailure increments failure count and applies cooldown if needed
func (r *Router) recordFailure(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.failures[name]++
	if r.failures[name] >= maxConsecutiveFailures {
		r.cooldowns[name] = time.Now().Add(cooldownDuration)
		log.Printf("[Router] Provider %s in cooldown for %v after %d failures",
			name, cooldownDuration, r.failures[name])
	}
}

// RouteChatCompletion tries each provider in priority order (with fallback)
func (r *Router) RouteChatCompletion(body []byte, headers http.Header) (*http.Response, error) {
	providers := r.getAvailableProviders()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no available providers")
	}

	var lastErr error
	for _, provider := range providers {
		log.Printf("[Router] Trying provider: %s (priority %d)", provider.Name, provider.Priority)

		resp, err := r.client.ProxyRequest(provider, body, headers)
		if err != nil {
			lastErr = err
			log.Printf("[Router] Provider %s failed: %v", provider.Name, err)
			r.recordFailure(provider.Name)
			continue
		}

		// Check if the response is an error
		if resp.StatusCode >= 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("provider %s returned %d: %s", provider.Name, resp.StatusCode, string(bodyBytes))
			log.Printf("[Router] Provider %s error: %v", provider.Name, lastErr)
			r.recordFailure(provider.Name)
			continue
		}

		r.recordSuccess(provider.Name)
		log.Printf("[Router] Provider %s succeeded", provider.Name)
		return resp, nil
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// RouteChatStream tries streaming from each provider with fallback
func (r *Router) RouteChatStream(body []byte, headers http.Header) (*http.Response, string, error) {
	providers := r.getAvailableProviders()

	if len(providers) == 0 {
		return nil, "", fmt.Errorf("no available providers")
	}

	var lastErr error
	for _, provider := range providers {
		log.Printf("[Router] Trying stream from: %s", provider.Name)

		resp, err := r.client.ProxyStream(provider, body, headers)
		if err != nil {
			lastErr = err
			r.recordFailure(provider.Name)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("provider %s stream returned %d", provider.Name, resp.StatusCode)
			r.recordFailure(provider.Name)
			continue
		}

		r.recordSuccess(provider.Name)
		return resp, provider.Name, nil
	}

	return nil, "", fmt.Errorf("all providers failed for streaming: %w", lastErr)
}

// HealthCheck checks all provider health
func (r *Router) HealthCheck() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]string)
	for _, p := range r.providers {
		if err := r.client.HealthCheck(p); err != nil {
			results[p.Name] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			results[p.Name] = "healthy"
		}
	}
	return results
}

// MarshalJSON for router status
func (r *Router) StatsJSON() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type ProviderStat struct {
		Name       string `json:"name"`
		Priority   int    `json:"priority"`
		Enabled    bool   `json:"enabled"`
		Failures   int    `json:"failures"`
		InCooldown bool   `json:"in_cooldown"`
	}

	now := time.Now()
	var stats []ProviderStat
	for _, p := range r.providers {
		cd, inCD := r.cooldowns[p.Name]
		stat := ProviderStat{
			Name:       p.Name,
			Priority:   p.Priority,
			Enabled:    p.Enabled,
			Failures:   r.failures[p.Name],
			InCooldown: inCD && now.Before(cd),
		}
		stats = append(stats, stat)
	}

	data, _ := json.Marshal(stats)
	return data
}
