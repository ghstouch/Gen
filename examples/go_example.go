package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "http://localhost:2500"

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   *Usage       `json:"usage,omitempty"`
}

type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Delta   *ChatDelta  `json:"delta,omitempty"`
}

type ChatDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ModelList struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	Providers int    `json:"providers"`
	Port      int    `json:"port"`
}

type ProviderStat struct {
	Name       string `json:"name"`
	Priority   int    `json:"priority"`
	Enabled    bool   `json:"enabled"`
	Failures   int    `json:"failures"`
	InCooldown bool   `json:"in_cooldown"`
}

func chatCompletion(model, message string, stream bool) {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)

	req := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "user", Content: message},
		},
		Stream: stream,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Message: %s\n", message)
	fmt.Println(strings.Repeat("-", 50))

	if stream {
		// Streaming response
		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("Error reading stream: %v\n", err)
				return
			}

			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				fmt.Println("\n[DONE]")
				break
			}

			var chunk ChatResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		}

		fmt.Println()
	} else {
		// Non-streaming response
		var result ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		if len(result.Choices) > 0 {
			fmt.Println(result.Choices[0].Message.Content)
		}

		if result.Usage != nil {
			fmt.Printf("\nTokens: %d (prompt: %d, completion: %d)\n",
				result.Usage.TotalTokens,
				result.Usage.PromptTokens,
				result.Usage.CompletionTokens)
		}
	}
}

func listModels() {
	url := fmt.Sprintf("%s/v1/models", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result ModelList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	fmt.Println("Available Models:")
	fmt.Println(strings.Repeat("-", 50))

	for _, model := range result.Data {
		fmt.Printf("  %s (by %s)\n", model.ID, model.OwnedBy)
	}
}

func healthCheck() {
	url := fmt.Sprintf("%s/health", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	fmt.Println("Health Status:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Version: %s\n", result.Version)
	fmt.Printf("  Uptime: %s\n", result.Uptime)
	fmt.Printf("  Providers: %d\n", result.Providers)
	fmt.Printf("  Port: %d\n", result.Port)
}

func providerStats() {
	url := fmt.Sprintf("%s/stats", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result []ProviderStat
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	fmt.Println("Provider Statistics:")
	fmt.Println(strings.Repeat("-", 50))

	for _, provider := range result {
		status := "✓"
		if !provider.Enabled {
			status = "✗"
		}
		cooldown := ""
		if provider.InCooldown {
			cooldown = " (cooldown)"
		}
		fmt.Printf("  %s %s (priority: %d, failures: %d)%s\n",
			status, provider.Name, provider.Priority, provider.Failures, cooldown)
	}
}

func main() {
	fmt.Println("Gen AI Router - Go Example")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Health check
	healthCheck()
	fmt.Println()

	// List models
	listModels()
	fmt.Println()

	// Provider stats
	providerStats()
	fmt.Println()

	// Chat completion (non-streaming)
	fmt.Println("Chat Completion (Non-streaming):")
	fmt.Println(strings.Repeat("=", 50))
	chatCompletion("gpt-4", "What is Go programming language?", false)
	fmt.Println()

	// Chat completion (streaming)
	fmt.Println("Chat Completion (Streaming):")
	fmt.Println(strings.Repeat("=", 50))
	chatCompletion("gpt-4", "Write a short poem about coding.", true)
}
