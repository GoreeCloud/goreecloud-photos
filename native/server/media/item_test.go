package media

import (
	"strings"
	"testing"
)

func TestItemValidate(t *testing.T) {
	item := Item{
		ID: "media-1", OwnerID: "user-1", Kind: KindPhoto,
		OriginalURI: "file:///library/photo.jpg", Filename: "photo.jpg",
		MIMEType: "image/jpeg", SizeBytes: 42, SHA256: strings.Repeat("a", 64),
	}
	if err := item.Validate(); err != nil {
		t.Fatalf("expected valid item: %v", err)
	}
}

func TestItemRejectsUnsupportedKind(t *testing.T) {
	item := Item{
		ID: "media-1", OwnerID: "user-1", Kind: Kind("audio"),
		OriginalURI: "file:///library/audio.mp3", Filename: "audio.mp3",
		SizeBytes: 42, SHA256: strings.Repeat("a", 64),
	}
	if err := item.Validate(); err == nil {
		t.Fatal("expected unsupported kind error")
	}
}
