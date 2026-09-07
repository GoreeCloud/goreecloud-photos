package sync

import (
	"errors"
	"strings"
)

// RestoreUploadCheckpoint reconstructs a storage-neutral checkpoint only from
// exact, internally coherent lifecycle and progress state. Persistence adapters
// should use this boundary rather than populating UploadCheckpoint fields.
func RestoreUploadCheckpoint(record Record, progress UploadProgress) (UploadCheckpoint, error) {
	if err := validateRestoredRecord(record); err != nil {
		return UploadCheckpoint{}, err
	}
	if !progress.valid() {
		return UploadCheckpoint{}, errors.New("invalid restored upload progress")
	}
	if record.State == StateSynced && !progress.Complete() {
		return UploadCheckpoint{}, errors.New("synced checkpoint requires complete upload progress")
	}
	return UploadCheckpoint{record: record, progress: progress}, nil
}

func validateRestoredRecord(record Record) error {
	if strings.TrimSpace(record.OwnerID) == "" || strings.TrimSpace(record.MediaID) == "" {
		return errors.New("restored checkpoint requires owner id and media id")
	}
	if record.UpdatedAt.IsZero() {
		return errors.New("restored checkpoint requires sync timestamp")
	}
	if record.Attempts < 0 {
		return errors.New("restored checkpoint attempts cannot be negative")
	}
	if record.State != StatePending && record.State != StateUploading && record.State != StateSynced && record.State != StateFailed {
		return errors.New("invalid restored sync state")
	}

	failureCode := strings.TrimSpace(record.FailureCode)
	if record.State == StateFailed {
		if failureCode == "" || failureCode != record.FailureCode {
			return errors.New("failed restored checkpoint requires canonical failure code")
		}
	} else if record.FailureCode != "" {
		return errors.New("non-failed restored checkpoint cannot carry failure code")
	}
	return nil
}
