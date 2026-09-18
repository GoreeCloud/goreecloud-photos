package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/buildinfo"
	"github.com/GoreeCloud/goreecloud-photos/internal/config"
	"github.com/GoreeCloud/goreecloud-photos/internal/httpapi"
	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

func main() {
	cfg := config.LoadFromEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var originalStore storage.OriginalStore
	if cfg.StorageRoot != "" {
		store, err := storage.NewFilesystem(cfg.StorageRoot)
		if err != nil {
			logger.Error("initialize original media store", "error", err)
			os.Exit(1)
		}
		originalStore = store
	}

	handler := httpapi.New(buildinfo.Version, buildinfo.Lifecycle, httpapi.Dependencies{
		Storage: originalStore,
	})

	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown server", "error", err)
		}
	}()

	logger.Info(
		"starting GoreeCloud Photos experimental service",
		"listen", cfg.ListenAddress,
		"storage_configured", originalStore != nil,
		"database_adapter", "not_implemented",
		"version", buildinfo.Version,
		"lifecycle", buildinfo.Lifecycle,
	)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("serve", "error", err)
		os.Exit(1)
	}
}
