package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

type readinessProbeFunc func(context.Context) error

func (f readinessProbeFunc) Probe(ctx context.Context) error {
	return f(ctx)
}

func TestHealth(t *testing.T) {
	server := New("0.1.0-experimental.2", "experimental", Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("health response is missing X-Request-ID")
	}

	var body healthResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "healthy" || body.Lifecycle != "experimental" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("unexpected Cache-Control: %q", cacheControl)
	}
}

func TestReadinessSucceedsWithHealthyRequiredDependencies(t *testing.T) {
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := New("test", "experimental", Dependencies{
		Storage: store,
		Database: readinessProbeFunc(func(context.Context) error {
			return nil
		}),
		UploadAdmission: testUploadAdmission{
			decision: UploadAdmissionDecision{
				ActorSubjectID: "subject:test",
				Policy: UploadPolicy{
					MaxAssetBytes: 100,
					PartSize:      4,
					SessionTTL:    time.Hour,
				},
			},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", response.Code, response.Body.String())
	}

	var body readinessResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ready" {
		t.Fatalf("unexpected readiness status: %q", body.Status)
	}
	if body.Components["database"].Status != "ready" ||
		body.Components["original_media_store"].Status != "ready" ||
		body.Components["upload_admission"].Status != "ready" {
		t.Fatalf("unexpected components: %#v", body.Components)
	}
}

func TestReadinessFailsClosedWithoutUploadAdmission(t *testing.T) {
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := New("test", "experimental", Dependencies{
		Storage: store,
		Database: readinessProbeFunc(func(context.Context) error {
			return nil
		}),
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	var body readinessResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Components["upload_admission"].Detail != "authorization_not_configured" {
		t.Fatalf("unexpected upload admission state: %#v", body.Components["upload_admission"])
	}
}

func TestReadinessFailsClosedWithoutDatabaseConfiguration(t *testing.T) {
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := New("test", "experimental", Dependencies{
		Storage:         store,
		UploadAdmission: testUploadAdmission{decision: UploadAdmissionDecision{ActorSubjectID: "subject:test"}},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	var body readinessResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "not_ready" || body.Components["database"].Detail != "not_configured" {
		t.Fatalf("unexpected readiness state: %#v", body)
	}
}

func TestReadinessFailsClosedWhenDatabaseProbeFails(t *testing.T) {
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := New("test", "experimental", Dependencies{
		Storage: store,
		Database: readinessProbeFunc(func(context.Context) error {
			return errors.New("database unavailable")
		}),
		UploadAdmission: testUploadAdmission{decision: UploadAdmissionDecision{ActorSubjectID: "subject:test"}},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	var body readinessResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Components["database"].Status != "unavailable" || body.Components["database"].Detail != "probe_failed" {
		t.Fatalf("unexpected database state: %#v", body.Components["database"])
	}
}

func TestReadinessReportsUnconfiguredStorage(t *testing.T) {
	server := New("test", "experimental", Dependencies{
		Database: readinessProbeFunc(func(context.Context) error {
			return nil
		}),
		UploadAdmission: testUploadAdmission{decision: UploadAdmissionDecision{ActorSubjectID: "subject:test"}},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	var body readinessResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Components["original_media_store"].Detail != "not_configured" {
		t.Fatalf("unexpected storage state: %#v", body.Components["original_media_store"])
	}
}
