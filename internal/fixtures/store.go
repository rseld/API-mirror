package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FixtureStore[T any] struct {
	mu      sync.Mutex
	entries map[string]T
	active  string
}

func (s *FixtureStore[T]) Get() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.entries[s.active]
	if !ok {
		var zero T
		return zero, fmt.Errorf("No fixture loaded for scenario %q", s.active)
	}

	return val, nil
}

func (s *FixtureStore[T]) Reload(dir string) error {
	entries, err := loadDir[T](dir)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.entries = entries
	s.mu.Unlock()

	return nil
}

func (s *FixtureStore[T]) SetActive(scenario string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.entries[scenario]; !ok {
		return fmt.Errorf("no fixture %q loaded for this store", scenario)
	}
	s.active = scenario

	return nil
}

func loadDir[T any](dir string) (map[string]T, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]T)
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", f.Name(), err)
		}

		var val T
		if err := json.Unmarshal(data, &val); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f.Name(), err)
		}

		key := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		entries[key] = val
	}

	return entries, nil
}

func (s *FixtureStore[T]) Load(dir, active string) error {
	entries, err := loadDir[T](dir)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = entries
	s.active = active
	return nil
}
