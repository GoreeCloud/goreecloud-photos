package sync

import (
	"bytes"
	"errors"
	"math"
)

type UploadCheckpointRevision uint64

var (
	ErrStaleUploadCheckpointRevision     = errors.New("stale upload checkpoint revision")
	ErrUploadCheckpointRevisionExhausted = errors.New("upload checkpoint revision exhausted")
	ErrInvalidUploadCheckpointRecord     = errors.New("invalid upload checkpoint record")
)

// UploadCheckpointRecord is a storage-neutral optimistic-concurrency wrapper
// around one canonical upload-checkpoint payload. It does not select or access
// any database, filesystem, object store, or synchronization provider.
type UploadCheckpointRecord struct {
	revision UploadCheckpointRevision
	payload  []byte
}

func NewUploadCheckpointRecord(checkpoint UploadCheckpoint) (UploadCheckpointRecord, error) {
	payload, err := EncodeUploadCheckpoint(checkpoint)
	if err != nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	return UploadCheckpointRecord{revision: 0, payload: append([]byte(nil), payload...)}, nil
}

func RestoreUploadCheckpointRecord(revision UploadCheckpointRevision, payload []byte) (UploadCheckpointRecord, error) {
	if len(payload) == 0 || len(payload) > MaxUploadCheckpointPayloadBytes {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	if _, err := DecodeUploadCheckpoint(payload); err != nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	return UploadCheckpointRecord{revision: revision, payload: append([]byte(nil), payload...)}, nil
}

func (r UploadCheckpointRecord) Revision() UploadCheckpointRevision { return r.revision }

func (r UploadCheckpointRecord) Payload() []byte {
	return append([]byte(nil), r.payload...)
}

func (r UploadCheckpointRecord) Restore() (UploadCheckpoint, error) {
	checkpoint, err := DecodeUploadCheckpoint(r.payload)
	if err != nil {
		return UploadCheckpoint{}, ErrInvalidUploadCheckpointRecord
	}
	return checkpoint, nil
}

// Update replaces the checkpoint payload only when the expected revision
// matches and immutable transfer identity (owner, media, total size) is
// unchanged. Mutable lifecycle/progress fields may advance normally.
func (r UploadCheckpointRecord) Update(
	expected UploadCheckpointRevision,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	if expected != r.revision {
		return r, ErrStaleUploadCheckpointRevision
	}
	if r.revision == UploadCheckpointRevision(math.MaxUint64) {
		return r, ErrUploadCheckpointRevisionExhausted
	}
	current, err := r.Restore()
	if err != nil {
		return r, err
	}
	validated, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress)
	if err != nil {
		return r, ErrInvalidUploadCheckpointRecord
	}
	if current.record.OwnerID != validated.record.OwnerID ||
		current.record.MediaID != validated.record.MediaID ||
		current.progress.TotalBytes != validated.progress.TotalBytes {
		return r, ErrInvalidUploadCheckpointRecord
	}
	payload, err := EncodeUploadCheckpoint(validated)
	if err != nil {
		return r, ErrInvalidUploadCheckpointRecord
	}
	return UploadCheckpointRecord{
		revision: r.revision + 1,
		payload:  append([]byte(nil), payload...),
	}, nil
}

func sameUploadCheckpointRecord(left, right UploadCheckpointRecord) bool {
	return left.Revision() == right.Revision() && bytes.Equal(left.Payload(), right.Payload())
}
