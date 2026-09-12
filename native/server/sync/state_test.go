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
	if record.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", record.Attempts)
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

func TestSyncRecordCountsRetryAttempts(t *testing.T) {
	now := time.Unix(100, 0)
	record, err := NewRecord("user-1", "media-1", now)
	if err != nil {
		t.Fatal(err)
	}
	record, err = record.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	record, err = record.Transition(StateFailed, now.Add(2*time.Second), "network")
	if err != nil {
		t.Fatal(err)
	}
	record, err = record.Transition(StatePending, now.Add(3*time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	if record.FailureCode != "" {
		t.Fatalf("failure code = %q, want cleared", record.FailureCode)
	}
	record, err = record.Transition(StateUploading, now.Add(4*time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	if record.Attempts != 2 {
		t.Fatalf("attempts = %d, want 2", record.Attempts)
	}
}

func TestSyncRecordRejectsBackwardTimestamp(t *testing.T) {
	now := time.Unix(100, 0)
	record, err := NewRecord("user-1", "media-1", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Transition(StateUploading, now.Add(-time.Second), ""); err == nil {
		t.Fatal("expected backward transition timestamp to be rejected")
	}
}

func TestNewSyncRecordRequiresTimestamp(t *testing.T) {
	if _, err := NewRecord("user-1", "media-1", time.Time{}); err == nil {
		t.Fatal("expected zero timestamp to be rejected")
	}
}
