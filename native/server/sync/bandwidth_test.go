package sync

import (
	"math"
	"testing"
	"time"
)

func TestBandwidthPolicyComputesWindowBudget(t *testing.T) {
	policy := BandwidthPolicy{BytesPerSecond: 1024 * 1024, MaxChunkBytes: 1024 * 1024, AllowMetered: true}
	limit, allowed, err := policy.ChunkLimit(TransferNetworkUnmetered, 250*time.Millisecond)
	if err != nil || !allowed || limit != 256*1024 {
		t.Fatalf("limit=%d allowed=%v err=%v", limit, allowed, err)
	}
}

func TestBandwidthPolicyCapsChunkSize(t *testing.T) {
	policy := BandwidthPolicy{BytesPerSecond: 10_000, MaxChunkBytes: 512, AllowMetered: true}
	limit, allowed, err := policy.ChunkLimit(TransferNetworkUnmetered, time.Second)
	if err != nil || !allowed || limit != 512 {
		t.Fatalf("limit=%d allowed=%v err=%v", limit, allowed, err)
	}
}

func TestBandwidthPolicyDeniesMeteredWhenDisabled(t *testing.T) {
	policy := BandwidthPolicy{BytesPerSecond: 1024, MaxChunkBytes: 1024}
	limit, allowed, err := policy.ChunkLimit(TransferNetworkMetered, time.Second)
	if err != nil || allowed || limit != 0 {
		t.Fatalf("limit=%d allowed=%v err=%v", limit, allowed, err)
	}
}

func TestBandwidthPolicyRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		policy  BandwidthPolicy
		network TransferNetwork
		window  time.Duration
	}{
		{BandwidthPolicy{}, TransferNetworkUnmetered, time.Second},
		{BandwidthPolicy{BytesPerSecond: 1, MaxChunkBytes: 1}, TransferNetwork("satellite"), time.Second},
		{BandwidthPolicy{BytesPerSecond: 1, MaxChunkBytes: 1}, TransferNetworkUnmetered, 0},
	}
	for i, tc := range cases {
		if _, _, err := tc.policy.ChunkLimit(tc.network, tc.window); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
}

func TestBandwidthPolicyReturnsNoBudgetForSubByteWindow(t *testing.T) {
	policy := BandwidthPolicy{BytesPerSecond: 1, MaxChunkBytes: 1, AllowMetered: true}
	limit, allowed, err := policy.ChunkLimit(TransferNetworkUnmetered, 500*time.Millisecond)
	if err != nil || allowed || limit != 0 {
		t.Fatalf("limit=%d allowed=%v err=%v", limit, allowed, err)
	}
}

func TestBandwidthPolicySaturatesBeforeChunkCap(t *testing.T) {
	policy := BandwidthPolicy{BytesPerSecond: math.MaxInt64, MaxChunkBytes: 4096, AllowMetered: true}
	limit, allowed, err := policy.ChunkLimit(TransferNetworkUnmetered, time.Duration(math.MaxInt64))
	if err != nil || !allowed || limit != 4096 {
		t.Fatalf("limit=%d allowed=%v err=%v", limit, allowed, err)
	}
}

func TestUploadCheckpointPlansBandwidthBoundedChunk(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 1000, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}

	chunk, ok, err := checkpoint.NextChunkWithBandwidth(
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 80, AllowMetered: true},
		TransferNetworkUnmetered,
		time.Second,
	)
	if err != nil || !ok || chunk.Offset != 0 || chunk.Length != 80 {
		t.Fatalf("chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
}

func TestUploadCheckpointRejectsInvalidLifecycleBeforeBandwidthDenial(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 1000, now)
	if err != nil {
		t.Fatal(err)
	}

	_, ok, err := checkpoint.NextChunkWithBandwidth(
		BandwidthPolicy{BytesPerSecond: 100, MaxChunkBytes: 80, AllowMetered: false},
		TransferNetworkMetered,
		time.Second,
	)
	if err == nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
