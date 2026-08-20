package handlers

import (
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func TagsHandler(store *fixtures.FixtureStore[types.ListResponse]) http.HandlerFunc {
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

func ChatHandler(store *fixtures.FixtureStore[types.ChatResponse]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Stream == nil || *req.Stream {
			http.Error(w, "streaming not implemented yet", http.StatusNotImplemented)
			return
		}

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
}
