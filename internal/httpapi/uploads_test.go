package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/database"
	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

type testUploadAdmission struct {
	decision UploadAdmissionDecision
	err      error
	probeErr error
}

func (a testUploadAdmission) Probe(context.Context) error {
	return a.probeErr
}

func (a testUploadAdmission) Authorize(*http.Request, UploadAuthorization) (UploadAdmissionDecision, error) {
	return a.decision, a.err
}

type memoryUploadRepository struct {
	mu          sync.Mutex
	sessions    map[string]database.UploadSession
	idempotency map[string]memoryIdempotency
}

type memoryIdempotency struct {
	hash     string
	uploadID string
}

func newMemoryUploadRepository() *memoryUploadRepository {
	return &memoryUploadRepository{
		sessions:    make(map[string]database.UploadSession),
		idempotency: make(map[string]memoryIdempotency),
	}
}

func (r *memoryUploadRepository) CreateUploadSessionIdempotent(_ context.Context, params database.CreateUploadSessionIdempotentParams) (database.UploadSession, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := params.Session.ActorSubjectID + ":" + params.IdempotencyKey
	if previous, ok := r.idempotency[key]; ok {
		if previous.hash != params.RequestSHA256 {
			return database.UploadSession{}, false, database.ErrIdempotencyConflict
		}
		return r.sessions[previous.uploadID], true, nil
	}
	session := database.UploadSession{
		UploadID:         params.Session.UploadID,
		ActorSubjectID:   params.Session.ActorSubjectID,
		LibraryID:        params.Session.LibraryID,
		OriginalFilename: params.Session.OriginalFilename,
		MediaType:        params.Session.MediaType,
		DeviceID:         params.Session.DeviceID,
		CaptureTime:      params.Session.CaptureTime,
		CaptureTimeZone:  params.Session.CaptureTimeZone,
		ExpectedSize:     params.Session.ExpectedSize,
		ExpectedSHA256:   params.Session.ExpectedSHA256,
		PartSize:         params.Session.PartSize,
		State:            "open",
		ExpiresAt:        params.Session.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Parts:            []database.UploadPart{},
	}
	r.sessions[session.UploadID] = session
	r.idempotency[key] = memoryIdempotency{hash: params.RequestSHA256, uploadID: session.UploadID}
	return session, false, nil
}

func (r *memoryUploadRepository) GetUploadSession(_ context.Context, uploadID string) (database.UploadSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[uploadID]
	if !ok {
		return database.UploadSession{}, database.ErrUploadSessionNotFound
	}
	return session, nil
}

func (r *memoryUploadRepository) PrepareUploadPart(_ context.Context, uploadID string, partNumber int32) (database.UploadSession, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[uploadID]
	if !ok {
		return database.UploadSession{}, 0, database.ErrUploadSessionNotFound
	}
	if !session.ExpiresAt.After(time.Now()) {
		session.State = "expired"
		r.sessions[uploadID] = session
		return database.UploadSession{}, 0, database.ErrUploadSessionExpired
	}
	offset := int64(partNumber-1) * session.PartSize
	if partNumber < 1 || offset < 0 || offset >= session.ExpectedSize {
		return database.UploadSession{}, 0, database.ErrInvalidUploadPart
	}
	expected := session.PartSize
	if remaining := session.ExpectedSize - offset; remaining < expected {
		expected = remaining
	}
	return session, expected, nil
}

func (r *memoryUploadRepository) RecordUploadPart(_ context.Context, params database.RecordUploadPartParams) (database.UploadSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[params.UploadID]
	if !ok {
		return database.UploadSession{}, database.ErrUploadSessionNotFound
	}
	for _, part := range session.Parts {
		if part.PartNumber == params.PartNumber {
			if part.ByteSize != params.ByteSize || part.SHA256 != params.SHA256 {
				return database.UploadSession{}, database.ErrUploadPartConflict
			}
			return session, nil
		}
	}
	session.Parts = append(session.Parts, database.UploadPart{
		PartNumber: params.PartNumber,
		ByteSize:   params.ByteSize,
		SHA256:     params.SHA256,
		CreatedAt:  time.Now(),
	})
	session.ReceivedBytes += params.ByteSize
	session.State = "receiving"
	session.UpdatedAt = time.Now()
	r.sessions[params.UploadID] = session
	return session, nil
}

