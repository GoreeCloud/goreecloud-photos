package media

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestHashContentSHA256StreamsExactBytes(t *testing.T) {
	payload := []byte("GoreeCloud exact media content")
	digest, err := HashContentSHA256(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatal(err)
	}
	const expected = "ee6d10383aa9c221e58c1ed0b78fc059a4cd863459c67a9ce7dc9d5dbef97a0e"
	if digest != expected {
		t.Fatalf("digest=%q", digest)
	}
	if canonical, ok := canonicalSHA256Digest(digest); !ok || canonical != digest {
		t.Fatalf("digest was not canonical: %q", digest)
	}
}

func TestHashContentSHA256SupportsEmptyContent(t *testing.T) {
	digest, err := HashContentSHA256(bytes.NewReader(nil), 0)
	if err != nil {
		t.Fatal(err)
	}
	if digest != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("digest=%q", digest)
	}
}

func TestHashContentSHA256RejectsShortAndLongReaders(t *testing.T) {
	if _, err := HashContentSHA256(strings.NewReader("abc"), 4); !errors.Is(err, ErrContentSizeMismatch) {
		t.Fatalf("short reader error=%v", err)
	}
	if _, err := HashContentSHA256(strings.NewReader("abcd"), 3); !errors.Is(err, ErrContentSizeMismatch) {
		t.Fatalf("long reader error=%v", err)
	}
}

func TestHashContentSHA256RejectsInvalidInputAndPropagatesReadFailure(t *testing.T) {
	if _, err := HashContentSHA256(nil, 0); !errors.Is(err, ErrInvalidContentHashInput) {
		t.Fatalf("nil reader error=%v", err)
	}
	if _, err := HashContentSHA256(strings.NewReader(""), -1); !errors.Is(err, ErrInvalidContentHashInput) {
		t.Fatalf("negative size error=%v", err)
	}
	want := errors.New("read failed")
	if _, err := HashContentSHA256(errorReader{err: want}, 1); !errors.Is(err, want) {
		t.Fatalf("read failure=%v", err)
	}
}

func TestHashContentIdentityBuildsOwnerScopedIdentity(t *testing.T) {
	payload := []byte("same bytes")
	left, err := HashContentIdentity("owner-1", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	right, err := HashContentIdentity("owner-1", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	otherOwner, err := HashContentIdentity("owner-2", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if !left.Matches(right) || left.Matches(otherOwner) || left.SizeBytes() != int64(len(payload)) {
		t.Fatalf("left=%+v right=%+v other=%+v", left, right, otherOwner)
	}
}

func TestHashContentIdentityRejectsInvalidOwnerBeforeReading(t *testing.T) {
	payload := []byte("content")
	reader := bytes.NewReader(payload)
	if _, err := HashContentIdentity(" owner ", int64(len(payload)), reader); !errors.Is(err, ErrInvalidContentIdentity) {
		t.Fatalf("owner error=%v", err)
	}
	if reader.Len() != len(payload) {
		t.Fatalf("invalid owner consumed %d bytes", len(payload)-reader.Len())
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

var _ io.Reader = errorReader{}
