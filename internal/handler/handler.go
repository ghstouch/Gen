package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ghstouch/Gen/internal/config"
	"github.com/ghstouch/Gen/internal/models"
	"github.com/ghstouch/Gen/internal/router"
)

// Handler holds all HTTP handlers
type Handler struct {
	router  *router.Router
	started time.Time
}

// NewHandler creates a new handler
func NewHandler(r *router.Router) *Handler {
	return &Handler{
		router:  r,
		started: time.Now(),
	}
}

// ChatCompletions handles POST /v1/chat/completions
func (h *Handler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, config.MaxBodySize))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req models.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "model is required")
		return
	}

	log.Printf("[Handler] Chat request: model=%s stream=%v messages=%d",
		req.Model, req.Stream, len(req.Messages))

	if req.Stream {
		h.handleStream(w, body, req.Model, r.Header)
	} else {
		h.handleNonStream(w, body, req.Model, r.Header)
	}
}

func (h *Handler) handleNonStream(w http.ResponseWriter, body []byte, model string, headers http.Header) {
	resp, err := h.router.RouteChatCompletion(body, headers)
	if err != nil {
		log.Printf("[Handler] All providers failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("all providers failed: %v", err))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Routed-By", "Gen/Go")
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

func (h *Handler) handleStream(w http.ResponseWriter, body []byte, model string, headers http.Header) {
	resp, providerName, err := h.router.RouteChatStream(body, headers)
	if err != nil {
		log.Printf("[Handler] Stream all providers failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("all providers failed: %v", err))
		return
	}
	defer resp.Body.Close()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Routed-By", fmt.Sprintf("Gen/Go (%s)", providerName))
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("[Handler] Streaming not supported")
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}

// Models handles GET /v1/models
func (h *Handler) Models(w http.ResponseWriter, r *http.Request) {
	modelList := h.router.GetModels()

	resp := models.ModelList{
		Object: "list",
		Data:   modelList,
	}

	writeJSON(w, http.StatusOK, resp)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	providers := h.router.GetProviders()
	enabled := 0
	for _, p := range providers {
		if p.Enabled {
			enabled++
		}
	}

	resp := models.HealthResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Uptime:    time.Since(h.started).Round(time.Second).String(),
		Providers: enabled,
		Port:      config.Port,
	}

	writeJSON(w, http.StatusOK, resp)
}

// Stats handles GET /stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(h.router.StatsJSON())
}

// writeJSON helper
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError helper
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ChatResponse{
		Error: &models.ErrorDetail{
			Message: message,
			Type:    "server_error",
			Code:    fmt.Sprintf("%d", status),
		},
	})
}

// CorsMiddleware adds CORS headers
func CorsMiddleware(next http.Handler) http.Handler {
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
