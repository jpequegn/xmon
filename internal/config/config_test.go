// internal/config/config_test.go
package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Digest.DefaultDays != 7 {
		t.Errorf("expected default days 7, got %d", cfg.Digest.DefaultDays)
	}
	if cfg.APIs.LLMModel != "llama3.2" {
		t.Errorf("expected llama3.2, got %s", cfg.APIs.LLMModel)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("config dir should not be empty")
	}
}
