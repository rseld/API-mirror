package ollamahandlers

import (
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"API-mirror/internal/logging"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func TagsHandler(store *fixtures.FixtureStore[ollamatypes.ListResponse]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logTagsRequest()
		resp, err := store.Get()
		if err != nil {
			http.Error(w, "Fixture unavailable", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		logTagsResponse(resp)
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
		logChatRequest(req)

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

	logChatResponse(resp)
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

	logStreamStart(req)
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
	logStreamEnd(req.Model, len(words))
}

func logChatRequest(req ollamatypes.ChatRequest) {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Printf("chat request: model = %s messages = %d stream = %t",
		req.Model, len(req.Messages), *req.Stream)
}

func logChatResponse(resp ollamatypes.ChatResponse) {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Printf("chat response: model = %s done = %v content = %q",
		resp.Model, resp.Done, resp.Message.Content)
}

func logTagsRequest() {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Println("tags request received")
}

func logTagsResponse(resp ollamatypes.ListResponse) {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Printf("tags response: %d models", len(resp.Models))
}

func logStreamStart(req ollamatypes.ChatRequest) {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Printf("stream started: model = %s", req.Model)
}

func logStreamEnd(model string, chunkCount int) {
	if !logging.Enabled(logging.Verbose) {
		return
	}
	log.Printf("stream ended: model = %s chunks = %d", model, chunkCount)
}
