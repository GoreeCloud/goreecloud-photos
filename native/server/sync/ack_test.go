package sync

import "testing"

func TestAcknowledgeChunkSupportsPartialAndFullConfirmation(t *testing.T) {
	progress, err := NewUploadProgress(1000)
	if err != nil {
		t.Fatal(err)
	}
	chunk, ok, err := progress.NextChunk(256)
	if err != nil || !ok {
		t.Fatalf("unexpected first chunk: %+v ok=%v err=%v", chunk, ok, err)
	}

	progress, err = progress.AcknowledgeChunk(chunk, 128)
	if err != nil || progress.CommittedBytes != 128 {
		t.Fatalf("partial acknowledgement: progress=%+v err=%v", progress, err)
	}

	chunk, ok, err = progress.NextChunk(256)
	if err != nil || !ok {
		t.Fatalf("unexpected resumed chunk: %+v ok=%v err=%v", chunk, ok, err)
	}
	progress, err = progress.AcknowledgeChunk(chunk, chunk.EndOffset())
	if err != nil || progress.CommittedBytes != 384 {
		t.Fatalf("full acknowledgement: progress=%+v err=%v", progress, err)
	}
}

func TestAcknowledgeChunkRejectsStaleFutureAndMalformedScopes(t *testing.T) {
	progress, err := NewUploadProgress(100)
	if err != nil {
		t.Fatal(err)
	}
	progress, err = progress.Advance(40)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		chunk UploadChunk
		ack   int64
	}{
		{chunk: UploadChunk{Offset: 0, Length: 10}, ack: 10},
		{chunk: UploadChunk{Offset: 41, Length: 10}, ack: 50},
		{chunk: UploadChunk{Offset: 40, Length: 0}, ack: 40},
		{chunk: UploadChunk{Offset: 40, Length: 61}, ack: 100},
		{chunk: UploadChunk{Offset: 40, Length: 10}, ack: 39},
		{chunk: UploadChunk{Offset: 40, Length: 10}, ack: 51},
	}
	for _, testCase := range cases {
		if _, err := progress.AcknowledgeChunk(testCase.chunk, testCase.ack); err == nil {
			t.Fatalf("expected acknowledgement to be rejected: chunk=%+v ack=%d", testCase.chunk, testCase.ack)
		}
	}
}
