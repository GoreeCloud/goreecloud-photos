package sync

import (
	"testing"
	"time"
)

func TestSyncRecordHappyPath(t *testing.T) {
	now := time.Unix(100, 0)
	record, err := NewRecord("user-1", "media-1", now)
	if err != nil {
		t.Fatal(err)
	}
	record, err = record.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	record, err = record.Transition(StateSynced, now.Add(2*time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	if record.State != StateSynced {
		t.Fatalf("state = %q, want %q", record.State, StateSynced)
	}
}

func TestSyncRecordFailureRequiresCode(t *testing.T) {
	record, err := NewRecord("user-1", "media-1", time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Transition(StateFailed, time.Unix(101, 0), ""); err == nil {
		t.Fatal("expected failure code requirement")
	}
}

func TestSyncRecordRejectsSkippedUploadState(t *testing.T) {
	record, err := NewRecord("user-1", "media-1", time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Transition(StateSynced, time.Unix(101, 0), ""); err == nil {
		t.Fatal("expected pending-to-synced transition to be rejected")
	}
}
