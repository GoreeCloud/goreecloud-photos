package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/id"
	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

type ReadinessProbe interface {
	Probe(context.Context) error
}

type Dependencies struct {
	Storage          storage.OriginalStore
	Database         ReadinessProbe
	UploadRepository UploadRepository
	UploadStaging    storage.UploadStagingStore
	UploadAdmission  UploadAdmission
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

type apiErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Retryable bool   `json:"retryable"`
}

type requestIDContextKey struct{}

var fallbackRequestCounter atomic.Uint64

func New(version string, lifecycle string, deps Dependencies) *Server {
	server := &Server{
		version:   version,
		lifecycle: lifecycle,
		deps:      deps,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.health)
	mux.HandleFunc("GET /api/v1/ready", server.ready)
	mux.HandleFunc("POST /api/v1/libraries/{library_id}/uploads", server.createUploadSession)
	mux.HandleFunc("GET /api/v1/uploads/{upload_id}", server.inspectUploadSession)
	mux.HandleFunc("PUT /api/v1/uploads/{upload_id}/parts/{part_number}", server.putUploadPart)
	server.handler = withRequestID(mux)
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
	components := make(map[string]componentState, 3)
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

	if s.deps.UploadAdmission == nil {
		ready = false
		components["upload_admission"] = componentState{Status: "blocked", Detail: "authorization_not_configured"}
	} else if err := s.deps.UploadAdmission.Probe(request.Context()); err != nil {
		ready = false
		components["upload_admission"] = componentState{Status: "unavailable", Detail: "probe_failed"}
	} else {
		components["upload_admission"] = componentState{Status: "ready"}
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

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID, err := id.NewUUIDv7()
		if err != nil {
			requestID = fmt.Sprintf("fallback-%x-%x", time.Now().UnixNano(), fallbackRequestCounter.Add(1))
		}
		writer.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(writer, request.WithContext(context.WithValue(request.Context(), requestIDContextKey{}, requestID)))
	})
}

func requestID(request *http.Request) string {
	value, _ := request.Context().Value(requestIDContextKey{}).(string)
	return value
}

func writeAPIError(writer http.ResponseWriter, request *http.Request, status int, code string, message string, retryable bool) {
	writeJSON(writer, status, apiErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID(request),
		Retryable: retryable,
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
