package media

import (
	"errors"
	"strings"
	"testing"
)

func TestContentIdentityCanonicalizesSHA256AndMatchesExactOwnerContent(t *testing.T) {
	upper := strings.Repeat("AB", 32)
	left, err := NewContentIdentity("owner-1", 42, upper)
	if err != nil {
		t.Fatal(err)
	}
	right, err := NewContentIdentity("owner-1", 42, strings.ToLower(upper))
	if err != nil {
		t.Fatal(err)
	}
	if left.SHA256() != strings.ToLower(upper) || !left.Matches(right) {
		t.Fatalf("left=%+v right=%+v", left, right)
	}

	otherOwner, _ := NewContentIdentity("owner-2", 42, upper)
	otherSize, _ := NewContentIdentity("owner-1", 43, upper)
	if left.Matches(otherOwner) || left.Matches(otherSize) {
		t.Fatal("identity matched across owner or size boundary")
	}
}

func TestContentIdentityRejectsMalformedInputs(t *testing.T) {
	validDigest := strings.Repeat("a", 64)
	for name, identity := range map[string]struct {
		owner  string
		size   int64
		digest string
	}{
		"blank owner":    {owner: "", size: 1, digest: validDigest},
		"padded owner":   {owner: " owner ", size: 1, digest: validDigest},
		"negative size":  {owner: "owner-1", size: -1, digest: validDigest},
		"short digest":   {owner: "owner-1", size: 1, digest: strings.Repeat("a", 63)},
		"non hex digest": {owner: "owner-1", size: 1, digest: strings.Repeat("z", 64)},
		"padded digest":  {owner: "owner-1", size: 1, digest: " " + strings.Repeat("a", 63)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewContentIdentity(identity.owner, identity.size, identity.digest); !errors.Is(err, ErrInvalidContentIdentity) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestItemValidateRejectsNonHexSHA256(t *testing.T) {
	item := duplicateTestItem("media-1", "owner-1", 42, strings.Repeat("z", 64))
	if err := item.Validate(); err == nil {
		t.Fatal("expected malformed SHA-256 rejection")
	}
}

func TestFindDuplicateItemIDsIsOwnerScopedDeterministicAndExcludesSelf(t *testing.T) {
	digest := strings.Repeat("a", 64)
	candidate := duplicateTestItem("media-2", "owner-1", 42, strings.ToUpper(digest))
	existing := []Item{
		duplicateTestItem("media-z", "owner-1", 42, digest),
		duplicateTestItem("media-2", "owner-1", 42, digest),
		duplicateTestItem("media-a", "owner-1", 42, digest),
		duplicateTestItem("media-private", "owner-2", 42, digest),
		duplicateTestItem("media-size", "owner-1", 43, digest),
		duplicateTestItem("media-z", "owner-1", 42, digest),
	}

	matches, err := FindDuplicateItemIDs(candidate, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0] != "media-a" || matches[1] != "media-z" {
		t.Fatalf("matches=%v", matches)
	}
}

func TestFindDuplicateItemIDsFailsClosedOnInvalidCandidateOrExistingItem(t *testing.T) {
	digest := strings.Repeat("a", 64)
	candidate := duplicateTestItem("media-1", "owner-1", 42, digest)
	invalidCandidate := candidate
	invalidCandidate.SHA256 = strings.Repeat("x", 64)
	if _, err := FindDuplicateItemIDs(invalidCandidate, nil); !errors.Is(err, ErrInvalidContentIdentity) {
		t.Fatalf("candidate error=%v", err)
	}

	invalidExisting := duplicateTestItem("media-2", "owner-1", 42, strings.Repeat("x", 64))
	if _, err := FindDuplicateItemIDs(candidate, []Item{invalidExisting}); !errors.Is(err, ErrInvalidContentIdentity) {
		t.Fatalf("existing error=%v", err)
	}
}

func duplicateTestItem(id, owner string, size int64, digest string) Item {
	return Item{
		ID:          id,
		OwnerID:     owner,
		Kind:        KindPhoto,
		OriginalURI: "file:///library/" + id + ".jpg",
		Filename:    id + ".jpg",
		MIMEType:    "image/jpeg",
		SizeBytes:   size,
		SHA256:      digest,
	}
}
