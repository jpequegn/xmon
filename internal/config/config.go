// internal/config/config.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	X      XConfig      `yaml:"x"`
	APIs   APIsConfig   `yaml:"apis"`
	Fetch  FetchConfig  `yaml:"fetch"`
	Digest DigestConfig `yaml:"digest"`
}

type XConfig struct {
	BearerToken string `yaml:"bearer_token"`
}

type APIsConfig struct {
	LLMProvider string `yaml:"llm_provider"`
	LLMModel    string `yaml:"llm_model"`
}

type FetchConfig struct {
	DefaultInterval int `yaml:"default_interval"`
}

type DigestConfig struct {
	DefaultDays int `yaml:"default_days"`
}

func DefaultConfig() *Config {
	return &Config{
		APIs: APIsConfig{
			LLMProvider: "ollama",
			LLMModel:    "llama3.2",
		},
		Fetch: FetchConfig{
			DefaultInterval: 1440,
		},
		Digest: DigestConfig{
			DefaultDays: 7,
		},
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".xmon")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func DBPath() string {
	return filepath.Join(ConfigDir(), "xmon.db")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0600)
}
