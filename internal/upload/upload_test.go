package upload

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type memoryRepository struct {
	session Session
	part    *Part
}

func (r *memoryRepository) CreateUploadSession(_ context.Context, session Session) (Session, error) {
	r.session = session
	return session, nil
}
func (r *memoryRepository) GetUploadSession(_ context.Context, actor, uploadID string) (Session, error) {
	if r.session.UploadID != uploadID || r.session.ActorSubjectID != actor {
		return Session{}, ErrNotFound
	}
	session := r.session
	if r.part != nil {
		session.Parts = []Part{*r.part}
	}
	return session, nil
}
func (r *memoryRepository) RecordUploadPart(_ context.Context, actor, uploadID string, part Part) (Session, error) {
	if r.session.UploadID != uploadID || r.session.ActorSubjectID != actor {
		return Session{}, ErrNotFound
	}
	if r.part != nil {
		if r.part.ContentSHA256 != part.ContentSHA256 || r.part.ByteSize != part.ByteSize {
			return Session{}, ErrPartConflict
		}
		return r.GetUploadSession(context.Background(), actor, uploadID)
	}
	r.part = &part
	r.session.ReceivedBytes += part.ByteSize
	r.session.State = "receiving"
	r.session.UpdatedAt = part.CreatedAt
	return r.GetUploadSession(context.Background(), actor, uploadID)
}
func (r *memoryRepository) CancelUploadSession(_ context.Context, actor, uploadID string) (Session, error) {
	if r.session.UploadID != uploadID || r.session.ActorSubjectID != actor {
		return Session{}, ErrNotFound
	}
	r.session.State = "cancelled"
	return r.session, nil
}

type memoryStaging struct{ parts map[int]string }

func (s *memoryStaging) PutUploadPart(_ context.Context, _ string, partNumber int, reader io.Reader, expectedSize int64) (StagedPart, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return StagedPart{}, err
	}
	if int64(len(data)) != expectedSize {
		return StagedPart{}, ErrPartSize
	}
	if s.parts == nil {
		s.parts = map[int]string{}
	}
	hash := SHA256(data)
	if existing, ok := s.parts[partNumber]; ok {
		if existing != hash {
			return StagedPart{}, ErrPartConflict
		}
		return StagedPart{StorageKey: "memory", SHA256: hash, Size: int64(len(data))}, nil
	}
	s.parts[partNumber] = hash
	return StagedPart{StorageKey: "memory", SHA256: hash, Size: int64(len(data)), Created: true}, nil
}
func (s *memoryStaging) RemoveUploadPart(_ context.Context, _ string, partNumber int) error {
	delete(s.parts, partNumber)
	return nil
}
func (s *memoryStaging) RemoveUploadSession(_ context.Context, _ string) error {
	s.parts = map[int]string{}
	return nil
}

func TestNewUUIDv7(t *testing.T) {
	id, err := newUUIDv7(time.UnixMilli(1789716000000))
	if err != nil {
		t.Fatal(err)
	}
	if !uuidPattern.MatchString(id) || id[14] != '7' || !strings.ContainsRune("89ab", rune(id[19])) {
		t.Fatalf("invalid UUIDv7: %q", id)
	}
}

func TestServiceRequiresExactPartSizeAndIsIdempotent(t *testing.T) {
	repository := &memoryRepository{}
	staging := &memoryStaging{}
	service := NewService(repository, staging)
	service.now = func() time.Time { return time.Date(2026, 9, 18, 7, 0, 0, 0, time.UTC) }

	session, err := service.Create(context.Background(), "subject-1", CreateRequest{
		LibraryID:        "01999999-0000-7000-8000-000000000001",
		OriginalFilename: "photo.jpg",
		MediaType:        "image/jpeg",
		ExpectedSize:     4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.PutPart(context.Background(), "subject-1", session.UploadID, 1, strings.NewReader("abc")); !errors.Is(err, ErrPartSize) {
		t.Fatalf("expected ErrPartSize, got %v", err)
	}
	first, err := service.PutPart(context.Background(), "subject-1", session.UploadID, 1, strings.NewReader("abcd"))
	if err != nil {
		t.Fatal(err)
	}
	if first.ReceivedBytes != 4 || len(first.Parts) != 1 {
		t.Fatalf("unexpected session: %#v", first)
	}
	second, err := service.PutPart(context.Background(), "subject-1", session.UploadID, 1, strings.NewReader("abcd"))
	if err != nil {
		t.Fatal(err)
	}
	if second.ReceivedBytes != 4 {
		t.Fatalf("idempotent retry changed bytes: %d", second.ReceivedBytes)
	}
	if _, err = service.PutPart(context.Background(), "subject-1", session.UploadID, 1, strings.NewReader("wxyz")); !errors.Is(err, ErrPartConflict) {
		t.Fatalf("expected ErrPartConflict, got %v", err)
	}
}
