package main

import (
	"API-mirror/api/ollama/handlers"
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"API-mirror/internal/registry"
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
		logger.Printf("failed to load tags store: %v", err)
	}
	chatStore := &fixtures.FixtureStore[types.ChatResponse]{}
	if err := chatStore.Load("../../fixtures/ollama/chat", "default"); err != nil {
		logger.Printf("failed to load chat store: %v", err)
	}

	reg := registry.NewRegistry()
	if err := reg.Register("GET", "/api/tags", handlers.TagsHandler(tagsStore)); err != nil {
		logger.Printf("failed to initiate new tags registry: %v", err)
	}
	if err := reg.Register("POST", "/api/chat", handlers.ChatHandler(chatStore)); err != nil {
		logger.Printf("failed to initial new chat registry: %v", err)
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
