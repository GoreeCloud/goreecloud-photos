package sync

import (
	"encoding/hex"
	"errors"
	"strings"
)

const SHA256HexLength = 64

var ErrInvalidContentFingerprint = errors.New("invalid content fingerprint")

// ContentFingerprint is a storage-neutral exact-content comparison value. The
// sync core does not open files or compute hashes here; callers must provide a
// previously authorized byte count and SHA-256 digest.
type ContentFingerprint struct {
	sizeBytes int64
	sha256Hex string
}

func NewContentFingerprint(sizeBytes int64, sha256Hex string) (ContentFingerprint, error) {
	if sizeBytes < 0 || len(sha256Hex) != SHA256HexLength || sha256Hex != strings.ToLower(sha256Hex) {
		return ContentFingerprint{}, ErrInvalidContentFingerprint
	}
	if _, err := hex.DecodeString(sha256Hex); err != nil {
		return ContentFingerprint{}, ErrInvalidContentFingerprint
	}
	return ContentFingerprint{sizeBytes: sizeBytes, sha256Hex: sha256Hex}, nil
}

func (f ContentFingerprint) SizeBytes() int64  { return f.sizeBytes }
func (f ContentFingerprint) SHA256Hex() string { return f.sha256Hex }

func (f ContentFingerprint) valid() bool {
	if f.sizeBytes < 0 || len(f.sha256Hex) != SHA256HexLength || f.sha256Hex != strings.ToLower(f.sha256Hex) {
		return false
	}
	_, err := hex.DecodeString(f.sha256Hex)
	return err == nil
}

// SameContent requires both canonical fingerprints, the same byte count, and
// the same SHA-256 digest. Invalid/zero values fail closed rather than matching.
func SameContent(left, right ContentFingerprint) bool {
	return left.valid() && right.valid() && left.sizeBytes == right.sizeBytes && left.sha256Hex == right.sha256Hex
}
