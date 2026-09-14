package ollamahandlers

import (
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func TagsHandler(store *fixtures.FixtureStore[ollamatypes.ListResponse]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := store.Get()
		if err != nil {
			http.Error(w, "Fixture unavailable", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("failed to encode tags response: %v", err)
		}
	}
}

func ChatHandler(store *fixtures.FixtureStore[ollamatypes.ChatResponse]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ollamatypes.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Stream == nil || *req.Stream {
			handleStreaming(w, req, store)
		} else {
			handleNonStreaming(w, req, store)
		}

	}
}

func handleNonStreaming(w http.ResponseWriter, req ollamatypes.ChatRequest, store *fixtures.FixtureStore[ollamatypes.ChatResponse]) {

	resp, err := store.Get()
	if err != nil {
		http.Error(w, "fixture unavailable", http.StatusInternalServerError)
		return
	}

	resp.Model = req.Model
	resp.CreatedAt = time.Now().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode chat response: %v", err)
	}
}

func handleStreaming(w http.ResponseWriter, req ollamatypes.ChatRequest, store *fixtures.FixtureStore[ollamatypes.ChatResponse]) {

	resp, err := store.Get()
	if err != nil {
		http.Error(w, "fixture unavailable", http.StatusInternalServerError)
		return
	}

	// check if w supports flushing beforehand
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")

	words := strings.Fields(resp.Message.Content)
	encode := json.NewEncoder(w)

	for _, word := range words {
		chunk := ollamatypes.ChatResponse{
			Model:     req.Model,
			CreatedAt: time.Now().Format(time.RFC3339),
			Message: ollamatypes.Message{
				Role:    resp.Message.Role,
				Content: word + " ",
			},
			Done: false,
		}

		if err := encode.Encode(chunk); err != nil {
			log.Printf("failed to encode chat chunk: %v", err)
			return
		}

		flusher.Flush()

		time.Sleep(100 * time.Millisecond)
	}

	final := ollamatypes.ChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: ollamatypes.Message{
			Role:    resp.Message.Role,
			Content: "",
		},
		Done:       true,
		DoneReason: resp.DoneReason,
	}

	if err := encode.Encode(final); err != nil {
		log.Printf("failed to encode final chat chunk: %v", err)
	}
}
