package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type InstanceConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Port        int    `json:"port"`
	FixturesDir string `json:"fixtures_dir"`
}

type Config struct {
	Instances []InstanceConfig `json:"instances"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// default config only defined for initial providers, any other API shapes will need to be explicitly configured
func Default() *Config {
	return &Config{
		Instances: []InstanceConfig{
			{
				Name:        "ollama",
				Type:        "ollama",
				Port:        11434,
				FixturesDir: "../../fixtures/ollama",
			},
		},
	}
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		log.Printf("config: using default (could not laod %q; %v)", path, err)
		return Default()
	}
	return cfg
}
