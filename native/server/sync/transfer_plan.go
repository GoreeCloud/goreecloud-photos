package sync

import (
	"errors"
	"time"
)

type TransferAction string

const (
	TransferActionStartUpload TransferAction = "start-upload"
	TransferActionSendChunk   TransferAction = "send-chunk"
	TransferActionBlocked     TransferAction = "blocked"
	TransferActionMarkSynced  TransferAction = "mark-synced"
	TransferActionWaitRetry   TransferAction = "wait-retry"
	TransferActionRetryReady  TransferAction = "retry-ready"
	TransferActionExhausted   TransferAction = "exhausted"
	TransferActionComplete    TransferAction = "complete"
)

type TransferStep struct {
	Action  TransferAction
	Chunk   UploadChunk
	RetryAt time.Time
}

// PlanTransferStep converts the validated Keepsake Sync state into the next
// executor-facing action. It never mutates the checkpoint, sleeps, observes the
// operating-system network, reads media, persists state, or executes transport.
func PlanTransferStep(
	checkpoint UploadCheckpoint,
	retryPolicy RetryPolicy,
	bandwidthPolicy BandwidthPolicy,
	network TransferNetwork,
	window time.Duration,
	now time.Time,
) (TransferStep, error) {
	if _, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress); err != nil {
		return TransferStep{}, errors.New("invalid transfer checkpoint")
	}

	switch checkpoint.record.State {
	case StatePending:
		return TransferStep{Action: TransferActionStartUpload}, nil
	case StateUploading:
		if checkpoint.progress.Complete() {
			return TransferStep{Action: TransferActionMarkSynced}, nil
		}
		chunk, ok, err := checkpoint.NextChunkWithBandwidth(bandwidthPolicy, network, window)
		if err != nil {
			return TransferStep{}, err
		}
		if !ok {
			return TransferStep{Action: TransferActionBlocked}, nil
		}
		return TransferStep{Action: TransferActionSendChunk, Chunk: chunk}, nil
	case StateFailed:
		retryAt, retryable, err := checkpoint.NextRetryAt(retryPolicy)
		if err != nil {
			return TransferStep{}, err
		}
		if !retryable {
			return TransferStep{Action: TransferActionExhausted}, nil
		}
		if now.IsZero() {
			return TransferStep{}, errors.New("retry planning requires current time")
		}
		now = now.UTC()
		if now.Before(retryAt) {
			return TransferStep{Action: TransferActionWaitRetry, RetryAt: retryAt}, nil
		}
		return TransferStep{Action: TransferActionRetryReady, RetryAt: retryAt}, nil
	case StateSynced:
		return TransferStep{Action: TransferActionComplete}, nil
	default:
		return TransferStep{}, errors.New("invalid transfer state")
	}
}
