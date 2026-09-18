package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/database"
	"github.com/GoreeCloud/goreecloud-photos/internal/id"
	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

var (
	ErrUploadAdmissionDenied              = errors.New("upload admission denied")
	ErrUploadAdmissionEvidenceUnavailable = errors.New("upload admission evidence unavailable")
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type UploadAuthorization struct {
	Action       string
	LibraryID    string
	UploadID     string
	ExpectedSize int64
	MediaType    string
}

type UploadPolicy struct {
	MaxAssetBytes int64
	PartSize      int64
	SessionTTL    time.Duration
}

type UploadAdmissionDecision struct {
	ActorSubjectID string
	Policy         UploadPolicy
}

type UploadAdmission interface {
	Probe(context.Context) error
	Authorize(*http.Request, UploadAuthorization) (UploadAdmissionDecision, error)
}

type UploadRepository interface {
	CreateUploadSessionIdempotent(context.Context, database.CreateUploadSessionIdempotentParams) (database.UploadSession, bool, error)
	GetUploadSession(context.Context, string) (database.UploadSession, error)
	PrepareUploadPart(context.Context, string, int32) (database.UploadSession, int64, error)
	RecordUploadPart(context.Context, database.RecordUploadPartParams) (database.UploadSession, error)
}

type createUploadRequest struct {
	OriginalFilename string     `json:"original_filename"`
	MediaType        string     `json:"media_type,omitempty"`
	DeviceID         string     `json:"device_id,omitempty"`
	CaptureTime      *time.Time `json:"capture_time,omitempty"`
	CaptureTimeZone  string     `json:"capture_time_zone,omitempty"`
	ExpectedSize     int64      `json:"expected_size"`
	ExpectedSHA256   *string    `json:"expected_sha256,omitempty"`
}

type uploadPartResponse struct {
	PartNumber int32  `json:"part_number"`
	ByteSize   int64  `json:"byte_size"`
	SHA256     string `json:"sha256"`
}

type uploadSessionResponse struct {
	UploadID        string               `json:"upload_id"`
	LibraryID       string               `json:"library_id"`
	ExpectedSize    int64                `json:"expected_size"`
	ExpectedSHA256  *string              `json:"expected_sha256,omitempty"`
	ReceivedBytes   int64                `json:"received_bytes"`
	PartSize        int64                `json:"part_size"`
	State           string               `json:"state"`
	ExpiresAt       time.Time            `json:"expires_at"`
	TransferMethod  string               `json:"transfer_method"`
	Parts           []uploadPartResponse `json:"parts"`
}

type putUploadPartResponse struct {
	Part    uploadPartResponse    `json:"part"`
	Session uploadSessionResponse `json:"session"`
}

func (s *Server) createUploadSession(writer http.ResponseWriter, request *http.Request) {
	if s.deps.UploadAdmission == nil || s.deps.UploadRepository == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "capability_unavailable", "Upload admission is not configured.", true)
		return
	}

	libraryID := request.PathValue("library_id")
	if !uuidPattern.MatchString(libraryID) {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_library_id", "The library identifier is invalid.", false)
		return
	}

	var payload createUploadRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 64<<10)
	if err := decodeSingleJSON(request.Body, &payload); err != nil {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_request", "The upload request is not valid JSON.", false)
		return
	}
	if err := normalizeAndValidateCreateUploadRequest(&payload); err != nil {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_request", err.Error(), false)
		return
	}

	decision, ok := s.authorizeUpload(writer, request, UploadAuthorization{
		Action:       "create",
		LibraryID:    libraryID,
		ExpectedSize: payload.ExpectedSize,
		MediaType:    payload.MediaType,
	})
	if !ok {
		return
	}
	if decision.Policy.MaxAssetBytes <= 0 || decision.Policy.PartSize <= 0 || decision.Policy.SessionTTL <= 0 {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "admission_evidence_unavailable", "Upload policy evidence is unavailable.", true)
		return
	}
	if payload.ExpectedSize > decision.Policy.MaxAssetBytes {
		writeAPIError(writer, request, http.StatusRequestEntityTooLarge, "asset_too_large", "The asset exceeds the authorized upload limit.", false)
		return
	}

	idempotencyKey := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" || len(idempotencyKey) > 128 {
		writeAPIError(writer, request, http.StatusBadRequest, "idempotency_key_required", "A bounded Idempotency-Key is required.", false)
		return
	}

	uploadID, err := id.NewUUIDv7()
	if err != nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "identifier_unavailable", "A server upload identifier could not be created.", true)
		return
	}
	requestSHA256, err := createUploadRequestFingerprint(libraryID, payload)
	if err != nil {
		writeAPIError(writer, request, http.StatusInternalServerError, "internal_error", "The upload request could not be prepared.", true)
		return
	}

	session, replayed, err := s.deps.UploadRepository.CreateUploadSessionIdempotent(request.Context(), database.CreateUploadSessionIdempotentParams{
		IdempotencyKey: idempotencyKey,
		RequestSHA256:  requestSHA256,
		Session: database.CreateUploadSessionParams{
			UploadID:         uploadID,
			ActorSubjectID:   decision.ActorSubjectID,
			LibraryID:        libraryID,
			OriginalFilename: payload.OriginalFilename,
			MediaType:        payload.MediaType,
			DeviceID:         payload.DeviceID,
			CaptureTime:      payload.CaptureTime,
			CaptureTimeZone:  payload.CaptureTimeZone,
			ExpectedSize:     payload.ExpectedSize,
			ExpectedSHA256:   payload.ExpectedSHA256,
			PartSize:         decision.Policy.PartSize,
			ExpiresAt:        time.Now().Add(decision.Policy.SessionTTL),
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, database.ErrIdempotencyConflict):
			writeAPIError(writer, request, http.StatusConflict, "idempotency_conflict", "The Idempotency-Key was already used for a different request.", false)
		case errors.Is(err, database.ErrInvalidUploadSession), errors.Is(err, database.ErrInvalidIdempotency):
			writeAPIError(writer, request, http.StatusBadRequest, "invalid_request", "The upload session could not be created from the supplied request.", false)
		default:
			writeAPIError(writer, request, http.StatusInternalServerError, "upload_session_create_failed", "The upload session could not be created.", true)
		}
		return
	}

	if replayed {
		writer.Header().Set("Idempotent-Replayed", "true")
	}
	writeJSON(writer, http.StatusCreated, makeUploadSessionResponse(session))
}

