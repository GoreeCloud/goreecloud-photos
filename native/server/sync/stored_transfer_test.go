package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStoredTransferCyclePlansAndAppliesStartAndPartialChunk(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 5, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)

	planned, err := PlanStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	)
	if err != nil || planned.Revision != 0 || planned.Step.Action != TransferActionStartUpload || repository.saves != 0 {
		t.Fatalf("planned=%+v saves=%d err=%v", planned, repository.saves, err)
	}

	record, err := ApplyStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		planned,
		now.Add(time.Second),
		TransferResult{},
	)
	if err != nil || record.Revision() != 1 || repository.saves != 1 {
		t.Fatalf("revision=%d saves=%d err=%v", record.Revision(), repository.saves, err)
	}
	checkpoint, err := record.Restore()
	if err != nil || checkpoint.Record().State != StateUploading || checkpoint.Record().Attempts != 1 {
		t.Fatalf("record=%+v err=%v", checkpoint.Record(), err)
	}

	planned, err = PlanStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if err != nil || planned.Revision != 1 || planned.Step.Action != TransferActionSendChunk || planned.Step.Chunk.Offset != 0 || planned.Step.Chunk.Length != 40 {
		t.Fatalf("planned=%+v err=%v", planned, err)
	}

	record, err = ApplyStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		planned,
		time.Time{},
		TransferResult{CommittedBytes: 20},
	)
	if err != nil || record.Revision() != 2 {
		t.Fatalf("revision=%d err=%v", record.Revision(), err)
	}
	checkpoint, err = record.Restore()
	if err != nil || checkpoint.Progress().CommittedBytes != 20 || checkpoint.Record().State != StateUploading {
		t.Fatalf("progress=%+v record=%+v err=%v", checkpoint.Progress(), checkpoint.Record(), err)
	}
}

func TestStoredTransferCycleRejectsStalePlannedRevision(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 5, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
	planned, err := PlanStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := MutateStoredUploadCheckpoint(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		func(current UploadCheckpoint) (UploadCheckpoint, error) {
			return current.Transition(StateUploading, now.Add(time.Second), "")
		},
	); err != nil {
		t.Fatal(err)
	}
	before := repository.saves
	if _, err := ApplyStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		planned,
		now.Add(2*time.Second),
		TransferResult{},
	); !errors.Is(err, ErrStaleStoredTransferStep) || repository.saves != before {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}
}

func TestStoredTransferCyclePropagatesNotFoundCancellationAndInvalidSaveResult(t *testing.T) {
	empty := &fakeUploadCheckpointRepository{}
	if _, err := PlanStoredTransferStep(
		context.Background(),
		empty,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	); !errors.Is(err, ErrUploadCheckpointRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	now := time.Date(2026, 9, 7, 20, 5, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := PlanStoredTransferStep(
		ctx,
		repository,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error=%v", err)
	}

	planned, err := PlanStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	)
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := NewUploadCheckpoint("owner-2", "media-2", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	wrongRecord, err := NewUploadCheckpointRecord(wrong)
	if err != nil {
		t.Fatal(err)
	}
	repository.saveResult = &wrongRecord
	if _, err := ApplyStoredTransferStep(
		context.Background(),
		repository,
		"owner-1",
		"media-1",
		planned,
		now.Add(time.Second),
		TransferResult{},
	); !errors.Is(err, ErrInvalidUploadCheckpointRepositoryResult) {
		t.Fatalf("invalid save error=%v", err)
	}
}

func newStoredTransferRepository(t *testing.T, now time.Time) *fakeUploadCheckpointRepository {
	t.Helper()
	checkpoint, err := NewUploadCheckpoint("owner-1", "media-1", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewUploadCheckpointRecord(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	return &fakeUploadCheckpointRepository{record: record, found: true}
}
