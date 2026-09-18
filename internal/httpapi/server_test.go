package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
)

func TestHealth(t *testing.T) {
	server := New("0.1.0-experimental.0", "experimental", Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
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

func TestReadinessFailsClosedWithoutDatabaseAdapter(t *testing.T) {
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := New("0.1.0-experimental.0", "experimental", Dependencies{Storage: store})
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
	if body.Status != "not_ready" {
		t.Fatalf("unexpected readiness status: %q", body.Status)
	}
	if body.Components["database"].Status != "blocked" {
		t.Fatalf("database must remain blocked: %#v", body.Components["database"])
	}
	if body.Components["original_media_store"].Status != "ready" {
		t.Fatalf("storage probe should pass: %#v", body.Components["original_media_store"])
	}
}

func TestReadinessReportsUnconfiguredStorage(t *testing.T) {
	server := New("0.1.0-experimental.0", "experimental", Dependencies{})
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
