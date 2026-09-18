package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"
)

func TestUploadSessionPersistenceAndPartIdempotency(t *testing.T) {
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

	if err := resetAndApplyMigrations(ctx, database); err != nil {
		database.Close()
		t.Fatalf("prepare database: %v", err)
	}

	const (
		libraryID = "01997d3a-0000-7000-8000-000000000001"
		uploadID  = "01997d3a-0000-7000-8000-000000000002"
		actorID   = "subject:test-user"
	)
	if _, err := database.pool.Exec(ctx, `
		INSERT INTO libraries (library_id, library_type, owner_subject_id)
		VALUES ($1, 'personal', $2)
	`, libraryID, actorID); err != nil {
		database.Close()
		t.Fatalf("create test library: %v", err)
	}

	expectedPayload := []byte("abcdefghij")
	expectedHash := sha256Hex(expectedPayload)
	session, err := database.CreateUploadSession(ctx, CreateUploadSessionParams{
		UploadID:       uploadID,
		ActorSubjectID: actorID,
		LibraryID:      libraryID,
		ExpectedSize:   int64(len(expectedPayload)),
		ExpectedSHA256: &expectedHash,
		PartSize:       4,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	if err != nil {
		database.Close()
		t.Fatalf("create upload session: %v", err)
	}
	if session.State != "open" || session.ReceivedBytes != 0 || len(session.Parts) != 0 {
		database.Close()
		t.Fatalf("unexpected new session: %#v", session)
	}

	database.Close()
	database, err = NewPostgreSQL(ctx, connectionString)
	if err != nil {
		t.Fatalf("reopen PostgreSQL adapter: %v", err)
	}
	defer database.Close()

	session, err = database.GetUploadSession(ctx, uploadID)
	if err != nil {
		t.Fatalf("reload durable upload session: %v", err)
	}
	if session.ExpectedSHA256 == nil || *session.ExpectedSHA256 != expectedHash {
		t.Fatalf("expected checksum did not survive restart: %#v", session.ExpectedSHA256)
	}

	part3 := expectedPayload[8:10]
	session, err = database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 3,
		ByteSize:   int64(len(part3)),
		SHA256:     sha256Hex(part3),
	})
	if err != nil {
		t.Fatalf("record final part out of order: %v", err)
	}
	if session.ReceivedBytes != 2 || session.State != "receiving" || len(session.Parts) != 1 {
		t.Fatalf("unexpected state after out-of-order part: %#v", session)
	}

	part1 := expectedPayload[0:4]
	session, err = database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 1,
		ByteSize:   int64(len(part1)),
		SHA256:     sha256Hex(part1),
	})
	if err != nil {
		t.Fatalf("record first part: %v", err)
	}

	part2 := expectedPayload[4:8]
	part2Hash := sha256Hex(part2)
	session, err = database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 2,
		ByteSize:   int64(len(part2)),
		SHA256:     part2Hash,
	})
	if err != nil {
		t.Fatalf("record second part: %v", err)
	}
	if session.ReceivedBytes != 10 || len(session.Parts) != 3 {
		t.Fatalf("unexpected completed receipt evidence: %#v", session)
	}
	for index, part := range session.Parts {
		if part.PartNumber != int32(index+1) {
			t.Fatalf("parts are not ordered: %#v", session.Parts)
		}
	}

	retried, err := database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 2,
		ByteSize:   int64(len(part2)),
		SHA256:     part2Hash,
	})
	if err != nil {
		t.Fatalf("retry identical part: %v", err)
	}
	if retried.ReceivedBytes != session.ReceivedBytes || len(retried.Parts) != len(session.Parts) {
		t.Fatalf("idempotent retry changed receipt evidence: before=%#v after=%#v", session, retried)
	}

	conflictingHash := sha256Hex([]byte("WXYZ"))
	if _, err := database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 2,
		ByteSize:   4,
		SHA256:     conflictingHash,
	}); !errors.Is(err, ErrUploadPartConflict) {
		t.Fatalf("expected ErrUploadPartConflict, got %v", err)
	}

	afterConflict, err := database.GetUploadSession(ctx, uploadID)
	if err != nil {
		t.Fatalf("load after conflict: %v", err)
	}
	if afterConflict.ReceivedBytes != 10 || len(afterConflict.Parts) != 3 || afterConflict.Parts[1].SHA256 != part2Hash {
		t.Fatalf("conflicting retry changed durable evidence: %#v", afterConflict)
	}
}

func TestUploadSessionRejectsInvalidPartBoundsAndPersistsExpiry(t *testing.T) {
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
		libraryID = "01997d3a-0000-7000-8000-000000000011"
		uploadID  = "01997d3a-0000-7000-8000-000000000012"
		actorID   = "subject:expiry-test"
	)
	if _, err := database.pool.Exec(ctx, `
		INSERT INTO libraries (library_id, library_type, owner_subject_id)
		VALUES ($1, 'personal', $2)
	`, libraryID, actorID); err != nil {
		t.Fatalf("create test library: %v", err)
	}

	if _, err := database.CreateUploadSession(ctx, CreateUploadSessionParams{
		UploadID:       uploadID,
		ActorSubjectID: actorID,
		LibraryID:      libraryID,
		ExpectedSize:   10,
		PartSize:       4,
		ExpiresAt:      time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create upload session: %v", err)
	}

	if _, err := database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 1,
		ByteSize:   3,
		SHA256:     sha256Hex([]byte("abc")),
	}); !errors.Is(err, ErrInvalidUploadPart) {
		t.Fatalf("expected partial non-final part rejection, got %v", err)
	}

	if _, err := database.pool.Exec(ctx, `
		UPDATE upload_sessions
		SET expires_at = NOW() - INTERVAL '1 second'
		WHERE upload_id = $1
	`, uploadID); err != nil {
		t.Fatalf("force upload session expiry: %v", err)
	}

	if _, err := database.RecordUploadPart(ctx, RecordUploadPartParams{
		UploadID:   uploadID,
		PartNumber: 1,
		ByteSize:   4,
		SHA256:     sha256Hex([]byte("abcd")),
	}); !errors.Is(err, ErrUploadSessionExpired) {
		t.Fatalf("expected ErrUploadSessionExpired, got %v", err)
	}

	session, err := database.GetUploadSession(ctx, uploadID)
	if err != nil {
		t.Fatalf("load expired session: %v", err)
	}
	if session.State != "expired" || session.ReceivedBytes != 0 || len(session.Parts) != 0 {
		t.Fatalf("unexpected expired session state: %#v", session)
	}
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
