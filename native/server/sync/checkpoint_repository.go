package sync

import (
	"context"
	"errors"
)

var (
	ErrInvalidUploadCheckpointRepository       = errors.New("invalid upload checkpoint repository")
	ErrUploadCheckpointRecordNotFound          = errors.New("upload checkpoint record not found")
	ErrInvalidUploadCheckpointMutation         = errors.New("invalid upload checkpoint mutation")
	ErrInvalidUploadCheckpointRepositoryResult = errors.New("invalid upload checkpoint repository result")
)

// UploadCheckpointRepository is the storage-neutral persistence boundary for
// resumable upload checkpoints. Implementations choose the actual durable store
// and must preserve UploadCheckpointRecord optimistic-revision semantics.
type UploadCheckpointRepository interface {
	Load(ctx context.Context, ownerID, mediaID string) (UploadCheckpointRecord, bool, error)
	Create(ctx context.Context, checkpoint UploadCheckpoint) (UploadCheckpointRecord, error)
	Save(ctx context.Context, expected UploadCheckpointRevision, checkpoint UploadCheckpoint) (UploadCheckpointRecord, error)
}

func InitializeStoredUploadCheckpoint(
	ctx context.Context,
	repository UploadCheckpointRepository,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	if repository == nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRepository
	}
	validated, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress)
	if err != nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRepository
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, err
	}
	expected, err := NewUploadCheckpointRecord(validated)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	created, err := repository.Create(ctx, validated)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	if !sameUploadCheckpointRecord(created, expected) {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRepositoryResult
	}
	return created, nil
}

// MutateStoredUploadCheckpoint performs one optimistic load-mutate-save cycle.
// The core validates immutable transfer identity and the repository's exact
// returned revision/payload before accepting the persisted result.
func MutateStoredUploadCheckpoint(
	ctx context.Context,
	repository UploadCheckpointRepository,
	ownerID, mediaID string,
	mutate func(UploadCheckpoint) (UploadCheckpoint, error),
) (UploadCheckpointRecord, error) {
	if repository == nil || mutate == nil || ownerID == "" || mediaID == "" {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointMutation
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, err
	}

	record, found, err := repository.Load(ctx, ownerID, mediaID)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	if !found {
		return UploadCheckpointRecord{}, ErrUploadCheckpointRecordNotFound
	}
	current, err := record.Restore()
	if err != nil || current.record.OwnerID != ownerID || current.record.MediaID != mediaID {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRepositoryResult
	}

	nextCheckpoint, err := mutate(current)
	if err != nil {
		return record, err
	}
	expected, err := record.Update(record.Revision(), nextCheckpoint)
	if err != nil {
		return record, err
	}
	if err := ctx.Err(); err != nil {
		return record, err
	}

	saved, err := repository.Save(ctx, record.Revision(), nextCheckpoint)
	if err != nil {
		return record, err
	}
	if !sameUploadCheckpointRecord(saved, expected) {
		return record, ErrInvalidUploadCheckpointRepositoryResult
	}
	return saved, nil
}
