package sync

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestUploadCheckpointRecordAdvancesRevisionAndRestoresProgress(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	record, err := NewUploadCheckpointRecord(checkpoint)
	if err != nil || record.Revision() != 0 {
		t.Fatalf("record=%+v err=%v", record, err)
	}

	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	chunk, ok, err := checkpoint.NextChunk(20)
	if err != nil || !ok {
		t.Fatalf("chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
	checkpoint, err = checkpoint.AcknowledgeChunk(chunk, 10)
	if err != nil {
		t.Fatal(err)
	}
	next, err := record.Update(0, checkpoint)
	if err != nil || next.Revision() != 1 {
		t.Fatalf("next=%+v err=%v", next, err)
	}
	restored, err := next.Restore()
	if err != nil || restored.Record().State != StateUploading || restored.Progress().CommittedBytes != 10 {
		t.Fatalf("record=%+v progress=%+v err=%v", restored.Record(), restored.Progress(), err)
	}
}

func TestUploadCheckpointRecordRejectsStaleWriterAndIdentityChanges(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	record, _ := NewUploadCheckpointRecord(checkpoint)

	if _, err := record.Update(1, checkpoint); !errors.Is(err, ErrStaleUploadCheckpointRevision) {
		t.Fatalf("error=%v", err)
	}

	changedOwner := checkpoint
	changedOwner.record.OwnerID = "other-owner"
	if _, err := record.Update(0, changedOwner); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("owner error=%v", err)
	}
	changedMedia := checkpoint
	changedMedia.record.MediaID = "other-media"
	if _, err := record.Update(0, changedMedia); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("media error=%v", err)
	}
	changedSize := checkpoint
	changedSize.progress.TotalBytes++
	if _, err := record.Update(0, changedSize); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("size error=%v", err)
	}
}

func TestUploadCheckpointRecordCopiesPayloadAndRejectsOverflow(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	payload, err := EncodeUploadCheckpoint(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	record, err := RestoreUploadCheckpointRecord(7, payload)
	if err != nil || record.Revision() != 7 {
		t.Fatalf("record=%+v err=%v", record, err)
	}

	copyOut := record.Payload()
	copyOut[0] = 'x'
	payload[0] = 'x'
	if _, err := record.Restore(); err != nil {
		t.Fatal("payload mutation escaped into record")
	}

	record.revision = UploadCheckpointRevision(math.MaxUint64)
	if _, err := record.Update(record.Revision(), checkpoint); !errors.Is(err, ErrUploadCheckpointRevisionExhausted) {
		t.Fatalf("error=%v", err)
	}
}

func TestUploadCheckpointRecordRejectsMalformedStoredPayload(t *testing.T) {
	if _, err := RestoreUploadCheckpointRecord(1, nil); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("error=%v", err)
	}
	if _, err := RestoreUploadCheckpointRecord(1, []byte(`{"schemaVersion":1}`)); !errors.Is(err, ErrInvalidUploadCheckpointRecord) {
		t.Fatalf("error=%v", err)
	}
}

func testStoredCheckpoint(t *testing.T, now time.Time) UploadCheckpoint {
	t.Helper()
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	return checkpoint
}
