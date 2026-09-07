package sync

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidStoredTransferCycle = errors.New("invalid stored transfer cycle")
	ErrStaleStoredTransferStep    = errors.New("stale stored transfer step")
)

// StoredTransferStep binds one executor-facing transfer action to the exact
// optimistic checkpoint revision from which it was planned. Applying the step
// later requires that same revision to still be current.
type StoredTransferStep struct {
	Revision UploadCheckpointRevision
	Step     TransferStep
}

// PlanStoredTransferStep loads and validates one persisted upload checkpoint,
// then derives the next pure transfer action without mutating or saving state.
func PlanStoredTransferStep(
	ctx context.Context,
	repository UploadCheckpointRepository,
	ownerID, mediaID string,
	retryPolicy RetryPolicy,
	bandwidthPolicy BandwidthPolicy,
	network TransferNetwork,
	window time.Duration,
	now time.Time,
) (StoredTransferStep, error) {
	if repository == nil || ownerID == "" || mediaID == "" {
		return StoredTransferStep{}, ErrInvalidStoredTransferCycle
	}
	if err := ctx.Err(); err != nil {
		return StoredTransferStep{}, err
	}

	record, found, err := repository.Load(ctx, ownerID, mediaID)
	if err != nil {
		return StoredTransferStep{}, err
	}
	if !found {
		return StoredTransferStep{}, ErrUploadCheckpointRecordNotFound
	}
	checkpoint, err := record.Restore()
	if err != nil || checkpoint.Record().OwnerID != ownerID || checkpoint.Record().MediaID != mediaID {
		return StoredTransferStep{}, ErrInvalidUploadCheckpointRepositoryResult
	}

	step, err := PlanTransferStep(checkpoint, retryPolicy, bandwidthPolicy, network, window, now)
	if err != nil {
		return StoredTransferStep{}, err
	}
	return StoredTransferStep{Revision: record.Revision(), Step: step}, nil
}

// ApplyStoredTransferStep performs one optimistic load-apply-save cycle for a
// previously planned transfer action. It rejects stale plans before mutation
// and accepts the saved state only when the repository returns the exact
// canonical next revision/payload.
func ApplyStoredTransferStep(
	ctx context.Context,
	repository UploadCheckpointRepository,
	ownerID, mediaID string,
	planned StoredTransferStep,
	now time.Time,
	result TransferResult,
) (UploadCheckpointRecord, error) {
	if repository == nil || ownerID == "" || mediaID == "" {
		return UploadCheckpointRecord{}, ErrInvalidStoredTransferCycle
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
	checkpoint, err := record.Restore()
	if err != nil || checkpoint.Record().OwnerID != ownerID || checkpoint.Record().MediaID != mediaID {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRepositoryResult
	}
	if record.Revision() != planned.Revision {
		return record, ErrStaleStoredTransferStep
	}

	nextCheckpoint, err := ApplyTransferStep(checkpoint, planned.Step, now, result)
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
