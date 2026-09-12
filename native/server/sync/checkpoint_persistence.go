package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"
)

const (
	UploadCheckpointSchemaVersion   = 1
	MaxUploadCheckpointPayloadBytes = 64 * 1024
)

var (
	ErrInvalidUploadCheckpointPayload     = errors.New("invalid upload checkpoint payload")
	ErrUnsupportedUploadCheckpointVersion = errors.New("unsupported upload checkpoint version")
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

// EncodeUploadCheckpoint emits a canonical, versioned persistence payload for
// an internally coherent upload checkpoint. It does not choose storage.
func EncodeUploadCheckpoint(checkpoint UploadCheckpoint) ([]byte, error) {
	validated, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress)
	if err != nil {
		return nil, err
	}
	record := validated.record
	return json.Marshal(persistedUploadCheckpoint{
		SchemaVersion:  UploadCheckpointSchemaVersion,
		OwnerID:        record.OwnerID,
		MediaID:        record.MediaID,
		State:          record.State,
		UpdatedAt:      record.UpdatedAt.UTC().Format(time.RFC3339Nano),
		FailureCode:    record.FailureCode,
		Attempts:       record.Attempts,
		TotalBytes:     validated.progress.TotalBytes,
		CommittedBytes: validated.progress.CommittedBytes,
	})
}

// DecodeUploadCheckpoint accepts only the canonical schema emitted by
// EncodeUploadCheckpoint and routes reconstructed state through the existing
// checkpoint restoration boundary.
func DecodeUploadCheckpoint(data []byte) (UploadCheckpoint, error) {
	if len(data) == 0 || len(data) > MaxUploadCheckpointPayloadBytes {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointPayload
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var persisted persistedUploadCheckpoint
	if err := decoder.Decode(&persisted); err != nil {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointPayload
	}
	if persisted.SchemaVersion != UploadCheckpointSchemaVersion {
		return UploadCheckpoint{}, ErrUnsupportedUploadCheckpointVersion
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointPayload
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, persisted.UpdatedAt)
	if err != nil || persisted.UpdatedAt != updatedAt.UTC().Format(time.RFC3339Nano) {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointPayload
	}

	return RestoreUploadCheckpoint(
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
}
