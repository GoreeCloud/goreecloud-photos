package sync

import (
	"testing"
	"time"
)

func TestApplyTransferStepStartsAcknowledgesAndFailsWithPartialProgress(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}

	checkpoint, err = ApplyTransferStep(
		checkpoint,
		TransferStep{Action: TransferActionStartUpload},
		now.Add(time.Second),
		TransferResult{},
	)
	if err != nil || checkpoint.Record().State != StateUploading || checkpoint.Record().Attempts != 1 {
		t.Fatalf("checkpoint=%+v err=%v", checkpoint.Record(), err)
	}

	step, err := PlanTransferStep(
		checkpoint,
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if err != nil || step.Action != TransferActionSendChunk {
		t.Fatalf("step=%+v err=%v", step, err)
	}
	checkpoint, err = ApplyTransferStep(checkpoint, step, time.Time{}, TransferResult{CommittedBytes: 20})
	if err != nil || checkpoint.Progress().CommittedBytes != 20 {
		t.Fatalf("progress=%+v err=%v", checkpoint.Progress(), err)
	}

	step, err = PlanTransferStep(
		checkpoint,
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if err != nil || step.Chunk.Offset != 20 {
		t.Fatalf("step=%+v err=%v", step, err)
	}
	checkpoint, err = ApplyTransferStep(
		checkpoint,
		step,
		now.Add(2*time.Second),
		TransferResult{CommittedBytes: 30, FailureCode: "transport-timeout"},
	)
	if err != nil || checkpoint.Record().State != StateFailed || checkpoint.Progress().CommittedBytes != 30 || checkpoint.Record().FailureCode != "transport-timeout" {
		t.Fatalf("record=%+v progress=%+v err=%v", checkpoint.Record(), checkpoint.Progress(), err)
	}
}

func TestApplyTransferStepHonorsRetryClock(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, _ = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	checkpoint, _ = checkpoint.Transition(StateFailed, now.Add(2*time.Second), "network")
	policy := RetryPolicy{BaseDelay: 10 * time.Second, MaxDelay: time.Minute, MaxAttempts: 3}

	wait, err := PlanTransferStep(checkpoint, policy, BandwidthPolicy{}, TransferNetworkUnmetered, 0, now.Add(5*time.Second))
	if err != nil || wait.Action != TransferActionWaitRetry {
		t.Fatalf("wait=%+v err=%v", wait, err)
	}
	unchanged, err := ApplyTransferStep(checkpoint, wait, now.Add(5*time.Second), TransferResult{})
	if err != nil || unchanged.Record().State != StateFailed {
		t.Fatalf("record=%+v err=%v", unchanged.Record(), err)
	}

	ready, err := PlanTransferStep(checkpoint, policy, BandwidthPolicy{}, TransferNetworkUnmetered, 0, wait.RetryAt)
	if err != nil || ready.Action != TransferActionRetryReady {
		t.Fatalf("ready=%+v err=%v", ready, err)
	}
	if _, err := ApplyTransferStep(checkpoint, ready, wait.RetryAt.Add(-time.Nanosecond), TransferResult{}); err == nil {
		t.Fatal("expected early retry rejection")
	}
	checkpoint, err = ApplyTransferStep(checkpoint, ready, wait.RetryAt, TransferResult{})
	if err != nil || checkpoint.Record().State != StatePending || checkpoint.Record().FailureCode != "" {
		t.Fatalf("record=%+v err=%v", checkpoint.Record(), err)
	}
}

func TestApplyTransferStepMarksSyncedAndAcceptsCompleteNoop(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 10, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, _ = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	checkpoint.progress.CommittedBytes = 10

	mark, err := PlanTransferStep(checkpoint, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{})
	if err != nil || mark.Action != TransferActionMarkSynced {
		t.Fatalf("mark=%+v err=%v", mark, err)
	}
	checkpoint, err = ApplyTransferStep(checkpoint, mark, now.Add(2*time.Second), TransferResult{})
	if err != nil || checkpoint.Record().State != StateSynced {
		t.Fatalf("record=%+v err=%v", checkpoint.Record(), err)
	}

	complete, err := PlanTransferStep(checkpoint, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{})
	if err != nil || complete.Action != TransferActionComplete {
		t.Fatalf("complete=%+v err=%v", complete, err)
	}
	unchanged, err := ApplyTransferStep(checkpoint, complete, time.Time{}, TransferResult{})
	if err != nil || unchanged.Record().State != StateSynced {
		t.Fatalf("record=%+v err=%v", unchanged.Record(), err)
	}
}

func TestApplyTransferStepRejectsStaleOrMalformedResults(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, _ = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	chunk, ok, err := checkpoint.NextChunk(20)
	if err != nil || !ok {
		t.Fatalf("chunk=%+v ok=%v err=%v", chunk, ok, err)
	}

	if _, err := ApplyTransferStep(
		checkpoint,
		TransferStep{Action: TransferActionSendChunk, Chunk: chunk},
		time.Time{},
		TransferResult{CommittedBytes: chunk.Offset},
	); err == nil {
		t.Fatal("expected non-advancing success rejection")
	}
	if _, err := ApplyTransferStep(
		checkpoint,
		TransferStep{Action: TransferActionSendChunk, Chunk: chunk},
		now.Add(2*time.Second),
		TransferResult{CommittedBytes: chunk.Offset, FailureCode: " timeout "},
	); err == nil {
		t.Fatal("expected noncanonical failure-code rejection")
	}
	if _, err := ApplyTransferStep(
		checkpoint,
		TransferStep{Action: TransferActionStartUpload},
		now.Add(2*time.Second),
		TransferResult{},
	); err == nil {
		t.Fatal("expected stale start-upload rejection")
	}
}
