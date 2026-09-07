package sync

import (
	"errors"
	"fmt"
)

const UploadCheckpointSnapshotSchemaVersion = UploadCheckpointSchemaVersion

var (
	ErrInvalidUploadCheckpointSnapshot     = errors.New("invalid upload checkpoint snapshot")
	ErrUnsupportedUploadCheckpointSnapshot = errors.New("unsupported upload checkpoint snapshot version")
)

// EncodeUploadCheckpointSnapshot is the snapshot-facing compatibility surface
// for the canonical bounded checkpoint persistence format. Keeping this wrapper
// avoids two independently evolving schema implementations in the sync package.
func EncodeUploadCheckpointSnapshot(checkpoint UploadCheckpoint) ([]byte, error) {
	data, err := EncodeUploadCheckpoint(checkpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUploadCheckpointSnapshot, err)
	}
	return data, nil
}

// DecodeUploadCheckpointSnapshot preserves the snapshot API while delegating
// schema, payload-bound, canonical-time, and checkpoint-coherence validation to
// DecodeUploadCheckpoint.
func DecodeUploadCheckpointSnapshot(data []byte) (UploadCheckpoint, error) {
	checkpoint, err := DecodeUploadCheckpoint(data)
	if err == nil {
		return checkpoint, nil
	}
	if errors.Is(err, ErrUnsupportedUploadCheckpointVersion) {
		return UploadCheckpoint{}, ErrUnsupportedUploadCheckpointSnapshot
	}
	return UploadCheckpoint{}, fmt.Errorf("%w: %v", ErrInvalidUploadCheckpointSnapshot, err)
}
