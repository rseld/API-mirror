package server

import (
	"API-mirror/internal/config"
	"API-mirror/internal/registry"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ShutdownTimeout = 5 * time.Second

type Reloadable interface {
	Load(dir, active string) error
}

type ReloadEntry struct {
	Name  string
	Dir   string
	Store Reloadable
}

type ServerInstance struct {
	Name     string
	Config   config.InstanceConfig
	Server   *http.Server
	Registry *registry.Registry
	Reloads  []ReloadEntry
	Running  bool
}

type ServerEvent struct {
	Name string
	Err  error
}

func startServer(name string, server *http.Server, events chan<- ServerEvent) {
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			events <- ServerEvent{Name: name, Err: nil}
			return
		}
		events <- ServerEvent{Name: name, Err: err}
	}()
}

func (s *ServerInstance) Start(events chan<- ServerEvent) {
	startServer(s.Name, s.Server, events)
	s.Running = true
}

func (s *ServerInstance) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

func (s *ServerInstance) ReloadAll() error {
	for _, entry := range s.Reloads {
		if err := entry.Store.Load(entry.Dir, "default"); err != nil {
			return fmt.Errorf("reloading %s: %w", entry.Name, err)
		}
	}
	return nil
}
