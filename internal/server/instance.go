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
	Instance *ServerInstance
	Err      error
}

/* Some design consequences that went unnoticed during the initial architecting
led to this function now crossing the boundary of no package should mutate objects that don't
belong to them. instance belongs to the cli package and should not be mutated here. A hard
limitation for preventing this will be investigated later */

func startServer(instance *ServerInstance, events chan<- ServerEvent) {
	go func() {
		err := instance.Server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			events <- ServerEvent{Instance: instance, Err: nil}
			return
		}
		events <- ServerEvent{Instance: instance, Err: err}
	}()
}

func (s *ServerInstance) Start(events chan<- ServerEvent) {
	s.Running = true
	startServer(s, events)
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
