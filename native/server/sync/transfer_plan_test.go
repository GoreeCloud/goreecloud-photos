package sync

import (
	"testing"
	"time"
)

func TestPlanTransferStepCoversPendingUploadingAndCompletion(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 20, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	step, err := PlanTransferStep(checkpoint, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{})
	if err != nil || step.Action != TransferActionStartUpload {
		t.Fatalf("pending step=%+v err=%v", step, err)
	}

	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	step, err = PlanTransferStep(
		checkpoint,
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
		time.Time{},
	)
	if err != nil || step.Action != TransferActionSendChunk || step.Chunk.Offset != 0 || step.Chunk.Length != 40 {
		t.Fatalf("uploading step=%+v err=%v", step, err)
	}

	checkpoint.progress.CommittedBytes = checkpoint.progress.TotalBytes
	step, err = PlanTransferStep(checkpoint, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{})
	if err != nil || step.Action != TransferActionMarkSynced {
		t.Fatalf("complete uploading step=%+v err=%v", step, err)
	}
	checkpoint, err = checkpoint.Transition(StateSynced, now.Add(2*time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	step, err = PlanTransferStep(checkpoint, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{})
	if err != nil || step.Action != TransferActionComplete {
		t.Fatalf("synced step=%+v err=%v", step, err)
	}
}

func TestPlanTransferStepReportsBandwidthBlock(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 20, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	step, err := PlanTransferStep(
		checkpoint,
		RetryPolicy{},
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 40},
		TransferNetworkMetered,
		time.Second,
		time.Time{},
	)
	if err != nil || step.Action != TransferActionBlocked {
		t.Fatalf("step=%+v err=%v", step, err)
	}
}

func TestPlanTransferStepSchedulesAndReleasesRetry(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 20, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateFailed, now.Add(2*time.Second), "network")
	if err != nil {
		t.Fatal(err)
	}
	policy := RetryPolicy{BaseDelay: 10 * time.Second, MaxDelay: time.Minute, MaxAttempts: 3}

	step, err := PlanTransferStep(checkpoint, policy, BandwidthPolicy{}, TransferNetworkUnmetered, 0, now.Add(5*time.Second))
	if err != nil || step.Action != TransferActionWaitRetry || !step.RetryAt.Equal(now.Add(12*time.Second)) {
		t.Fatalf("wait step=%+v err=%v", step, err)
	}
	step, err = PlanTransferStep(checkpoint, policy, BandwidthPolicy{}, TransferNetworkUnmetered, 0, now.Add(12*time.Second))
	if err != nil || step.Action != TransferActionRetryReady {
		t.Fatalf("ready step=%+v err=%v", step, err)
	}
}

func TestPlanTransferStepReportsRetryExhaustion(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 20, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Duration(1+attempt*3)*time.Second), "")
		if err != nil {
			t.Fatal(err)
		}
		checkpoint, err = checkpoint.Transition(StateFailed, now.Add(time.Duration(2+attempt*3)*time.Second), "network")
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 0 {
			checkpoint, err = checkpoint.Transition(StatePending, now.Add(3*time.Second), "")
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	step, err := PlanTransferStep(
		checkpoint,
		RetryPolicy{BaseDelay: time.Second, MaxDelay: time.Second, MaxAttempts: 2},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		now.Add(time.Minute),
	)
	if err != nil || step.Action != TransferActionExhausted {
		t.Fatalf("step=%+v err=%v attempts=%d", step, err, checkpoint.Record().Attempts)
	}
}

func TestPlanTransferStepRejectsMalformedCheckpointAndMissingRetryClock(t *testing.T) {
	if _, err := PlanTransferStep(UploadCheckpoint{}, RetryPolicy{}, BandwidthPolicy{}, TransferNetworkUnmetered, 0, time.Time{}); err == nil {
		t.Fatal("expected malformed checkpoint rejection")
	}

	now := time.Date(2026, 9, 7, 19, 20, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, _ = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	checkpoint, _ = checkpoint.Transition(StateFailed, now.Add(2*time.Second), "network")
	_, err = PlanTransferStep(
		checkpoint,
		RetryPolicy{BaseDelay: time.Second, MaxDelay: time.Second, MaxAttempts: 3},
		BandwidthPolicy{},
		TransferNetworkUnmetered,
		0,
		time.Time{},
	)
	if err == nil {
		t.Fatal("expected missing retry clock rejection")
	}
}
