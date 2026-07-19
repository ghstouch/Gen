package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/handler"
	"github.com/ghstouch/Gen/internal/router"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load providers from env or use defaults
	providers := config.LoadProviders()

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
	handl := handler.CorsMiddleware(mux)

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
		Handler:      handl,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
