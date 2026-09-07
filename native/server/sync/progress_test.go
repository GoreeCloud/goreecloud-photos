package sync

import "testing"

func TestUploadProgressAdvancesMonotonically(t *testing.T) {
	progress, err := NewUploadProgress(1000)
	if err != nil {
		t.Fatal(err)
	}
	if progress.ResumeOffset() != 0 || progress.RemainingBytes() != 1000 || progress.Complete() {
		t.Fatalf("unexpected initial progress: %+v", progress)
	}

	progress, err = progress.Advance(400)
	if err != nil {
		t.Fatal(err)
	}
	if progress.ResumeOffset() != 400 || progress.RemainingBytes() != 600 || progress.Complete() {
		t.Fatalf("unexpected partial progress: %+v", progress)
	}

	progress, err = progress.Advance(1000)
	if err != nil {
		t.Fatal(err)
	}
	if progress.ResumeOffset() != 1000 || progress.RemainingBytes() != 0 || !progress.Complete() {
		t.Fatalf("unexpected completed progress: %+v", progress)
	}
}

func TestUploadProgressAdvanceIsIdempotentAtCurrentOffset(t *testing.T) {
	progress, err := NewUploadProgress(100)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(40)
	if err != nil {
		t.Fatal(err)
	}
	again, err := progress.Advance(40)
	if err != nil {
		t.Fatal(err)
	}
	if again != progress {
		t.Fatalf("idempotent advance changed progress: got %+v want %+v", again, progress)
	}
}

func TestUploadProgressRejectsBackwardOrOversizedOffsets(t *testing.T) {
	progress, err := NewUploadProgress(100)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(40)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := progress.Advance(39); err == nil {
		t.Fatal("expected backward progress to be rejected")
	}
	if _, err := progress.Advance(101); err == nil {
		t.Fatal("expected progress beyond total bytes to be rejected")
	}
}

func TestUploadProgressRejectsInvalidState(t *testing.T) {
	for _, total := range []int64{0, -1} {
		if _, err := NewUploadProgress(total); err == nil {
			t.Fatalf("expected total bytes %d to be rejected", total)
		}
	}

	invalid := UploadProgress{TotalBytes: 100, CommittedBytes: 101}
	if _, err := invalid.Advance(101); err == nil {
		t.Fatal("expected invalid stored progress state to be rejected")
	}
	if invalid.ResumeOffset() != 0 {
		t.Fatalf("invalid progress resume offset = %d, want 0", invalid.ResumeOffset())
	}
	if invalid.RemainingBytes() != 0 {
		t.Fatalf("invalid progress remaining bytes = %d, want 0", invalid.RemainingBytes())
	}
	if invalid.Complete() {
		t.Fatal("invalid progress must not report complete")
	}
}
