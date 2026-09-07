package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeUploadCheckpointRepository struct {
	record       UploadCheckpointRecord
	found        bool
	loadErr      error
	createErr    error
	saveErr      error
	createResult *UploadCheckpointRecord
	saveResult   *UploadCheckpointRecord
	creates      int
	saves        int
}

func (f *fakeUploadCheckpointRepository) Load(context.Context, string, string) (UploadCheckpointRecord, bool, error) {
	return f.record, f.found, f.loadErr
}

func (f *fakeUploadCheckpointRepository) Create(
	_ context.Context,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	f.creates++
	if f.createErr != nil {
		return UploadCheckpointRecord{}, f.createErr
	}
	if f.createResult != nil {
		return *f.createResult, nil
	}
	record, err := NewUploadCheckpointRecord(checkpoint)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	f.record = record
	f.found = true
	return record, nil
}

func (f *fakeUploadCheckpointRepository) Save(
	_ context.Context,
	expected UploadCheckpointRevision,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	f.saves++
	if f.saveErr != nil {
		return f.record, f.saveErr
	}
	if f.saveResult != nil {
		return *f.saveResult, nil
	}
	next, err := f.record.Update(expected, checkpoint)
	if err != nil {
		return f.record, err
	}
	f.record = next
	return next, nil
}

func TestInitializeStoredUploadCheckpointValidatesRepositoryResult(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 25, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	repository := &fakeUploadCheckpointRepository{}

	record, err := InitializeStoredUploadCheckpoint(context.Background(), repository, checkpoint)
	if err != nil || record.Revision() != 0 || repository.creates != 1 {
		t.Fatalf("record=%+v creates=%d err=%v", record, repository.creates, err)
	}

	wrong, err := NewUploadCheckpoint("owner-2", "media-2", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	wrongRecord, err := NewUploadCheckpointRecord(wrong)
	if err != nil {
		t.Fatal(err)
	}
	repository = &fakeUploadCheckpointRepository{createResult: &wrongRecord}
	if _, err := InitializeStoredUploadCheckpoint(context.Background(), repository, checkpoint); !errors.Is(err, ErrInvalidUploadCheckpointRepositoryResult) {
		t.Fatalf("error=%v", err)
	}
}

func TestMutateStoredUploadCheckpointSavesCanonicalNextRevision(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 25, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	record, _ := NewUploadCheckpointRecord(checkpoint)
	repository := &fakeUploadCheckpointRepository{record: record, found: true}

	next, err := MutateStoredUploadCheckpoint(context.Background(), repository, "owner-1", "media-1", func(current UploadCheckpoint) (UploadCheckpoint, error) {
		return current.Transition(StateUploading, now.Add(time.Second), "")
	})
	if err != nil || next.Revision() != 1 || repository.saves != 1 {
		t.Fatalf("next=%+v saves=%d err=%v", next, repository.saves, err)
	}
	restored, err := next.Restore()
	if err != nil || restored.Record().State != StateUploading || restored.Record().Attempts != 1 {
		t.Fatalf("restored=%+v err=%v", restored.Record(), err)
	}
}

func TestMutateStoredUploadCheckpointRejectsWrongLoadAndSaveResults(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 25, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	record, _ := NewUploadCheckpointRecord(checkpoint)
	wrong, _ := NewUploadCheckpoint("owner-2", "media-2", 100, now)
	wrongRecord, _ := NewUploadCheckpointRecord(wrong)

	repository := &fakeUploadCheckpointRepository{record: wrongRecord, found: true}
	if _, err := MutateStoredUploadCheckpoint(context.Background(), repository, "owner-1", "media-1", func(current UploadCheckpoint) (UploadCheckpoint, error) {
		return current, nil
	}); !errors.Is(err, ErrInvalidUploadCheckpointRepositoryResult) {
		t.Fatalf("wrong load error=%v", err)
	}

	repository = &fakeUploadCheckpointRepository{record: record, found: true, saveResult: &wrongRecord}
	if _, err := MutateStoredUploadCheckpoint(context.Background(), repository, "owner-1", "media-1", func(current UploadCheckpoint) (UploadCheckpoint, error) {
		return current.Transition(StateUploading, now.Add(time.Second), "")
	}); !errors.Is(err, ErrInvalidUploadCheckpointRepositoryResult) {
		t.Fatalf("wrong save error=%v", err)
	}
}

func TestMutateStoredUploadCheckpointPropagatesNotFoundMutationAndContext(t *testing.T) {
	repository := &fakeUploadCheckpointRepository{}
	if _, err := MutateStoredUploadCheckpoint(context.Background(), repository, "owner-1", "media-1", func(current UploadCheckpoint) (UploadCheckpoint, error) {
		return current, nil
	}); !errors.Is(err, ErrUploadCheckpointRecordNotFound) {
		t.Fatalf("not found error=%v", err)
	}

	now := time.Date(2026, 9, 7, 20, 25, 0, 0, time.UTC)
	checkpoint := testStoredCheckpoint(t, now)
	record, _ := NewUploadCheckpointRecord(checkpoint)
	repository = &fakeUploadCheckpointRepository{record: record, found: true}
	want := errors.New("mutation failed")
	if _, err := MutateStoredUploadCheckpoint(context.Background(), repository, "owner-1", "media-1", func(UploadCheckpoint) (UploadCheckpoint, error) {
		return UploadCheckpoint{}, want
	}); !errors.Is(err, want) || repository.saves != 0 {
		t.Fatalf("mutation error=%v saves=%d", err, repository.saves)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := MutateStoredUploadCheckpoint(ctx, repository, "owner-1", "media-1", func(current UploadCheckpoint) (UploadCheckpoint, error) {
		return current, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("context error=%v", err)
	}
}
