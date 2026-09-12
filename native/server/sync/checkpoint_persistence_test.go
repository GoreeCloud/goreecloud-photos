package sync

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCheckpointPersistenceRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 7, 18, 30, 0, 123, time.UTC)
	original := UploadCheckpoint{record: Record{OwnerID: "owner-1", MediaID: "media-1", State: StateUploading, UpdatedAt: now, Attempts: 2}, progress: UploadProgress{TotalBytes: 100, CommittedBytes: 40}}
	data, err := EncodeUploadCheckpoint(original)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeUploadCheckpoint(data)
	if err != nil {
		t.Fatal(err)
	}
	if restored.record != original.record || restored.progress != original.progress {
		t.Fatalf("restored=%+v/%+v want=%+v/%+v", restored.record, restored.progress, original.record, original.progress)
	}
}

func TestCheckpointPersistenceRejectsUnsupportedUnknownTrailingAndNonCanonicalTime(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"schemaVersion":2,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:30:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0}`),
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:30:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0,"extra":true}`),
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T18:30:00Z","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0} {}`),
		[]byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"pending","updatedAt":"2026-09-07T13:30:00-05:00","failureCode":"","attempts":0,"totalBytes":10,"committedBytes":0}`),
	}
	for i, data := range cases {
		if _, err := DecodeUploadCheckpoint(data); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
	if _, err := DecodeUploadCheckpoint(cases[0]); !errors.Is(err, ErrUnsupportedUploadCheckpointVersion) {
		t.Fatalf("unsupported error=%v", err)
	}
}

func TestCheckpointPersistenceRejectsOversizedAndFalseSynced(t *testing.T) {
	if _, err := DecodeUploadCheckpoint([]byte(strings.Repeat("x", MaxUploadCheckpointPayloadBytes+1))); !errors.Is(err, ErrInvalidUploadCheckpointPayload) {
		t.Fatalf("oversize error=%v", err)
	}
	data := []byte(`{"schemaVersion":1,"ownerId":"o","mediaId":"m","state":"synced","updatedAt":"2026-09-07T18:30:00Z","failureCode":"","attempts":1,"totalBytes":10,"committedBytes":9}`)
	if _, err := DecodeUploadCheckpoint(data); err == nil {
		t.Fatal("expected false synced state rejection")
	}
}
