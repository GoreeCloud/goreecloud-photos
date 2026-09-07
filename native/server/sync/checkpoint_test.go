package sync

import (
	"errors"
	"testing"
	"time"
)

func TestUploadCheckpointRequiresUploadingStateForChunks(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := checkpoint.NextChunk(25); err == nil {
		t.Fatal("expected pending checkpoint to reject chunk planning")
	}

	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	chunk, ok, err := checkpoint.NextChunk(25)
	if err != nil || !ok || chunk.Offset != 0 || chunk.Length != 25 {
		t.Fatalf("unexpected chunk decision: chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
}

func TestUploadCheckpointAcknowledgementAdvancesBoundProgress(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	chunk, _, err := checkpoint.NextChunk(40)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.AcknowledgeChunk(chunk, 30)
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Progress().CommittedBytes != 30 {
		t.Fatalf("committed bytes = %d, want 30", checkpoint.Progress().CommittedBytes)
	}
}

func TestUploadCheckpointCannotSyncBeforeCompletion(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 10, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}

	unchanged, err := checkpoint.Transition(StateSynced, now.Add(2*time.Second), "")
	if err == nil {
		t.Fatal("expected incomplete checkpoint to reject synced state")
	}
	if unchanged.Record().State != StateUploading || unchanged.Progress().CommittedBytes != 0 {
		t.Fatal("failed synced transition mutated checkpoint")
	}
}

func TestUploadCheckpointCompletesThenSyncs(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 10, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	chunk, _, err := checkpoint.NextChunk(10)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.AcknowledgeChunk(chunk, chunk.EndOffset())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateSynced, now.Add(2*time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Record().State != StateSynced || !checkpoint.Progress().Complete() {
		t.Fatal("expected completed checkpoint to be synced")
	}
}

func TestUploadCheckpointDelegatesRetryPolicy(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 10, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateFailed, now.Add(2*time.Second), "transport")
	if err != nil {
		t.Fatal(err)
	}

	retryAt, ok, err := checkpoint.NextRetryAt(RetryPolicy{
		BaseDelay:   time.Second,
		MaxDelay:    8 * time.Second,
		MaxAttempts: 3,
	})
	if err != nil || !ok {
		t.Fatalf("unexpected retry decision: at=%v ok=%v err=%v", retryAt, ok, err)
	}
	if want := now.Add(3 * time.Second); !retryAt.Equal(want) {
		t.Fatalf("retry at = %v, want %v", retryAt, want)
	}
}

func TestUploadCheckpointRejectsInvalidConstruction(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	if _, err := NewUploadCheckpoint("", "media-1", 10, now); err == nil {
		t.Fatal("expected blank owner to fail")
	}
	if _, err := NewUploadCheckpoint("owner-1", "media-1", 0, now); err == nil {
		t.Fatal("expected zero total bytes to fail")
	}

	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 10, now)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = checkpoint.NextChunk(1)
	if err == nil || errors.Is(err, nil) {
		t.Fatal("expected pending state to reject chunk planning")
	}
}
