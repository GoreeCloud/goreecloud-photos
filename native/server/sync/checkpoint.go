package sync

import (
	"errors"
	"time"
)

// UploadCheckpoint binds sync lifecycle state and resumable byte progress into
// one storage-neutral value. It does not persist itself or execute transport.
type UploadCheckpoint struct {
	record   Record
	progress UploadProgress
}

func NewUploadCheckpoint(ownerID, mediaID string, totalBytes int64, now time.Time) (UploadCheckpoint, error) {
	record, err := NewRecord(ownerID, mediaID, now)
	if err != nil {
		return UploadCheckpoint{}, err
	}
	progress, err := NewUploadProgress(totalBytes)
	if err != nil {
		return UploadCheckpoint{}, err
	}
	return UploadCheckpoint{record: record, progress: progress}, nil
}

func (c UploadCheckpoint) Record() Record {
	return c.record
}

func (c UploadCheckpoint) Progress() UploadProgress {
	return c.progress
}

func (c UploadCheckpoint) Transition(next State, now time.Time, failureCode string) (UploadCheckpoint, error) {
	if next == StateSynced && !c.progress.Complete() {
		return c, errors.New("upload cannot be marked synced before all bytes are committed")
	}
	record, err := c.record.Transition(next, now, failureCode)
	if err != nil {
		return c, err
	}
	c.record = record
	return c, nil
}

func (c UploadCheckpoint) NextChunk(maxBytes int64) (UploadChunk, bool, error) {
	if c.record.State != StateUploading {
		return UploadChunk{}, false, errors.New("upload chunk planning requires uploading sync state")
	}
	return c.progress.NextChunk(maxBytes)
}

func (c UploadCheckpoint) AcknowledgeChunk(chunk UploadChunk, committedBytes int64) (UploadCheckpoint, error) {
	if c.record.State != StateUploading {
		return c, errors.New("upload acknowledgement requires uploading sync state")
	}
	progress, err := c.progress.AcknowledgeChunk(chunk, committedBytes)
	if err != nil {
		return c, err
	}
	c.progress = progress
	return c, nil
}

func (c UploadCheckpoint) NextRetryAt(policy RetryPolicy) (time.Time, bool, error) {
	return policy.NextRetryAt(c.record)
}
