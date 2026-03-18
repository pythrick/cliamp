package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTelemetryEnabled(t *testing.T) {
	if !Default().Telemetry {
		t.Fatalf("Default().Telemetry = false, want true")
	}
}

func TestLoadTelemetryTopLevel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(os.Getenv("HOME"), ".config", "cliamp", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("telemetry = false\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Telemetry {
		t.Fatalf("Telemetry = true, want false")
	}
}

func TestLoadTelemetryDottedKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(os.Getenv("HOME"), ".config", "cliamp", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("telemetry.enabled = false\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Telemetry {
		t.Fatalf("Telemetry = true, want false")
	}
}

func TestLoadTelemetrySection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(os.Getenv("HOME"), ".config", "cliamp", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data := "[telemetry]\nenabled = false\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Telemetry {
		t.Fatalf("Telemetry = true, want false")
	}
}
