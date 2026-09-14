package ollama

import (
	"API-mirror/api/ollama/handlers"
	"API-mirror/api/ollama/types"
	"API-mirror/internal/fixtures"
	"API-mirror/internal/registry"
	"API-mirror/internal/server"
	"net/http"
)

func NewInstance() (*server.ServerInstance, error) {
	tagsStore := &fixtures.FixtureStore[ollamatypes.ListResponse]{}
	if err := tagsStore.Load("../../fixtures/ollama/tags", "default"); err != nil {
		return nil, err
	}

	chatStore := &fixtures.FixtureStore[ollamatypes.ChatResponse]{}
	if err := chatStore.Load("../../fixtures/ollama/chat", "default"); err != nil {
		return nil, err
	}

	reg := registry.NewRegistry()
	if err := reg.Register("GET", "/api/tags", ollamahandlers.TagsHandler(tagsStore)); err != nil {
		return nil, err
	}
	if err := reg.Register("POST", "/api/chat", ollamahandlers.ChatHandler(chatStore)); err != nil {
		return nil, err
	}

	mux := reg.Build()
	instance := &http.Server{
		Addr:    ":11434",
		Handler: mux,
	}

	return &server.ServerInstance{
		Name:     "ollama",
		Server:   instance,
		Registry: reg,
		Reloads: []server.ReloadEntry{
			{Name: "tags", Dir: "fixtures/ollama/tags", Store: tagsStore},
			{Name: "chat", Dir: "fixtures/ollama/chat", Store: chatStore},
		},
	}, nil
}
