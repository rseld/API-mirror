package ollama

import (
	"API-mirror/api/ollama/handlers"
	"API-mirror/api/ollama/types"
	"API-mirror/internal/config"
	"API-mirror/internal/fixtures"
	"API-mirror/internal/registry"
	"API-mirror/internal/server"
	"fmt"
	"net/http"
)

func NewInstance(cfg config.InstanceConfig) (*server.ServerInstance, error) {
	tagsStore := &fixtures.FixtureStore[ollamatypes.ListResponse]{}
	if err := tagsStore.Load(cfg.FixturesDir+"/tags", "default"); err != nil {
		return nil, err
	}

	chatStore := &fixtures.FixtureStore[ollamatypes.ChatResponse]{}
	if err := chatStore.Load(cfg.FixturesDir+"/chat", "default"); err != nil {
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
	serv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return &server.ServerInstance{
		Name:     cfg.Name,
		Config:   cfg,
		Server:   serv,
		Registry: reg,
		Reloads: []server.ReloadEntry{
			{Name: "tags", Dir: cfg.FixturesDir + "/tags", Store: tagsStore},
			{Name: "chat", Dir: cfg.FixturesDir + "/chat", Store: chatStore},
		},
	}, nil
}
