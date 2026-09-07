package media

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

var (
	ErrInvalidContentHashInput = errors.New("invalid media content hash input")
	ErrContentSizeMismatch     = errors.New("media content size mismatch")
)

// HashContentSHA256 streams exactly expectedBytes into SHA-256 and verifies the
// supplied reader contains neither fewer nor additional bytes. It does not
// interpret filesystem paths, URLs, or storage identifiers and does not buffer
// the entire media object in memory.
func HashContentSHA256(reader io.Reader, expectedBytes int64) (string, error) {
	if reader == nil || expectedBytes < 0 {
		return "", ErrInvalidContentHashInput
	}
	hasher := sha256.New()
	copied, err := io.Copy(hasher, io.LimitReader(reader, expectedBytes))
	if err != nil {
		return "", err
	}
	if copied != expectedBytes {
		return "", ErrContentSizeMismatch
	}

	var extra [1]byte
	n, err := io.ReadFull(reader, extra[:])
	if n != 0 {
		return "", ErrContentSizeMismatch
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashContentIdentity combines the streaming content digest with the existing
// owner-scoped exact-content identity. Storage authorization remains with the
// caller that supplied the reader. Invalid identity metadata is rejected before
// the reader is consumed.
func HashContentIdentity(ownerID string, expectedBytes int64, reader io.Reader) (ContentIdentity, error) {
	if ownerID == "" || strings.TrimSpace(ownerID) != ownerID {
		return ContentIdentity{}, ErrInvalidContentIdentity
	}
	digest, err := HashContentSHA256(reader, expectedBytes)
	if err != nil {
		return ContentIdentity{}, err
	}
	return NewContentIdentity(ownerID, expectedBytes, digest)
}