func (s *Server) inspectUploadSession(writer http.ResponseWriter, request *http.Request) {
	if s.deps.UploadAdmission == nil || s.deps.UploadRepository == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "capability_unavailable", "Upload admission is not configured.", true)
		return
	}
	uploadID := request.PathValue("upload_id")
	if !uuidPattern.MatchString(uploadID) {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_upload_id", "The upload identifier is invalid.", false)
		return
	}

	decision, ok := s.authorizeUpload(writer, request, UploadAuthorization{Action: "inspect", UploadID: uploadID})
	if !ok {
		return
	}
	session, err := s.deps.UploadRepository.GetUploadSession(request.Context(), uploadID)
	if err != nil {
		writeUploadRepositoryError(writer, request, err)
		return
	}
	if session.ActorSubjectID != decision.ActorSubjectID {
		writeAPIError(writer, request, http.StatusNotFound, "upload_not_found", "The upload session was not found.", false)
		return
	}
	writeJSON(writer, http.StatusOK, makeUploadSessionResponse(session))
}

func (s *Server) putUploadPart(writer http.ResponseWriter, request *http.Request) {
	if s.deps.UploadAdmission == nil || s.deps.UploadRepository == nil || s.deps.UploadStaging == nil {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "capability_unavailable", "Upload staging is not configured.", true)
		return
	}

	uploadID := request.PathValue("upload_id")
	if !uuidPattern.MatchString(uploadID) {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_upload_id", "The upload identifier is invalid.", false)
		return
	}
	partNumber64, err := strconv.ParseInt(request.PathValue("part_number"), 10, 32)
	if err != nil || partNumber64 < 1 {
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_part_number", "The upload part number is invalid.", false)
		return
	}
	partNumber := int32(partNumber64)

	decision, ok := s.authorizeUpload(writer, request, UploadAuthorization{Action: "upload_part", UploadID: uploadID})
	if !ok {
		return
	}
	current, err := s.deps.UploadRepository.GetUploadSession(request.Context(), uploadID)
	if err != nil {
		writeUploadRepositoryError(writer, request, err)
		return
	}
	if current.ActorSubjectID != decision.ActorSubjectID {
		writeAPIError(writer, request, http.StatusNotFound, "upload_not_found", "The upload session was not found.", false)
		return
	}

	prepared, expectedSize, err := s.deps.UploadRepository.PrepareUploadPart(request.Context(), uploadID, partNumber)
	if err != nil {
		writeUploadRepositoryError(writer, request, err)
		return
	}
	if prepared.ActorSubjectID != decision.ActorSubjectID {
		writeAPIError(writer, request, http.StatusNotFound, "upload_not_found", "The upload session was not found.", false)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, expectedSize+1)
	staged, err := s.deps.UploadStaging.PutUploadPart(request.Context(), uploadID, partNumber, request.Body, expectedSize)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrSizeMismatch):
			writeAPIError(writer, request, http.StatusBadRequest, "invalid_part_size", "The request body does not match the negotiated part size.", false)
		case errors.Is(err, storage.ErrUploadPartConflict):
			writeAPIError(writer, request, http.StatusConflict, "upload_part_conflict", "Different bytes are already staged for this upload part.", false)
		default:
			writeAPIError(writer, request, http.StatusInternalServerError, "upload_part_stage_failed", "The upload part could not be staged.", true)
		}
		return
	}

	session, err := s.deps.UploadRepository.RecordUploadPart(request.Context(), database.RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: partNumber,
		ByteSize:   staged.Size,
		SHA256:     staged.SHA256,
	})
	if err != nil {
		if staged.Created {
			_ = s.deps.UploadStaging.DeleteUploadPart(context.Background(), uploadID, partNumber, staged.SHA256)
		}
		writeUploadRepositoryError(writer, request, err)
		return
	}

	writeJSON(writer, http.StatusOK, putUploadPartResponse{
		Part: uploadPartResponse{
			PartNumber: partNumber,
			ByteSize:   staged.Size,
			SHA256:     staged.SHA256,
		},
		Session: makeUploadSessionResponse(session),
	})
}

