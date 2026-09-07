package sync

import (
	"errors"
	"strings"
	"time"
)

// TransferResult is the executor-neutral outcome applied to one previously
// planned transfer step. CommittedBytes is an absolute resumable offset, not a
// byte count. FailureCode is a stable caller-supplied code and never an error
// string or transport payload.
type TransferResult struct {
	CommittedBytes int64
	FailureCode    string
}

// ApplyTransferStep folds one executor result back into the validated upload
// checkpoint. It does not perform network I/O, persist state, sleep, observe
// connectivity, or manufacture transport acknowledgements.
func ApplyTransferStep(
	checkpoint UploadCheckpoint,
	step TransferStep,
	now time.Time,
	result TransferResult,
) (UploadCheckpoint, error) {
	if _, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress); err != nil {
		return checkpoint, errors.New("invalid transfer checkpoint")
	}

	failureCode := strings.TrimSpace(result.FailureCode)
	if failureCode != result.FailureCode {
		return checkpoint, errors.New("transfer failure code must already be canonical")
	}

	switch step.Action {
	case TransferActionStartUpload:
		if checkpoint.record.State != StatePending || result.CommittedBytes != 0 {
			return checkpoint, errors.New("stale or invalid start-upload result")
		}
		if failureCode != "" {
			return checkpoint.Transition(StateFailed, now, failureCode)
		}
		return checkpoint.Transition(StateUploading, now, "")

	case TransferActionSendChunk:
		if checkpoint.record.State != StateUploading {
			return checkpoint, errors.New("stale send-chunk result")
		}
		if failureCode == "" && result.CommittedBytes <= step.Chunk.Offset {
			return checkpoint, errors.New("successful chunk result must advance committed bytes")
		}
		updated, err := checkpoint.AcknowledgeChunk(step.Chunk, result.CommittedBytes)
		if err != nil {
			return checkpoint, err
		}
		if failureCode != "" {
			return updated.Transition(StateFailed, now, failureCode)
		}
		return updated, nil

	case TransferActionMarkSynced:
		if checkpoint.record.State != StateUploading || !checkpoint.progress.Complete() || result.CommittedBytes != 0 {
			return checkpoint, errors.New("stale or invalid mark-synced result")
		}
		if failureCode != "" {
			return checkpoint.Transition(StateFailed, now, failureCode)
		}
		return checkpoint.Transition(StateSynced, now, "")

	case TransferActionRetryReady:
		if checkpoint.record.State != StateFailed || result.CommittedBytes != 0 || failureCode != "" || step.RetryAt.IsZero() || now.IsZero() {
			return checkpoint, errors.New("stale or invalid retry-ready result")
		}
		now = now.UTC()
		if now.Before(step.RetryAt) {
			return checkpoint, errors.New("retry-ready result arrived before retry time")
		}
		return checkpoint.Transition(StatePending, now, "")

	case TransferActionBlocked:
		if checkpoint.record.State != StateUploading || checkpoint.progress.Complete() || result.CommittedBytes != 0 || failureCode != "" {
			return checkpoint, errors.New("stale or invalid blocked result")
		}
		return checkpoint, nil

	case TransferActionWaitRetry:
		if checkpoint.record.State != StateFailed || result.CommittedBytes != 0 || failureCode != "" || step.RetryAt.IsZero() || now.IsZero() {
			return checkpoint, errors.New("stale or invalid wait-retry result")
		}
		if !now.UTC().Before(step.RetryAt) {
			return checkpoint, errors.New("wait-retry result is no longer waiting")
		}
		return checkpoint, nil

	case TransferActionExhausted:
		if checkpoint.record.State != StateFailed || result.CommittedBytes != 0 || failureCode != "" {
			return checkpoint, errors.New("stale or invalid exhausted result")
		}
		return checkpoint, nil

	case TransferActionComplete:
		if checkpoint.record.State != StateSynced || result.CommittedBytes != 0 || failureCode != "" {
			return checkpoint, errors.New("stale or invalid complete result")
		}
		return checkpoint, nil

	default:
		return checkpoint, errors.New("unknown transfer action")
	}
}
