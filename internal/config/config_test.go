package config

import "testing"

func TestLoadFromEnvDefaultsToLoopback(t *testing.T) {
	t.Setenv("GC_PHOTOS_LISTEN", "")
	t.Setenv("GC_PHOTOS_STORAGE_ROOT", "")

	cfg := LoadFromEnv()
	if cfg.ListenAddress != "127.0.0.1:8780" {
		t.Fatalf("unexpected listen address: %q", cfg.ListenAddress)
	}
	if cfg.StorageRoot != "" {
		t.Fatalf("expected empty storage root, got %q", cfg.StorageRoot)
	}
}

func TestLoadFromEnvUsesExplicitValues(t *testing.T) {
	t.Setenv("GC_PHOTOS_LISTEN", "127.0.0.1:9999")
	t.Setenv("GC_PHOTOS_STORAGE_ROOT", "/tmp/photos-test")

	cfg := LoadFromEnv()
	if cfg.ListenAddress != "127.0.0.1:9999" {
		t.Fatalf("unexpected listen address: %q", cfg.ListenAddress)
	}
	if cfg.StorageRoot != "/tmp/photos-test" {
		t.Fatalf("unexpected storage root: %q", cfg.StorageRoot)
	}
}
