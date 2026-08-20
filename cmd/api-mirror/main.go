package main

import (
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"API-mirror/internal/registry"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	file, err := os.OpenFile("server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.Ldate|log.Ltime)
	logger.Println("Server logging started")

	tagsStore := &fixtures.FixtureStore[types.ListResponse]{}
	if err := tagsStore.Load("../../fixtures/ollama/tags", "default"); err != nil {
		logger.Printf("failed to load store: %v", err)
	}

	reg := registry.NewRegistry()
	if err := reg.Register("GET", "/api/tags", TagsHandler(tagsStore)); err != nil {
		logger.Println("failed to initiate new registry")
	}

	mux := reg.Build()

	server := &http.Server{
		Addr:    ":11434",
		Handler: mux,
	}

	logger.Println("listening on :11434")
	if err := server.ListenAndServe(); err != nil {
		logger.Println("server failed to start")
	}
}

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
