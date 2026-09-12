package sync

import (
	"testing"
	"time"
)

func TestRestoreUploadCheckpointAcceptsFailedPartialResumeState(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	record := Record{
		OwnerID:     "owner-1",
		MediaID:     "media-1",
		State:       StateFailed,
		UpdatedAt:   now,
		FailureCode: "transport",
		Attempts:    2,
	}
	progress := UploadProgress{TotalBytes: 100, CommittedBytes: 40}

	checkpoint, err := RestoreUploadCheckpoint(record, progress)
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Record() != record || checkpoint.Progress() != progress {
		t.Fatalf("restored checkpoint = %+v/%+v", checkpoint.Record(), checkpoint.Progress())
	}
}

func TestRestoreUploadCheckpointRejectsInvalidRecordState(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	base := Record{OwnerID: "owner-1", MediaID: "media-1", State: StatePending, UpdatedAt: now}
	progress := UploadProgress{TotalBytes: 100, CommittedBytes: 40}

	cases := []Record{
		{MediaID: "media-1", State: StatePending, UpdatedAt: now},
		{OwnerID: "owner-1", State: StatePending, UpdatedAt: now},
		{OwnerID: "owner-1", MediaID: "media-1", State: StatePending},
		{OwnerID: "owner-1", MediaID: "media-1", State: State("unknown"), UpdatedAt: now},
		{OwnerID: "owner-1", MediaID: "media-1", State: StatePending, UpdatedAt: now, Attempts: -1},
	}
	for _, record := range cases {
		if _, err := RestoreUploadCheckpoint(record, progress); err == nil {
			t.Fatalf("expected record %+v to be rejected", record)
		}
	}
	if _, err := RestoreUploadCheckpoint(base, progress); err != nil {
		t.Fatalf("valid base record rejected: %v", err)
	}
}

func TestRestoreUploadCheckpointEnforcesFailureCodeInvariant(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	progress := UploadProgress{TotalBytes: 100, CommittedBytes: 40}

	cases := []Record{
		{OwnerID: "owner-1", MediaID: "media-1", State: StateFailed, UpdatedAt: now},
		{OwnerID: "owner-1", MediaID: "media-1", State: StateFailed, UpdatedAt: now, FailureCode: " transport "},
		{OwnerID: "owner-1", MediaID: "media-1", State: StateUploading, UpdatedAt: now, FailureCode: "transport"},
	}
	for _, record := range cases {
		if _, err := RestoreUploadCheckpoint(record, progress); err == nil {
			t.Fatalf("expected failure-code record %+v to be rejected", record)
		}
	}
}

func TestRestoreUploadCheckpointRejectsMalformedProgress(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	record := Record{OwnerID: "owner-1", MediaID: "media-1", State: StateUploading, UpdatedAt: now, Attempts: 1}

	for _, progress := range []UploadProgress{
		{TotalBytes: 0, CommittedBytes: 0},
		{TotalBytes: 10, CommittedBytes: -1},
		{TotalBytes: 10, CommittedBytes: 11},
	} {
		if _, err := RestoreUploadCheckpoint(record, progress); err == nil {
			t.Fatalf("expected progress %+v to be rejected", progress)
		}
	}
}

func TestRestoreUploadCheckpointRequiresCompleteProgressForSyncedState(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC)
	record := Record{OwnerID: "owner-1", MediaID: "media-1", State: StateSynced, UpdatedAt: now, Attempts: 1}

	if _, err := RestoreUploadCheckpoint(record, UploadProgress{TotalBytes: 10, CommittedBytes: 9}); err == nil {
		t.Fatal("expected partial synced checkpoint to be rejected")
	}
	if _, err := RestoreUploadCheckpoint(record, UploadProgress{TotalBytes: 10, CommittedBytes: 10}); err != nil {
		t.Fatalf("complete synced checkpoint rejected: %v", err)
	}
}
