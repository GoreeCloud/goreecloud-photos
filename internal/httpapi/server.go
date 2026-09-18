package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

type ReadinessProbe interface {
	Probe(context.Context) error
}

type Dependencies struct {
	Storage  storage.OriginalStore
	Database ReadinessProbe
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
	components := make(map[string]componentState, 2)
	ready := true

	if s.deps.Database == nil {
		ready = false
		components["database"] = componentState{Status: "blocked", Detail: "not_configured"}
	} else if err := s.deps.Database.Probe(request.Context()); err != nil {
		ready = false
		components["database"] = componentState{Status: "unavailable", Detail: "probe_failed"}
	} else {
		components["database"] = componentState{Status: "ready"}
	}

	if s.deps.Storage == nil {
		ready = false
		components["original_media_store"] = componentState{Status: "blocked", Detail: "not_configured"}
	} else if err := s.deps.Storage.Probe(request.Context()); err != nil {
		ready = false
		components["original_media_store"] = componentState{Status: "unavailable", Detail: "probe_failed"}
	} else {
		components["original_media_store"] = componentState{Status: "ready"}
	}

	status := http.StatusServiceUnavailable
	payloadStatus := "not_ready"
	if ready {
		status = http.StatusOK
		payloadStatus = "ready"
	}

	writeJSON(writer, status, readinessResponse{
		Status:     payloadStatus,
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
