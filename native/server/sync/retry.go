package sync

import (
	"errors"
	"time"
)

type RetryPolicy struct {
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxAttempts int
}

func (p RetryPolicy) NextRetryAt(record Record) (time.Time, bool, error) {
	if record.State != StateFailed {
		return time.Time{}, false, errors.New("retry scheduling requires failed sync state")
	}
	if record.UpdatedAt.IsZero() {
		return time.Time{}, false, errors.New("failed sync record timestamp is required")
	}
	if p.BaseDelay <= 0 || p.MaxDelay < p.BaseDelay || p.MaxAttempts <= 0 {
		return time.Time{}, false, errors.New("invalid retry policy")
	}
	if record.Attempts >= p.MaxAttempts {
		return time.Time{}, false, nil
	}

	retryOrdinal := record.Attempts
	if retryOrdinal < 1 {
		retryOrdinal = 1
	}
	delay := p.BaseDelay
	for step := 1; step < retryOrdinal && delay < p.MaxDelay; step++ {
		if delay > p.MaxDelay/2 {
			delay = p.MaxDelay
			break
		}
		delay *= 2
	}
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return record.UpdatedAt.Add(delay), true, nil
}
