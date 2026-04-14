package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Port != 3001 {
		t.Errorf("default port = %d, want 3001", cfg.Server.Port)
	}
	if cfg.Minima.DataPort != 9004 {
		t.Errorf("default data port = %d, want 9004", cfg.Minima.DataPort)
	}
	if cfg.Minima.MiniDAPPPort != 9005 {
		t.Errorf("default minidapp port = %d, want 9005", cfg.Minima.MiniDAPPPort)
	}
	if cfg.Auth.Mode != "tiered" {
		t.Errorf("default auth mode = %q, want tiered", cfg.Auth.Mode)
	}
	if !cfg.Audit.Enabled {
		t.Error("audit should be enabled by default")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("SERVER_PORT", "4001")
	os.Setenv("MINIMA_DATA_PORT", "8000")
	os.Setenv("AUTH_MODE", "false")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Port != 4001 {
		t.Errorf("port = %d, want 4001", cfg.Server.Port)
	}
	if cfg.Minima.DataPort != 8000 {
		t.Errorf("data port = %d, want 8000", cfg.Minima.DataPort)
	}
	if cfg.Auth.Mode != "false" {
		t.Errorf("auth mode = %q, want false", cfg.Auth.Mode)
	}
}

func TestDataURL(t *testing.T) {
	cfg := &MinimaConfig{DataHost: "myhost", DataPort: 9999}
	if cfg.DataURL() != "http://myhost:9999" {
		t.Errorf("DataURL = %q, want http://myhost:9999", cfg.DataURL())
	}
}

func TestMiniDAPPURL(t *testing.T) {
	cfg := &MinimaConfig{MiniDAPPHost: "mds-host", MiniDAPPPort: 7777}
	if cfg.MiniDAPPURL() != "http://mds-host:7777" {
		t.Errorf("MiniDAPPURL = %q, want http://mds-host:7777", cfg.MiniDAPPURL())
	}
}

func TestInvalidPort(t *testing.T) {
	os.Setenv("MINIMA_DATA_PORT", "0")
	defer os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid port")
	}
}
