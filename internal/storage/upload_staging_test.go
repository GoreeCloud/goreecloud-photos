package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

func TestFilesystemUploadPartPersistenceAndConflict(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const uploadID = "01997d3a-0000-7000-8000-000000000100"

	part, err := store.PutUploadPart(context.Background(), uploadID, 1, bytes.NewBufferString("abcd"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !part.Created || part.Size != 4 {
		t.Fatalf("unexpected first part: %#v", part)
	}

	retried, err := store.PutUploadPart(context.Background(), uploadID, 1, bytes.NewBufferString("abcd"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Created || retried.SHA256 != part.SHA256 {
		t.Fatalf("unexpected idempotent retry: %#v", retried)
	}

	if _, err := store.PutUploadPart(context.Background(), uploadID, 1, bytes.NewBufferString("WXYZ"), 4); !errors.Is(err, ErrUploadPartConflict) {
		t.Fatalf("expected ErrUploadPartConflict, got %v", err)
	}

	reader, err := store.OpenUploadPart(context.Background(), uploadID, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	payload, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != "abcd" {
		t.Fatalf("unexpected staged payload: %q", payload)
	}
}

func TestFilesystemUploadPartRequiresExactSize(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const uploadID = "01997d3a-0000-7000-8000-000000000101"

	if _, err := store.PutUploadPart(context.Background(), uploadID, 1, bytes.NewBufferString("abc"), 4); !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("expected short ErrSizeMismatch, got %v", err)
	}
	if _, err := store.PutUploadPart(context.Background(), uploadID, 1, bytes.NewBufferString("abcde"), 4); !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("expected oversized ErrSizeMismatch, got %v", err)
	}
	if _, err := store.OpenUploadPart(context.Background(), uploadID, 1); !errors.Is(err, ErrUploadPartNotFound) {
		t.Fatalf("invalid-sized part must not be committed, got %v", err)
	}
}

func TestFilesystemDeleteUploadPartIsChecksumGuarded(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const uploadID = "01997d3a-0000-7000-8000-000000000102"

	part, err := store.PutUploadPart(context.Background(), uploadID, 2, bytes.NewBufferString("payload"), 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteUploadPart(context.Background(), uploadID, 2, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); !errors.Is(err, ErrUploadPartConflict) {
		t.Fatalf("expected checksum-guard conflict, got %v", err)
	}
	if err := store.DeleteUploadPart(context.Background(), uploadID, 2, part.SHA256); err != nil {
		t.Fatal(err)
	}
	if _, err := store.OpenUploadPart(context.Background(), uploadID, 2); !errors.Is(err, ErrUploadPartNotFound) {
		t.Fatalf("expected deleted part to be absent, got %v", err)
	}
}
