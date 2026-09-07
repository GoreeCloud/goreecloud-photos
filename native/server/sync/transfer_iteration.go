package sync

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidTransferIteration = errors.New("invalid transfer iteration")

// UploadChunkExecutor is the narrow transport boundary used by one transfer
// iteration. Implementations may send only the already-planned resumable chunk
// and return an absolute committed-byte acknowledgement. Stable transfer
// failure codes belong in TransferResult; raw executor errors are propagated
// without manufacturing checkpoint failure state.
type UploadChunkExecutor interface {
	ExecuteChunk(ctx context.Context, ownerID, mediaID string, chunk UploadChunk) (TransferResult, error)
}

type TransferIteration struct {
	Planned   StoredTransferStep
	Record    UploadCheckpointRecord
	Persisted bool
	Terminal  bool
}

// RunStoredTransferIteration performs at most one transfer action. Internal
// lifecycle actions are persisted directly, send-chunk invokes the transport
// boundary exactly once, and blocked/wait/terminal observations do not create
// revision churn. It never loops, sleeps, polls, or observes OS connectivity.
func RunStoredTransferIteration(
	ctx context.Context,
	repository UploadCheckpointRepository,
	executor UploadChunkExecutor,
	ownerID, mediaID string,
	retryPolicy RetryPolicy,
	bandwidthPolicy BandwidthPolicy,
	network TransferNetwork,
	window time.Duration,
	now time.Time,
) (TransferIteration, error) {
	planned, err := PlanStoredTransferStep(
		ctx,
		repository,
		ownerID,
		mediaID,
		retryPolicy,
		bandwidthPolicy,
		network,
		window,
		now,
	)
	if err != nil {
		return TransferIteration{}, err
	}

	iteration := TransferIteration{Planned: planned}
	switch planned.Step.Action {
	case TransferActionSendChunk:
		if executor == nil {
			return iteration, ErrInvalidTransferIteration
		}
		if err := ctx.Err(); err != nil {
			return iteration, err
		}
		result, err := executor.ExecuteChunk(ctx, ownerID, mediaID, planned.Step.Chunk)
		if err != nil {
			return iteration, err
		}
		record, err := ApplyStoredTransferStep(ctx, repository, ownerID, mediaID, planned, now, result)
		if err != nil {
			return iteration, err
		}
		iteration.Record = record
		iteration.Persisted = true
		return iteration, nil

	case TransferActionStartUpload, TransferActionMarkSynced, TransferActionRetryReady:
		record, err := ApplyStoredTransferStep(ctx, repository, ownerID, mediaID, planned, now, TransferResult{})
		if err != nil {
			return iteration, err
		}
		iteration.Record = record
		iteration.Persisted = true
		return iteration, nil

	case TransferActionBlocked, TransferActionWaitRetry:
		return iteration, nil

	case TransferActionExhausted, TransferActionComplete:
		iteration.Terminal = true
		return iteration, nil

	default:
		return iteration, ErrInvalidTransferIteration
	}
}
