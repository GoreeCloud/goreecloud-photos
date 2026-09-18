package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"testing"
)

func TestFilesystemPutImmutableAndOpen(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("goreecloud photos original")
	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:])

	object, err := store.PutImmutable(context.Background(), "asset_00000001", bytes.NewReader(payload), expected)
	if err != nil {
		t.Fatal(err)
	}
	if object.SHA256 != expected {
		t.Fatalf("unexpected hash: %s", object.SHA256)
	}
	if object.Size != int64(len(payload)) {
		t.Fatalf("unexpected size: %d", object.Size)
	}

	reader, err := store.Open(context.Background(), object.Key)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	actual, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, payload) {
		t.Fatalf("stored payload mismatch: %q", actual)
	}
}

func TestFilesystemRefusesOverwrite(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	key := "asset_00000002"
	if _, err := store.PutImmutable(context.Background(), key, bytes.NewBufferString("first"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutImmutable(context.Background(), key, bytes.NewBufferString("second"), ""); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	reader, err := store.Open(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	actual, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != "first" {
		t.Fatalf("immutable object was replaced: %q", actual)
	}
}

func TestFilesystemChecksumMismatchDoesNotCommit(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	key := "asset_00000003"
	wrongHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := store.PutImmutable(context.Background(), key, bytes.NewBufferString("payload"), wrongHash); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}

	reader, err := store.Open(context.Background(), key)
	if err == nil {
		reader.Close()
		t.Fatal("checksum-mismatched object was committed")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expected object to be absent, got %v", err)
	}
}

func TestFilesystemRejectsUnsafeKey(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutImmutable(context.Background(), "../escape", bytes.NewBufferString("payload"), ""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestFilesystemProbe(t *testing.T) {
	store, err := NewFilesystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Probe(context.Background()); err != nil {
		t.Fatal(err)
	}
}
