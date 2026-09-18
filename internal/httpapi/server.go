package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

type Dependencies struct {
	Storage storage.OriginalStore
}

type Server struct {
	version   string
	lifecycle string
	deps      Dependencies
	handler   http.Handler
}

type healthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Lifecycle string `json:"lifecycle"`
}

type componentState struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type readinessResponse struct {
	Status     string                    `json:"status"`
	Version    string                    `json:"version"`
	Lifecycle  string                    `json:"lifecycle"`
	Components map[string]componentState `json:"components"`
}

func New(version string, lifecycle string, deps Dependencies) *Server {
	server := &Server{
		version:   version,
		lifecycle: lifecycle,
		deps:      deps,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.health)
	mux.HandleFunc("GET /api/v1/ready", server.ready)
	server.handler = mux
	return server
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.handler.ServeHTTP(writer, request)
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, healthResponse{
		Status:    "healthy",
		Version:   s.version,
		Lifecycle: s.lifecycle,
	})
}

func (s *Server) ready(writer http.ResponseWriter, request *http.Request) {
	components := map[string]componentState{
		"database": {
			Status: "blocked",
			Detail: "adapter_not_implemented",
		},
	}

	if s.deps.Storage == nil {
		components["original_media_store"] = componentState{
			Status: "blocked",
			Detail: "not_configured",
		}
	} else if err := s.deps.Storage.Probe(request.Context()); err != nil {
		components["original_media_store"] = componentState{
			Status: "unavailable",
			Detail: "probe_failed",
		}
	} else {
		components["original_media_store"] = componentState{Status: "ready"}
	}

	writeJSON(writer, http.StatusServiceUnavailable, readinessResponse{
		Status:     "not_ready",
		Version:    s.version,
		Lifecycle:  s.lifecycle,
		Components: components,
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func probeWithContext(ctx context.Context, probe func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return probe(ctx)
}
