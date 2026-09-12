package sync

import (
	"errors"
	"testing"
)

func TestContentFingerprintValidationAndComparison(t *testing.T) {
	digest := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	fingerprint, err := NewContentFingerprint(123, digest)
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint.SizeBytes() != 123 || fingerprint.SHA256Hex() != digest {
		t.Fatalf("unexpected fingerprint: size=%d digest=%q", fingerprint.SizeBytes(), fingerprint.SHA256Hex())
	}

	invalid := []struct {
		size   int64
		digest string
	}{
		{-1, digest},
		{1, "ABCDEF0123456789abcdef0123456789abcdef0123456789abcdef0123456789"},
		{1, "short"},
		{1, "g123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}
	for _, candidate := range invalid {
		if _, err := NewContentFingerprint(candidate.size, candidate.digest); !errors.Is(err, ErrInvalidContentFingerprint) {
			t.Fatalf("candidate %+v error=%v", candidate, err)
		}
	}

	same, _ := NewContentFingerprint(123, digest)
	differentSize, _ := NewContentFingerprint(124, digest)
	differentDigest, _ := NewContentFingerprint(123, "1123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if !SameContent(fingerprint, same) {
		t.Fatal("expected matching canonical fingerprints")
	}
	if SameContent(fingerprint, differentSize) || SameContent(fingerprint, differentDigest) {
		t.Fatal("different fingerprints must not match")
	}
	if SameContent(ContentFingerprint{}, ContentFingerprint{}) {
		t.Fatal("invalid zero-value fingerprints must fail closed")
	}
}
