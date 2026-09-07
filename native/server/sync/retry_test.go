package sync

import (
	"testing"
	"time"
)

func TestRetryPolicyUsesCappedExponentialBackoff(t *testing.T) {
	policy := RetryPolicy{BaseDelay: 5 * time.Second, MaxDelay: 20 * time.Second, MaxAttempts: 10}
	failedAt := time.Unix(100, 0).UTC()
	for _, tc := range []struct {
		attempts int
		want     time.Duration
	}{{1, 5 * time.Second}, {2, 10 * time.Second}, {3, 20 * time.Second}, {7, 20 * time.Second}} {
		record := Record{State: StateFailed, UpdatedAt: failedAt, Attempts: tc.attempts}
		got, retry, err := policy.NextRetryAt(record)
		if err != nil {
			t.Fatal(err)
		}
		if !retry {
			t.Fatalf("attempts %d unexpectedly exhausted", tc.attempts)
		}
		if got != failedAt.Add(tc.want) {
			t.Fatalf("attempts %d retry = %v, want %v", tc.attempts, got, failedAt.Add(tc.want))
		}
	}
}

func TestRetryPolicyStopsAtAttemptLimit(t *testing.T) {
	policy := RetryPolicy{BaseDelay: time.Second, MaxDelay: time.Minute, MaxAttempts: 3}
	got, retry, err := policy.NextRetryAt(Record{State: StateFailed, UpdatedAt: time.Unix(100, 0), Attempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	if retry {
		t.Fatalf("retry = true, want false at limit; time=%v", got)
	}
	if !got.IsZero() {
		t.Fatalf("retry time = %v, want zero", got)
	}
}

func TestRetryPolicyHandlesPreflightFailure(t *testing.T) {
	policy := RetryPolicy{BaseDelay: 3 * time.Second, MaxDelay: time.Minute, MaxAttempts: 3}
	failedAt := time.Unix(100, 0).UTC()
	got, retry, err := policy.NextRetryAt(Record{State: StateFailed, UpdatedAt: failedAt})
	if err != nil {
		t.Fatal(err)
	}
	if !retry || got != failedAt.Add(3*time.Second) {
		t.Fatalf("retry=%v at=%v", retry, got)
	}
}

func TestRetryPolicyRejectsInvalidInput(t *testing.T) {
	failed := Record{State: StateFailed, UpdatedAt: time.Unix(100, 0), Attempts: 1}
	for _, p := range []RetryPolicy{{}, {BaseDelay: time.Second, MaxDelay: time.Second, MaxAttempts: 0}, {BaseDelay: 2 * time.Second, MaxDelay: time.Second, MaxAttempts: 2}} {
		if _, _, err := p.NextRetryAt(failed); err == nil {
			t.Fatalf("expected invalid policy error for %+v", p)
		}
	}
	valid := RetryPolicy{BaseDelay: time.Second, MaxDelay: time.Minute, MaxAttempts: 3}
	if _, _, err := valid.NextRetryAt(Record{State: StatePending, UpdatedAt: time.Unix(100, 0)}); err == nil {
		t.Fatal("expected non-failed state error")
	}
	if _, _, err := valid.NextRetryAt(Record{State: StateFailed}); err == nil {
		t.Fatal("expected missing timestamp error")
	}
}
