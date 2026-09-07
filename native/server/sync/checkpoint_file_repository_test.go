package sync

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileUploadCheckpointRepositoryCreateLoadAndSave(t *testing.T) {
	now := time.Date(2026, 9, 7, 21, 30, 0, 0, time.UTC)
	repository, err := NewFileUploadCheckpointRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewUploadCheckpoint("owner/../one", "media/../../one", 512, now)
	if err != nil {
		t.Fatal(err)
	}

	created, err := repository.Create(context.Background(), checkpoint)
	if err != nil || created.Revision() != 0 {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	path := repository.recordPath("owner/../one", "media/../../one")
	if filepath.Dir(path) != repository.root || strings.Contains(filepath.Base(path), "owner") || strings.Contains(filepath.Base(path), "media") {
		t.Fatalf("unsafe checkpoint path %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("checkpoint permissions=%#o", info.Mode().Perm())
	}

	loaded, found, err := repository.Load(context.Background(), "owner/../one", "media/../../one")
	if err != nil || !found || !sameUploadCheckpointRecord(loaded, created) {
		t.Fatalf("loaded=%+v found=%v err=%v", loaded, found, err)
	}

	uploading, err := checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	saved, err := repository.Save(context.Background(), 0, uploading)
	if err != nil || saved.Revision() != 1 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	restored, err := saved.Restore()
	if err != nil || restored.Record().State != StateUploading || restored.Record().Attempts != 1 {
		t.Fatalf("restored=%+v err=%v", restored.Record(), err)
	}
}

func TestFileUploadCheckpointRepositoryRejectsDuplicateAndStaleWrites(t *testing.T) {
	now := time.Date(2026, 9, 7, 21, 30, 0, 0, time.UTC)
	repository, _ := NewFileUploadCheckpointRepository(t.TempDir())
	checkpoint, _ := NewUploadCheckpoint("owner-1", "media-1", 128, now)
	if _, err := repository.Create(context.Background(), checkpoint); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Create(context.Background(), checkpoint); !errors.Is(err, ErrUploadCheckpointAlreadyExists) {
		t.Fatalf("duplicate create error=%v", err)
	}

	uploading, _ := checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	current, err := repository.Save(context.Background(), 0, uploading)
	if err != nil {
		t.Fatal(err)
	}
	failed, _ := uploading.Transition(StateFailed, now.Add(2*time.Second), "network")
	returned, err := repository.Save(context.Background(), 0, failed)
	if !errors.Is(err, ErrStaleUploadCheckpointRevision) || returned.Revision() != current.Revision() {
		t.Fatalf("returned=%+v error=%v", returned, err)
	}
}

func TestFileUploadCheckpointRepositoryRejectsMalformedAndWrongIdentityFiles(t *testing.T) {
	now := time.Date(2026, 9, 7, 21, 30, 0, 0, time.UTC)
	repository, _ := NewFileUploadCheckpointRepository(t.TempDir())
	checkpoint, _ := NewUploadCheckpoint("owner-1", "media-1", 64, now)
	record, _ := NewUploadCheckpointRecord(checkpoint)

	path := repository.recordPath("owner-1", "media-1")
	if err := os.WriteFile(path, []byte(`{"revision":0,"payload":"not-base64"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "owner-1", "media-1"); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("malformed error=%v", err)
	}

	wrong, _ := NewUploadCheckpoint("owner-2", "media-2", 64, now)
	wrongRecord, _ := NewUploadCheckpointRecord(wrong)
	encoded, _ := encodeUploadCheckpointFileEnvelope(wrongRecord)
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "owner-1", "media-1"); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("wrong identity error=%v", err)
	}

	canonical, _ := encodeUploadCheckpointFileEnvelope(record)
	if err := os.WriteFile(path, append(canonical, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "owner-1", "media-1"); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("noncanonical error=%v", err)
	}
}

func TestFileUploadCheckpointRepositoryNotFoundCancellationAndBounds(t *testing.T) {
	if _, err := NewFileUploadCheckpointRepository("   "); !errors.Is(err, ErrInvalidUploadCheckpointFileRepository) {
		t.Fatalf("blank root error=%v", err)
	}
	repository, _ := NewFileUploadCheckpointRepository(t.TempDir())
	if _, found, err := repository.Load(context.Background(), "owner-1", "missing"); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	checkpoint, _ := NewUploadCheckpoint("owner-1", "media-1", 1, time.Now().UTC())
	if _, err := repository.Create(ctx, checkpoint); !errors.Is(err, context.Canceled) {
		t.Fatalf("create cancellation=%v", err)
	}
	if _, _, err := repository.Load(ctx, "owner-1", "media-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("load cancellation=%v", err)
	}

	path := repository.recordPath("owner-1", "oversized")
	if err := os.WriteFile(path, make([]byte, MaxUploadCheckpointFileBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "owner-1", "oversized"); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("oversized error=%v", err)
	}
}
