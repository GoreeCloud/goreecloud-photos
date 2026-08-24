package sync

import (
	"errors"
	"strings"
	"time"
)

type State string

const (
	StatePending   State = "pending"
	StateUploading State = "uploading"
	StateSynced    State = "synced"
	StateFailed    State = "failed"
)

type Record struct {
	OwnerID     string
	MediaID     string
	State       State
	UpdatedAt   time.Time
	FailureCode string
}

func NewRecord(ownerID, mediaID string, now time.Time) (Record, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(mediaID) == "" {
		return Record{}, errors.New("owner id and media id are required")
	}
	return Record{OwnerID: ownerID, MediaID: mediaID, State: StatePending, UpdatedAt: now.UTC()}, nil
}

func (r Record) Transition(next State, now time.Time, failureCode string) (Record, error) {
	if !allowed(r.State, next) {
		return Record{}, errors.New("invalid sync state transition")
	}
	failureCode = strings.TrimSpace(failureCode)
	if next == StateFailed && failureCode == "" {
		return Record{}, errors.New("failed sync state requires a failure code")
	}
	if next != StateFailed {
		failureCode = ""
	}
	r.State = next
	r.UpdatedAt = now.UTC()
	r.FailureCode = failureCode
	return r, nil
}

func allowed(current, next State) bool {
	switch current {
	case StatePending:
		return next == StateUploading || next == StateFailed
	case StateUploading:
		return next == StateSynced || next == StateFailed
	case StateFailed:
		return next == StatePending
	case StateSynced:
		return next == StatePending
	default:
		return false
	}
}
