package sync

import "testing"

func TestNextChunkUsesResumeOffsetAndBoundsLength(t *testing.T) {
	progress, err := NewUploadProgress(1000)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(400)
	if err != nil {
		t.Fatal(err)
	}

	chunk, ok, err := progress.NextChunk(256)
	if err != nil || !ok {
		t.Fatalf("unexpected chunk decision: chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
	if chunk.Offset != 400 || chunk.Length != 256 || chunk.EndOffset() != 656 {
		t.Fatalf("unexpected chunk: %+v", chunk)
	}
}

func TestNextChunkUsesShortFinalChunk(t *testing.T) {
	progress, err := NewUploadProgress(1000)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(900)
	if err != nil {
		t.Fatal(err)
	}

	chunk, ok, err := progress.NextChunk(256)
	if err != nil || !ok {
		t.Fatalf("unexpected chunk decision: chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
	if chunk.Offset != 900 || chunk.Length != 100 || chunk.EndOffset() != 1000 {
		t.Fatalf("unexpected final chunk: %+v", chunk)
	}
}

func TestNextChunkStopsWhenComplete(t *testing.T) {
	progress, err := NewUploadProgress(10)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(10)
	if err != nil {
		t.Fatal(err)
	}

	chunk, ok, err := progress.NextChunk(4)
	if err != nil || ok || chunk != (UploadChunk{}) {
		t.Fatalf("unexpected completed chunk decision: chunk=%+v ok=%v err=%v", chunk, ok, err)
	}
}

func TestNextChunkRejectsInvalidInputs(t *testing.T) {
	progress, err := NewUploadProgress(10)
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []int64{0, -1} {
		if _, _, err := progress.NextChunk(size); err == nil {
			t.Fatalf("expected chunk size %d to be rejected", size)
		}
	}

	invalid := UploadProgress{TotalBytes: 10, CommittedBytes: 11}
	if _, _, err := invalid.NextChunk(1); err == nil {
		t.Fatal("expected invalid stored progress to be rejected")
	}
}
