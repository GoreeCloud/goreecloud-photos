package sync

import (
	"errors"
	"testing"
	"time"
)

func TestUploadCheckpointSnapshotRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 0, 0, 123, time.UTC)
	checkpoint, err := NewUploadCheckpoint("owner", "media", 100, now)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err = checkpoint.Transition(StateUploading, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.progress.CommittedBytes = 40

	data, err := EncodeUploadCheckpointSnapshot(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeUploadCheckpointSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Record() != checkpoint.Record() || restored.Progress() != checkpoint.Progress() {
		t.Fatalf("round trip mismatch: restored=%+v checkpoint=%+v", restored, checkpoint)
	}
}

func TestUploadCheckpointSnapshotRejectsUnsupportedVersion(t *testing.T) {
	_, err := DecodeUploadCheckpointSnapshot([]byte(`{"schemaVersion":2,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:00:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0}`))
	if !errors.Is(err, ErrUnsupportedUploadCheckpointSnapshot) {
		t.Fatalf("error = %v, want unsupported snapshot version", err)
	}
}

func TestUploadCheckpointSnapshotRejectsUnknownTrailingAndInvalidState(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:00:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0,"extra":true}`),
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:00:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0} {}`),
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"synced","updatedAt":"2026-09-07T18:00:00Z","failureCode":"","attempts":1,"totalBytes":10,"committedBytes":9}`),
	}
	for _, data := range cases {
		if _, err := DecodeUploadCheckpointSnapshot(data); !errors.Is(err, ErrInvalidUploadCheckpointSnapshot) {
			t.Fatalf("payload %s error = %v, want invalid snapshot", data, err)
		}
	}
}