func TestUploadRoutesFailClosedWithoutAdmission(t *testing.T) {
	server := New("test", "experimental", Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/libraries/01997d3a-0000-7000-8000-000000000301/uploads", bytes.NewBufferString(`{"original_filename":"a.jpg","expected_size":4}`))
	request.Header.Set("Idempotency-Key", "test-key")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
	var body apiErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "capability_unavailable" || body.RequestID == "" {
		t.Fatalf("unexpected error response: %#v", body)
	}
}

func TestCreateInspectAndUploadParts(t *testing.T) {
	repository := newMemoryUploadRepository()
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	admission := testUploadAdmission{
		decision: UploadAdmissionDecision{
			ActorSubjectID: "subject:test",
			Policy: UploadPolicy{
				MaxAssetBytes: 100,
				PartSize:      4,
				SessionTTL:    time.Hour,
			},
		},
	}
	server := New("test", "experimental", Dependencies{
		UploadRepository: repository,
		UploadStaging:    store,
		UploadAdmission:  admission,
	})

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/libraries/01997d3a-0000-7000-8000-000000000302/uploads",
		bytes.NewBufferString(`{"original_filename":"IMG_1.jpg","media_type":"image/jpeg","expected_size":6}`),
	)
	createRequest.Header.Set("Idempotency-Key", "create-1")
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}
	var created uploadSessionResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.UploadID == "" || created.PartSize != 4 || created.ExpectedSize != 6 {
		t.Fatalf("unexpected created session: %#v", created)
	}

	replayRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/libraries/01997d3a-0000-7000-8000-000000000302/uploads",
		bytes.NewBufferString(`{"original_filename":"IMG_1.jpg","media_type":"image/jpeg","expected_size":6}`),
	)
	replayRequest.Header.Set("Idempotency-Key", "create-1")
	replayResponse := httptest.NewRecorder()
	server.ServeHTTP(replayResponse, replayRequest)
	if replayResponse.Code != http.StatusCreated || replayResponse.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("expected idempotent replay, status=%d headers=%v body=%s", replayResponse.Code, replayResponse.Header(), replayResponse.Body.String())
	}

	part1Request := httptest.NewRequest(http.MethodPut, "/api/v1/uploads/"+created.UploadID+"/parts/1", bytes.NewBufferString("abcd"))
	part1Response := httptest.NewRecorder()
	server.ServeHTTP(part1Response, part1Request)
	if part1Response.Code != http.StatusOK {
		t.Fatalf("part1 status=%d body=%s", part1Response.Code, part1Response.Body.String())
	}

	part2Request := httptest.NewRequest(http.MethodPut, "/api/v1/uploads/"+created.UploadID+"/parts/2", bytes.NewBufferString("ef"))
	part2Response := httptest.NewRecorder()
	server.ServeHTTP(part2Response, part2Request)
	if part2Response.Code != http.StatusOK {
		t.Fatalf("part2 status=%d body=%s", part2Response.Code, part2Response.Body.String())
	}

	retryRequest := httptest.NewRequest(http.MethodPut, "/api/v1/uploads/"+created.UploadID+"/parts/2", bytes.NewBufferString("ef"))
	retryResponse := httptest.NewRecorder()
	server.ServeHTTP(retryResponse, retryRequest)
	if retryResponse.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%s", retryResponse.Code, retryResponse.Body.String())
	}

	conflictRequest := httptest.NewRequest(http.MethodPut, "/api/v1/uploads/"+created.UploadID+"/parts/2", bytes.NewBufferString("zz"))
	conflictResponse := httptest.NewRecorder()
	server.ServeHTTP(conflictResponse, conflictRequest)
	if conflictResponse.Code != http.StatusConflict {
		t.Fatalf("conflict status=%d body=%s", conflictResponse.Code, conflictResponse.Body.String())
	}

	inspectRequest := httptest.NewRequest(http.MethodGet, "/api/v1/uploads/"+created.UploadID, nil)
	inspectResponse := httptest.NewRecorder()
	server.ServeHTTP(inspectResponse, inspectRequest)
	if inspectResponse.Code != http.StatusOK {
		t.Fatalf("inspect status=%d body=%s", inspectResponse.Code, inspectResponse.Body.String())
	}
	var inspected uploadSessionResponse
	if err := json.Unmarshal(inspectResponse.Body.Bytes(), &inspected); err != nil {
		t.Fatal(err)
	}
	if inspected.ReceivedBytes != 6 || len(inspected.Parts) != 2 {
		t.Fatalf("unexpected inspected session: %#v", inspected)
	}
}

func TestUploadPartRejectsShortBodyWithoutCommittingStaging(t *testing.T) {
	repository := newMemoryUploadRepository()
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	admission := testUploadAdmission{
		decision: UploadAdmissionDecision{
			ActorSubjectID: "subject:test",
			Policy: UploadPolicy{MaxAssetBytes: 100, PartSize: 4, SessionTTL: time.Hour},
		},
	}
	server := New("test", "experimental", Dependencies{
		UploadRepository: repository,
		UploadStaging:    store,
		UploadAdmission:  admission,
	})

	session := database.UploadSession{
		UploadID:       "01997d3a-0000-7000-8000-000000000303",
		ActorSubjectID: "subject:test",
		LibraryID:      "01997d3a-0000-7000-8000-000000000304",
		ExpectedSize:   4,
		PartSize:       4,
		State:          "open",
		ExpiresAt:      time.Now().Add(time.Hour),
		Parts:          []database.UploadPart{},
	}
	repository.sessions[session.UploadID] = session

	request := httptest.NewRequest(http.MethodPut, "/api/v1/uploads/"+session.UploadID+"/parts/1", bytes.NewBufferString("abc"))
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if _, err := store.OpenUploadPart(context.Background(), session.UploadID, 1); !errors.Is(err, storage.ErrUploadPartNotFound) {
		t.Fatalf("short part must not be staged, got %v", err)
	}
}
