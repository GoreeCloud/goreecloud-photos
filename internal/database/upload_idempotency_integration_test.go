package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestIdempotentUploadSessionCreationPersistsMetadata(t *testing.T) {
	connectionString := os.Getenv("GC_PHOTOS_TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("GC_PHOTOS_TEST_DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := NewPostgreSQL(ctx, connectionString)
	if err != nil {
		t.Fatalf("initialize PostgreSQL adapter: %v", err)
	}
	defer database.Close()
	if err := resetAndApplyMigrations(ctx, database); err != nil {
		t.Fatalf("prepare database: %v", err)
	}

	const (
		libraryID = "01997d3a-0000-7000-8000-000000000201"
		uploadID1 = "01997d3a-0000-7000-8000-000000000202"
		uploadID2 = "01997d3a-0000-7000-8000-000000000203"
		actorID   = "subject:idempotency-test"
		key       = "create-upload-1"
		hash      = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	if _, err := database.pool.Exec(ctx, `
		INSERT INTO libraries (library_id, library_type, owner_subject_id)
		VALUES ($1, 'personal', $2)
	`, libraryID, actorID); err != nil {
		t.Fatalf("create test library: %v", err)
	}

	captureTime := time.Date(2026, 9, 18, 7, 0, 0, 0, time.UTC)
	params := CreateUploadSessionParams{
		UploadID:         uploadID1,
		ActorSubjectID:   actorID,
		LibraryID:        libraryID,
		OriginalFilename: "IMG_0001.jpg",
		MediaType:        "image/jpeg",
		DeviceID:         "device:test",
		CaptureTime:      &captureTime,
		CaptureTimeZone:  "America/New_York",
		ExpectedSize:     10,
		PartSize:         4,
		ExpiresAt:        time.Now().Add(time.Hour),
	}
	first, replayed, err := database.CreateUploadSessionIdempotent(ctx, CreateUploadSessionIdempotentParams{
		IdempotencyKey: key,
		RequestSHA256:  hash,
		Session:        params,
	})
	if err != nil {
		t.Fatal(err)
	}
	if replayed {
		t.Fatal("first request must not be a replay")
	}
	if first.OriginalFilename != params.OriginalFilename || first.MediaType != params.MediaType || first.DeviceID != params.DeviceID || first.CaptureTimeZone != params.CaptureTimeZone {
		t.Fatalf("metadata was not preserved: %#v", first)
	}
	if first.CaptureTime == nil || !first.CaptureTime.Equal(captureTime) {
		t.Fatalf("capture time was not preserved: %#v", first.CaptureTime)
	}

	params.UploadID = uploadID2
	second, replayed, err := database.CreateUploadSessionIdempotent(ctx, CreateUploadSessionIdempotentParams{
		IdempotencyKey: key,
		RequestSHA256:  hash,
		Session:        params,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || second.UploadID != uploadID1 {
		t.Fatalf("expected replay of original session, got replayed=%v session=%#v", replayed, second)
	}

	if _, _, err := database.CreateUploadSessionIdempotent(ctx, CreateUploadSessionIdempotentParams{
		IdempotencyKey: key,
		RequestSHA256:  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Session:        params,
	}); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
}
