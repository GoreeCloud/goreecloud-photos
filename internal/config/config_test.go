package config

import "testing"

func TestLoadFromEnvDefaultsToLoopback(t *testing.T) {
	t.Setenv("GC_PHOTOS_LISTEN", "")
	t.Setenv("GC_PHOTOS_STORAGE_ROOT", "")
	t.Setenv("GC_PHOTOS_DATABASE_URL", "")

	cfg := LoadFromEnv()
	if cfg.ListenAddress != "127.0.0.1:8780" {
		t.Fatalf("unexpected listen address: %q", cfg.ListenAddress)
	}
	if cfg.StorageRoot != "" {
		t.Fatalf("expected empty storage root, got %q", cfg.StorageRoot)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("expected empty database URL, got %q", cfg.DatabaseURL)
	}
}

func TestLoadFromEnvUsesExplicitValues(t *testing.T) {
	t.Setenv("GC_PHOTOS_LISTEN", "127.0.0.1:9999")
	t.Setenv("GC_PHOTOS_STORAGE_ROOT", "/tmp/photos-test")
	t.Setenv("GC_PHOTOS_DATABASE_URL", "postgres://photos:secret@localhost:5432/photos?sslmode=disable")

	cfg := LoadFromEnv()
	if cfg.ListenAddress != "127.0.0.1:9999" {
		t.Fatalf("unexpected listen address: %q", cfg.ListenAddress)
	}
	if cfg.StorageRoot != "/tmp/photos-test" {
		t.Fatalf("unexpected storage root: %q", cfg.StorageRoot)
	}
	if cfg.DatabaseURL != "postgres://photos:secret@localhost:5432/photos?sslmode=disable" {
		t.Fatalf("unexpected database URL: %q", cfg.DatabaseURL)
	}
}
