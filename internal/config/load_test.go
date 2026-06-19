//go:build unit

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jedi-knights/api-gateway/internal/config"
)

func setenv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

// writeConfigFile creates a temp gateway.yaml under t.TempDir, writes yaml
// into it, and returns the absolute path. Cleanup is handled by t.TempDir.
func writeConfigFile(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gateway.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	return path
}

func TestLoad_UsesDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
}

func TestLoad_EnvVarsOverrideDefaults(t *testing.T) {
	setenv(t, "GATEWAY_SERVER_HOST", "127.0.0.1")
	setenv(t, "GATEWAY_SERVER_PORT", "9090")
	setenv(t, "GATEWAY_LOG_LEVEL", "debug")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
}

func TestLoad_InvalidPortReturnsError(t *testing.T) {
	setenv(t, "GATEWAY_SERVER_PORT", "99999")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for out-of-range port, got nil")
	}
}

func TestLoad_ConfigFileOverridesDefaults(t *testing.T) {
	yaml := `
server:
  host: "10.0.0.1"
  port: 7070
log:
  level: "warn"
`
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, yaml))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "10.0.0.1")
	}
	if cfg.Server.Port != 7070 {
		t.Errorf("Server.Port = %d, want 7070", cfg.Server.Port)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "warn")
	}
}

func TestLoad_Validate_DuplicateRouteNameReturnsError(t *testing.T) {
	yaml := `
routes:
  - name: "svc"
    upstream:
      url: "http://svc:8080"
  - name: "svc"
    upstream:
      url: "http://svc2:8080"
`
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, yaml))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for duplicate route name, got nil")
	}
}

func TestLoad_Validate_MissingRouteNameReturnsError(t *testing.T) {
	yaml := `
routes:
  - upstream:
      url: "http://svc:8080"
`
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, yaml))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing route name, got nil")
	}
}

func TestLoad_Validate_MissingUpstreamURLReturnsError(t *testing.T) {
	yaml := `
routes:
  - name: "svc"
    upstream:
      url: ""
`
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, yaml))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for empty upstream URL, got nil")
	}
}

func TestLoad_InvalidConfigFileContentReturnsError(t *testing.T) {
	// Deliberately malformed YAML — viper returns a parse error, not ConfigFileNotFoundError.
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, "server:\n  port: [invalid yaml here\n"))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for malformed config file, got nil")
	}
}

func TestLoad_Validate_ValidConfigWithRoutesSucceeds(t *testing.T) {
	yaml := `
routes:
  - name: "identity"
    match:
      path_prefix: "/api/identity"
      methods: ["GET", "POST"]
    upstream:
      url: "http://identity:8080"
      strip_prefix: "/api/identity"
`
	setenv(t, "GATEWAY_CONFIG_FILE", writeConfigFile(t, yaml))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
	if cfg.Routes[0].Name != "identity" {
		t.Errorf("route name = %q, want %q", cfg.Routes[0].Name, "identity")
	}
}