func (s *Server) authorizeUpload(writer http.ResponseWriter, request *http.Request, authorization UploadAuthorization) (UploadAdmissionDecision, bool) {
	decision, err := s.deps.UploadAdmission.Authorize(request, authorization)
	if err != nil {
		switch {
		case errors.Is(err, ErrUploadAdmissionDenied):
			writeAPIError(writer, request, http.StatusForbidden, "upload_not_authorized", "The upload operation is not authorized.", false)
		case errors.Is(err, ErrUploadAdmissionEvidenceUnavailable):
			writeAPIError(writer, request, http.StatusServiceUnavailable, "admission_evidence_unavailable", "Required upload authorization evidence is unavailable.", true)
		default:
			writeAPIError(writer, request, http.StatusServiceUnavailable, "admission_evidence_unavailable", "Required upload authorization evidence is unavailable.", true)
		}
		return UploadAdmissionDecision{}, false
	}
	if strings.TrimSpace(decision.ActorSubjectID) == "" {
		writeAPIError(writer, request, http.StatusServiceUnavailable, "admission_evidence_unavailable", "Required upload actor evidence is unavailable.", true)
		return UploadAdmissionDecision{}, false
	}
	return decision, true
}

func normalizeAndValidateCreateUploadRequest(payload *createUploadRequest) error {
	payload.OriginalFilename = strings.TrimSpace(payload.OriginalFilename)
	payload.MediaType = strings.TrimSpace(payload.MediaType)
	payload.DeviceID = strings.TrimSpace(payload.DeviceID)
	payload.CaptureTimeZone = strings.TrimSpace(payload.CaptureTimeZone)

	if payload.OriginalFilename == "" || len(payload.OriginalFilename) > 1024 ||
		strings.ContainsAny(payload.OriginalFilename, "/\\ 
") {
		return errors.New("The original filename is invalid.")
	}
	if payload.ExpectedSize <= 0 {
		return errors.New("The expected asset size must be greater than zero.")
	}
	if len(payload.MediaType) > 255 || strings.ContainsAny(payload.MediaType, " 
") {
		return errors.New("The media type is invalid.")
	}
	if len(payload.DeviceID) > 256 || strings.ContainsAny(payload.DeviceID, " 
") {
		return errors.New("The device identifier is invalid.")
	}
	if len(payload.CaptureTimeZone) > 128 || strings.ContainsAny(payload.CaptureTimeZone, " 
") {
		return errors.New("The capture time zone is invalid.")
	}
	if payload.ExpectedSHA256 != nil {
		value := strings.ToLower(strings.TrimSpace(*payload.ExpectedSHA256))
		if len(value) != 64 {
			return errors.New("The expected SHA-256 is invalid.")
		}
		for _, character := range value {
			if !strings.ContainsRune("0123456789abcdef", character) {
				return errors.New("The expected SHA-256 is invalid.")
			}
		}
		payload.ExpectedSHA256 = &value
	}
	return nil
}

