package upload_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/database"
	"github.com/GoreeCloud/goreecloud-photos/internal/storage"
	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUploadServicePersistsPartsAcrossServiceRestart(t *testing.T) {
	connectionString := os.Getenv("GC_PHOTOS_TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("GC_PHOTOS_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"000001_core.up.sql", "000002_upload_parts.up.sql"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		connection, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, execErr := connection.Conn().PgConn().Exec(ctx, string(data)).ReadAll()
		connection.Release()
		if execErr != nil {
			t.Fatalf("apply %s: %v", name, execErr)
		}
	}

	const libraryID = "01999999-0000-7000-8000-000000000001"
	const subjectID = "subject-integration"
	if _, err := pool.Exec(ctx, `
		INSERT INTO libraries (library_id, library_type, owner_subject_id)
		VALUES ($1::uuid, 'personal', $2);
		INSERT INTO library_memberships (library_id, subject_id, role)
		VALUES ($1::uuid, $2, 'owner');
	`, libraryID, subjectID); err != nil {
		t.Fatalf("seed library: %v", err)
	}

	db, err := database.NewPostgreSQL(ctx, connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := storage.NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := upload.NewService(db, store)

	payload := []byte("durable-part")
	session, err := service.Create(ctx, subjectID, upload.CreateRequest{
		LibraryID: libraryID, OriginalFilename: "photo.jpg", MediaType: "image/jpeg",
		ExpectedSize: int64(len(payload)), ExpectedSHA256: upload.SHA256(payload), DeviceID: "integration-device",
	})
	if err != nil {
		t.Fatal(err)
	}
	afterPart, err := service.PutPart(ctx, subjectID, session.UploadID, 1, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if afterPart.ReceivedBytes != int64(len(payload)) || len(afterPart.Parts) != 1 {
		t.Fatalf("unexpected persisted session: %#v", afterPart)
	}

	restartedService := upload.NewService(db, store)
	afterRestart, err := restartedService.Get(ctx, subjectID, session.UploadID)
	if err != nil {
		t.Fatal(err)
	}
	if afterRestart.ReceivedBytes != int64(len(payload)) || len(afterRestart.Parts) != 1 {
		t.Fatalf("state did not survive service restart: %#v", afterRestart)
	}
	retry, err := restartedService.PutPart(ctx, subjectID, session.UploadID, 1, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if retry.ReceivedBytes != int64(len(payload)) {
		t.Fatalf("idempotent retry changed received bytes: %d", retry.ReceivedBytes)
	}

	conflict := append([]byte(nil), payload...)
	conflict[0] ^= 0xff
	if _, err := restartedService.PutPart(ctx, subjectID, session.UploadID, 1, bytes.NewReader(conflict)); !errors.Is(err, upload.ErrPartConflict) {
		t.Fatalf("expected ErrPartConflict, got %v", err)
	}
	cancelled, err := restartedService.Cancel(ctx, subjectID, session.UploadID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.State != "cancelled" {
		t.Fatalf("unexpected state: %q", cancelled.State)
	}
	if _, err := restartedService.PutPart(ctx, subjectID, session.UploadID, 1, bytes.NewReader(payload)); !errors.Is(err, upload.ErrInvalidState) {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
}
