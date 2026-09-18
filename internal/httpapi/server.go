package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
)

type ReadinessProbe interface {
	Probe(context.Context) error
}

type SubjectResolver interface {
	ResolveSubject(*http.Request) (string, error)
}

type UploadService interface {
	Create(context.Context, string, upload.CreateRequest) (upload.Session, error)
	Get(context.Context, string, string) (upload.Session, error)
	PutPart(context.Context, string, string, int, io.Reader) (upload.Session, error)
	Cancel(context.Context, string, string) (upload.Session, error)
}

type Dependencies struct {
	Storage  storage.OriginalStore
	Database ReadinessProbe
	Uploads  UploadService
	Identity SubjectResolver
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

type requestIDKey struct{}

func New(version string, lifecycle string, deps Dependencies) *Server {
	server := &Server{version: version, lifecycle: lifecycle, deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.health)
	mux.HandleFunc("GET /api/v1/ready", server.ready)
	mux.HandleFunc("POST /api/v1/libraries/{library_id}/uploads", server.createUpload)
	mux.HandleFunc("GET /api/v1/uploads/{upload_id}", server.getUpload)
	mux.HandleFunc("PUT /api/v1/uploads/{upload_id}/parts/{part_number}", server.putUploadPart)
	mux.HandleFunc("DELETE /api/v1/uploads/{upload_id}", server.cancelUpload)
	server.handler = mux
	return server
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	requestID := newRequestID()
	writer.Header().Set("X-Request-ID", requestID)
	ctx := context.WithValue(request.Context(), requestIDKey{}, requestID)
	s.handler.ServeHTTP(writer, request.WithContext(ctx))
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, healthResponse{Status: "healthy", Version: s.version, Lifecycle: s.lifecycle})
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
	writeJSON(writer, status, readinessResponse{Status: payloadStatus, Version: s.version, Lifecycle: s.lifecycle, Components: components})
}

func (s *Server) createUpload(writer http.ResponseWriter, request *http.Request) {
	subject, ok := s.uploadSubject(writer, request)
	if !ok {
		return
	}
	if s.deps.Uploads == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "uploads_unavailable", "Upload service is unavailable.", true)
		return
	}
	var body struct {
		OriginalFilename string `json:"original_filename"`
		MediaType        string `json:"media_type"`
		ExpectedSize     int64  `json:"expected_size"`
		ExpectedSHA256   string `json:"expected_sha256"`
		DeviceID         string `json:"device_id"`
	}
	request.Body = http.MaxBytesReader(writer, request.Body, 64*1024)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil || ensureJSONEOF(decoder) != nil {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_request", "The upload request is invalid.", false)
		return
	}
	session, err := s.deps.Uploads.Create(request.Context(), subject, upload.CreateRequest{
		LibraryID: request.PathValue("library_id"), OriginalFilename: body.OriginalFilename,
		MediaType: body.MediaType, ExpectedSize: body.ExpectedSize,
		ExpectedSHA256: body.ExpectedSHA256, DeviceID: body.DeviceID,
	})
	if err != nil {
		writeUploadError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusCreated, session)
}

func (s *Server) getUpload(writer http.ResponseWriter, request *http.Request) {
	subject, ok := s.uploadSubject(writer, request)
	if !ok {
		return
	}
	if s.deps.Uploads == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "uploads_unavailable", "Upload service is unavailable.", true)
		return
	}
	session, err := s.deps.Uploads.Get(request.Context(), subject, request.PathValue("upload_id"))
	if err != nil {
		writeUploadError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, session)
}

func (s *Server) putUploadPart(writer http.ResponseWriter, request *http.Request) {
	subject, ok := s.uploadSubject(writer, request)
	if !ok {
		return
	}
	if s.deps.Uploads == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "uploads_unavailable", "Upload service is unavailable.", true)
		return
	}
	partNumber, err := strconv.Atoi(request.PathValue("part_number"))
	if err != nil || partNumber < 1 {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_part_number", "The upload part number is invalid.", false)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, upload.DefaultPartSize+1)
	session, err := s.deps.Uploads.PutPart(request.Context(), subject, request.PathValue("upload_id"), partNumber, request.Body)
	if err != nil {
		writeUploadError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, session)
}

func (s *Server) cancelUpload(writer http.ResponseWriter, request *http.Request) {
	subject, ok := s.uploadSubject(writer, request)
	if !ok {
		return
	}
	if s.deps.Uploads == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "uploads_unavailable", "Upload service is unavailable.", true)
		return
	}
	session, err := s.deps.Uploads.Cancel(request.Context(), subject, request.PathValue("upload_id"))
	if err != nil {
		writeUploadError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, session)
}

func (s *Server) uploadSubject(writer http.ResponseWriter, request *http.Request) (string, bool) {
	if s.deps.Identity == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "identity_unavailable", "Identity integration is not configured.", true)
		return "", false
	}
	subject, err := s.deps.Identity.ResolveSubject(request)
	if err != nil || strings.TrimSpace(subject) == "" {
		writeAPIError(writer, request, http.StatusUnauthorized, "authentication_required", "Authentication is required.", false)
		return "", false
	}
	return strings.TrimSpace(subject), true
}

func writeUploadError(writer http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, upload.ErrInvalidRequest), errors.Is(err, upload.ErrPartSize):
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_upload_request", "The upload request is invalid.", false)
	case errors.Is(err, upload.ErrNotFound):
		writeAPIError(writer, request, http.StatusNotFound, "upload_not_found", "The upload session was not found.", false)
	case errors.Is(err, upload.ErrExpired):
		writeAPIError(writer, request, http.StatusGone, "upload_expired", "The upload session has expired.", false)
	case errors.Is(err, upload.ErrPartConflict):
		writeAPIError(writer, request, http.StatusConflict, "upload_part_conflict", "The upload part conflicts with the existing part.", false)
	case errors.Is(err, upload.ErrInvalidState):
		writeAPIError(writer, request, http.StatusConflict, "upload_state_conflict", "The upload session state does not allow this operation.", false)
	case errors.Is(err, upload.ErrUnavailable):
		writeAPIError(writer, request, http.StatusServiceUnavailable, "upload_unavailable", "The upload service is unavailable.", true)
	default:
		writeAPIError(writer, request, http.StatusInternalServerError, "internal_error", "The request could not be completed.", false)
	}
}

func writeAPIError(writer http.ResponseWriter, request *http.Request, status int, code, message string, retryable bool) {
	requestID, _ := request.Context().Value(requestIDKey{}).(string)
	writeJSON(writer, status, apiErrorResponse{Code: code, Message: message, RequestID: requestID, Retryable: retryable})
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == io.EOF {
		return nil
	} else {
		return err
	}
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(value[:])
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
