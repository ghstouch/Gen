package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/handler"
	"github.com/ghstouch/Gen/internal/router"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load providers from env or use defaults
	providers := loadProviders()

	// Create router
	r := router.NewRouter(providers)

	// Create handler
	h := handler.NewHandler(r)

	// Setup routes
	mux := http.NewServeMux()

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/chat/completions", h.ChatCompletions)
	mux.HandleFunc("/v1/models", h.Models)

	// Health & stats
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/stats", h.Stats)

	// CORS middleware wrapper
	handler := corsMiddleware(mux)

	// Server
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("========================================")
	log.Printf("  Gen - AI Router (Go)")
	log.Printf("  Port: %d", config.Port)
	log.Printf("  Providers: %d", len(providers))
	log.Printf("  Endpoints:")
	log.Printf("    POST /v1/chat/completions")
	log.Printf("    GET  /v1/models")
	log.Printf("    GET  /health")
	log.Printf("    GET  /stats")
	log.Printf("========================================")

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-request-id")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loadProviders from env vars or use defaults
// Example env vars:
//
//	GEN_OPENAI_KEY=sk-xxx
//	GEN_GROQ_KEY=gsk_xxx
//	GEN_TOGETHER_KEY=xxx
//	GEN_OPENROUTER_KEY=sk-or-xxx
//	GEN_GEMINI_KEY=xxx
func loadProviders() []config.Provider {
	providers := config.DefaultProviders()

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
