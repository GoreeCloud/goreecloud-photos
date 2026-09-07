package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeUploadChunkExecutor struct {
	result TransferResult
	err    error
	calls  int
	chunk  UploadChunk
}

func (f *fakeUploadChunkExecutor) ExecuteChunk(
	_ context.Context,
	_, _ string,
	chunk UploadChunk,
) (TransferResult, error) {
	f.calls++
	f.chunk = chunk
	return f.result, f.err
}

func TestRunStoredTransferIterationPersistsInternalStartWithoutExecutor(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)

	iteration, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		nil,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		now.Add(time.Second),
	)
	if err != nil || !iteration.Persisted || iteration.Terminal || iteration.Planned.Step.Action != TransferActionStartUpload {
		t.Fatalf("iteration=%+v err=%v", iteration, err)
	}
	checkpoint, err := iteration.Record.Restore()
	if err != nil || checkpoint.Record().State != StateUploading || checkpoint.Record().Attempts != 1 {
		t.Fatalf("record=%+v err=%v", checkpoint.Record(), err)
	}
}

func TestRunStoredTransferIterationExecutesOneChunkAndPersistsAcknowledgement(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
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
	executor := &fakeUploadChunkExecutor{result: TransferResult{CommittedBytes: 25}}

	iteration, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		executor,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if err != nil || executor.calls != 1 || !iteration.Persisted || iteration.Planned.Step.Action != TransferActionSendChunk {
		t.Fatalf("iteration=%+v calls=%d err=%v", iteration, executor.calls, err)
	}
	if executor.chunk.Offset != 0 || executor.chunk.Length != 40 {
		t.Fatalf("chunk=%+v", executor.chunk)
	}
	checkpoint, err := iteration.Record.Restore()
	if err != nil || checkpoint.Progress().CommittedBytes != 25 {
		t.Fatalf("progress=%+v err=%v", checkpoint.Progress(), err)
	}
}

func TestRunStoredTransferIterationPropagatesExecutorErrorWithoutPersistence(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
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
	want := errors.New("transport unavailable")
	executor := &fakeUploadChunkExecutor{err: want}
	before := repository.saves

	iteration, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		executor,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if !errors.Is(err, want) || executor.calls != 1 || iteration.Persisted || repository.saves != before {
		t.Fatalf("iteration=%+v calls=%d saves=%d err=%v", iteration, executor.calls, repository.saves, err)
	}
	checkpoint, restoreErr := repository.record.Restore()
	if restoreErr != nil || checkpoint.Record().State != StateUploading || checkpoint.Progress().CommittedBytes != 0 {
		t.Fatalf("record=%+v progress=%+v err=%v", checkpoint.Record(), checkpoint.Progress(), restoreErr)
	}
}

func TestRunStoredTransferIterationDoesNotPersistBlockedOrCompleteObservation(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
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
	blocked, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		nil,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40},
		TransferNetworkMetered,
		time.Second,
		time.Time{},
	)
	if err != nil || blocked.Persisted || blocked.Terminal || blocked.Planned.Step.Action != TransferActionBlocked || repository.saves != before {
		t.Fatalf("blocked=%+v saves=%d err=%v", blocked, repository.saves, err)
	}

	checkpoint, err := repository.record.Restore()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.progress.CommittedBytes = checkpoint.progress.TotalBytes
	repository.record, err = repository.record.Update(repository.record.Revision(), checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	repository.record, err = repository.record.Update(repository.record.Revision(), mustTransition(t, checkpoint, StateSynced, now.Add(2*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	before = repository.saves
	complete, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		nil,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	)
	if err != nil || complete.Persisted || !complete.Terminal || complete.Planned.Step.Action != TransferActionComplete || repository.saves != before {
		t.Fatalf("complete=%+v saves=%d err=%v", complete, repository.saves, err)
	}
}

func TestRunStoredTransferIterationRequiresExecutorOnlyForSendChunk(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	repository := newStoredTransferRepository(t, now)
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
	iteration, err := RunStoredTransferIteration(
		context.Background(),
		repository,
		nil,
		"owner-1",
		"media-1",
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if !errors.Is(err, ErrInvalidTransferIteration) || iteration.Planned.Step.Action != TransferActionSendChunk {
		t.Fatalf("iteration=%+v err=%v", iteration, err)
	}
}

func mustTransition(t *testing.T, checkpoint UploadCheckpoint, state State, now time.Time) UploadCheckpoint {
	t.Helper()
	next, err := checkpoint.Transition(state, now, "")
	if err != nil {
		t.Fatal(err)
	}
	return next
}