func createUploadRequestFingerprint(libraryID string, payload createUploadRequest) (string, error) {
	canonical := struct {
		LibraryID string              `json:"library_id"`
		Payload   createUploadRequest `json:"payload"`
	}{
		LibraryID: libraryID,
		Payload:   payload,
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func makeUploadSessionResponse(session database.UploadSession) uploadSessionResponse {
	parts := make([]uploadPartResponse, 0, len(session.Parts))
	for _, part := range session.Parts {
		parts = append(parts, uploadPartResponse{
			PartNumber: part.PartNumber,
			ByteSize:   part.ByteSize,
			SHA256:     part.SHA256,
		})
	}
	state := session.State
	if (state == "open" || state == "receiving") && !session.ExpiresAt.After(time.Now()) {
		state = "expired"
	}
	return uploadSessionResponse{
		UploadID:       session.UploadID,
		LibraryID:      session.LibraryID,
		ExpectedSize:   session.ExpectedSize,
		ExpectedSHA256: session.ExpectedSHA256,
		ReceivedBytes:  session.ReceivedBytes,
		PartSize:       session.PartSize,
		State:          state,
		ExpiresAt:      session.ExpiresAt,
		TransferMethod: "http-put-parts-v1",
		Parts:          parts,
	}
}

func writeUploadRepositoryError(writer http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, database.ErrUploadSessionNotFound):
		writeAPIError(writer, request, http.StatusNotFound, "upload_not_found", "The upload session was not found.", false)
	case errors.Is(err, database.ErrUploadSessionExpired):
		writeAPIError(writer, request, http.StatusGone, "upload_expired", "The upload session has expired.", false)
	case errors.Is(err, database.ErrUploadSessionClosed):
		writeAPIError(writer, request, http.StatusConflict, "upload_not_writable", "The upload session is not writable.", false)
	case errors.Is(err, database.ErrUploadPartConflict):
		writeAPIError(writer, request, http.StatusConflict, "upload_part_conflict", "The upload part conflicts with committed receipt evidence.", false)
	case errors.Is(err, database.ErrInvalidUploadPart), errors.Is(err, database.ErrInvalidUploadSession):
		writeAPIError(writer, request, http.StatusBadRequest, "invalid_upload_request", "The upload request is invalid.", false)
	default:
		writeAPIError(writer, request, http.StatusInternalServerError, "upload_operation_failed", "The upload operation could not be completed.", true)
	}
}

func decodeSingleJSON(reader io.Reader, destination any) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values are not allowed")
	}
	return nil
}
