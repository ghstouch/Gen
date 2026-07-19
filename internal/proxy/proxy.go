package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/models"
)

// Client wraps HTTP client with timeout
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new proxy client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(config.WriteTimeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ProxyRequest forwards a chat completion request to a provider
func (c *Client) ProxyRequest(provider config.Provider, body []byte, headers http.Header) (*http.Response, error) {
	url := fmt.Sprintf("%s/chat/completions", provider.BaseURL)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	// Copy relevant headers from original request
	for _, key := range []string{"x-request-id", "x-stainless-os", "x-stainless-arch"} {
		if v := headers.Get(key); v != "" {
			req.Header.Set(key, v)
		}
	}

	return c.httpClient.Do(req)
}

// ProxyStream forwards a streaming chat completion request to a provider
func (c *Client) ProxyStream(provider config.Provider, body []byte, headers http.Header) (*http.Response, error) {
	url := fmt.Sprintf("%s/chat/completions", provider.BaseURL)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create stream request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	return c.httpClient.Do(req)
}

// HealthCheck pings the provider's models endpoint
func (c *Client) HealthCheck(provider config.Provider) error {
	url := fmt.Sprintf("%s/models", provider.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed [%d]: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ReadResponseBody reads and returns the response body
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// FilterModels returns only models supported by enabled providers
func FilterModels(providers []config.Provider) []models.Model {
	var modelList []models.Model
	seen := make(map[string]bool)

	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		for _, m := range p.Models {
			if !seen[m] {
				seen[m] = true
				modelList = append(modelList, models.Model{
					ID:      m,
					Object:  "model",
					Created: 1700000000,
					OwnedBy: p.Name,
				})
			}
		}
	}
	return modelList
}
