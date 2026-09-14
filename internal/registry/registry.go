package registry

import (
	"fmt"
	"net/http"
	"sort"
)

type Registry struct {
	routes map[string]http.HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{routes: make(map[string]http.HandlerFunc)}
}

func (r *Registry) Register(method, path string, handler http.HandlerFunc) error {
	key := method + " " + path
	if _, exists := r.routes[key]; exists {
		return fmt.Errorf("route %q already registered", key)
	}
	r.routes[key] = handler
	return nil
}

func (r *Registry) Build() *http.ServeMux {
	mux := http.NewServeMux()
	for pattern, handler := range r.routes {
		mux.HandleFunc(pattern, handler)
	}
	return mux
}

func (r *Registry) Routes() []string {
	routes := make([]string, 0, len(r.routes))
	for pattern := range r.routes {
		routes = append(routes, pattern)
	}
	sort.Strings(routes)
	return routes
}
