package media

import (
	"encoding/hex"
	"errors"
	"sort"
	"strings"
)

var ErrInvalidContentIdentity = errors.New("invalid media content identity")

type ContentIdentity struct {
	ownerID   string
	sizeBytes int64
	sha256    string
}

func NewContentIdentity(ownerID string, sizeBytes int64, sha256 string) (ContentIdentity, error) {
	if ownerID == "" || strings.TrimSpace(ownerID) != ownerID || sizeBytes < 0 {
		return ContentIdentity{}, ErrInvalidContentIdentity
	}
	canonicalDigest, ok := canonicalSHA256Digest(sha256)
	if !ok {
		return ContentIdentity{}, ErrInvalidContentIdentity
	}
	return ContentIdentity{ownerID: ownerID, sizeBytes: sizeBytes, sha256: canonicalDigest}, nil
}

func ContentIdentityForItem(item Item) (ContentIdentity, error) {
	if err := item.Validate(); err != nil {
		return ContentIdentity{}, ErrInvalidContentIdentity
	}
	return NewContentIdentity(item.OwnerID, item.SizeBytes, item.SHA256)
}

func (i ContentIdentity) OwnerID() string { return i.ownerID }
func (i ContentIdentity) SizeBytes() int64 { return i.sizeBytes }
func (i ContentIdentity) SHA256() string { return i.sha256 }

func (i ContentIdentity) Matches(other ContentIdentity) bool {
	return validContentIdentity(i) && validContentIdentity(other) &&
		i.ownerID == other.ownerID &&
		i.sizeBytes == other.sizeBytes &&
		i.sha256 == other.sha256
}

// FindDuplicateItemIDs returns deterministic same-owner exact-content matches.
// It never exposes cross-owner matches and ignores the candidate's own item ID.
func FindDuplicateItemIDs(candidate Item, existing []Item) ([]string, error) {
	candidateIdentity, err := ContentIdentityForItem(candidate)
	if err != nil {
		return nil, err
	}
	matches := make([]string, 0)
	seenIDs := make(map[string]struct{})
	for _, item := range existing {
		identity, err := ContentIdentityForItem(item)
		if err != nil {
			return nil, ErrInvalidContentIdentity
		}
		if item.ID == candidate.ID || !candidateIdentity.Matches(identity) {
			continue
		}
		if _, duplicate := seenIDs[item.ID]; duplicate {
			continue
		}
		seenIDs[item.ID] = struct{}{}
		matches = append(matches, item.ID)
	}
	sort.Strings(matches)
	return matches, nil
}

func canonicalSHA256Digest(value string) (string, bool) {
	if len(value) != 64 || strings.TrimSpace(value) != value {
		return "", false
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return "", false
	}
	return strings.ToLower(value), true
}

func validContentIdentity(identity ContentIdentity) bool {
	if identity.ownerID == "" || strings.TrimSpace(identity.ownerID) != identity.ownerID || identity.sizeBytes < 0 {
		return false
	}
	canonical, ok := canonicalSHA256Digest(identity.sha256)
	return ok && canonical == identity.sha256
}
