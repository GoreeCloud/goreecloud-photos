package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

const UploadCheckpointSnapshotSchemaVersion = 1

var (
	ErrInvalidUploadCheckpointSnapshot     = errors.New("invalid upload checkpoint snapshot")
	ErrUnsupportedUploadCheckpointSnapshot = errors.New("unsupported upload checkpoint snapshot version")
)

type persistedUploadCheckpoint struct {
	SchemaVersion  int    `json:"schemaVersion"`
	OwnerID        string `json:"ownerId"`
	MediaID        string `json:"mediaId"`
	State          State  `json:"state"`
	UpdatedAt      string `json:"updatedAt"`
	FailureCode    string `json:"failureCode"`
	Attempts       int    `json:"attempts"`
	TotalBytes     int64  `json:"totalBytes"`
	CommittedBytes int64  `json:"committedBytes"`
}

// EncodeUploadCheckpointSnapshot serializes a validated, storage-neutral
// checkpoint. The payload does not select a database, transport, or device API.
func EncodeUploadCheckpointSnapshot(checkpoint UploadCheckpoint) ([]byte, error) {
	if _, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUploadCheckpointSnapshot, err)
	}
	record := checkpoint.record
	return json.Marshal(persistedUploadCheckpoint{
		SchemaVersion:  UploadCheckpointSnapshotSchemaVersion,
		OwnerID:        record.OwnerID,
		MediaID:        record.MediaID,
		State:          record.State,
		UpdatedAt:      record.UpdatedAt.UTC().Format(time.RFC3339Nano),
		FailureCode:    record.FailureCode,
		Attempts:       record.Attempts,
		TotalBytes:     checkpoint.progress.TotalBytes,
		CommittedBytes: checkpoint.progress.CommittedBytes,
	})
}

// DecodeUploadCheckpointSnapshot accepts only the current exact schema and
// routes reconstructed state through RestoreUploadCheckpoint.
func DecodeUploadCheckpointSnapshot(data []byte) (UploadCheckpoint, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var persisted persistedUploadCheckpoint
	if err := decoder.Decode(&persisted); err != nil {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointSnapshot
	}
	if persisted.SchemaVersion != UploadCheckpointSnapshotSchemaVersion {
		return UploadCheckpoint{}, ErrUnsupportedUploadCheckpointSnapshot
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointSnapshot
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, persisted.UpdatedAt)
	if err != nil {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointSnapshot
	}
	checkpoint, err := RestoreUploadCheckpoint(
		Record{
			OwnerID:     persisted.OwnerID,
			MediaID:     persisted.MediaID,
			State:       persisted.State,
			UpdatedAt:   updatedAt.UTC(),
			FailureCode: persisted.FailureCode,
			Attempts:    persisted.Attempts,
		},
		UploadProgress{
			TotalBytes:     persisted.TotalBytes,
			CommittedBytes: persisted.CommittedBytes,
		},
	)
	if err != nil {
		return UploadCheckpoint{}, fmt.Errorf("%w: %v", ErrInvalidUploadCheckpointSnapshot, err)
	}
	return checkpoint, nil
}
