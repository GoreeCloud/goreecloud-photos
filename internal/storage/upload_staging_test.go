package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
)

func TestFilesystemUploadPartIsIdempotentAndConflictSafe(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const uploadID = "01999999-0000-7000-8000-000000000010"

	first, err := store.PutUploadPart(context.Background(), uploadID, 1, strings.NewReader("abcd"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || first.Size != 4 {
		t.Fatalf("unexpected first part: %#v", first)
	}
	second, err := store.PutUploadPart(context.Background(), uploadID, 1, strings.NewReader("abcd"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created || second.SHA256 != first.SHA256 {
		t.Fatalf("unexpected idempotent part: %#v", second)
	}
	if _, err := store.PutUploadPart(context.Background(), uploadID, 1, strings.NewReader("wxyz"), 4); !errors.Is(err, upload.ErrPartConflict) {
		t.Fatalf("expected ErrPartConflict, got %v", err)
	}
}

func TestFilesystemUploadPartRequiresExactSize(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const uploadID = "01999999-0000-7000-8000-000000000011"
	if _, err := store.PutUploadPart(context.Background(), uploadID, 1, strings.NewReader("abc"), 4); !errors.Is(err, upload.ErrPartSize) {
		t.Fatalf("expected ErrPartSize, got %v", err)
	}
}
