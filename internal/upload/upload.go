package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultPartSize = int64(8 * 1024 * 1024)
	sessionLifetime = 24 * time.Hour
)

var (
	ErrInvalidRequest = errors.New("invalid upload request")
	ErrUnavailable    = errors.New("upload service unavailable")
	ErrNotFound       = errors.New("upload session not found")
	ErrInvalidState   = errors.New("upload session state does not allow this operation")
	ErrExpired        = errors.New("upload session expired")
	ErrPartConflict   = errors.New("upload part conflicts with an existing part")
	ErrPartSize       = errors.New("upload part size is invalid")
)

var (
	uuidPattern   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type CreateRequest struct {
	LibraryID        string
	OriginalFilename string
	MediaType        string
	ExpectedSize     int64
	ExpectedSHA256   string
	DeviceID         string
}

type Part struct {
	PartNumber    int       `json:"part_number"`
	ByteOffset    int64     `json:"byte_offset"`
	ByteSize      int64     `json:"byte_size"`
	ContentSHA256 string    `json:"content_sha256"`
	StorageKey    string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type Session struct {
	UploadID         string    `json:"upload_id"`
	LibraryID        string    `json:"library_id"`
	ActorSubjectID   string    `json:"-"`
	OriginalFilename string    `json:"original_filename"`
	MediaType        string    `json:"media_type"`
	ExpectedSize     int64     `json:"expected_size"`
	ExpectedSHA256   string    `json:"expected_sha256,omitempty"`
	ReceivedBytes    int64     `json:"received_bytes"`
	PartSize         int64     `json:"part_size"`
	State            string    `json:"state"`
	DeviceID         string    `json:"device_id,omitempty"`
	TransferMethod   string    `json:"transfer_method"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Parts            []Part    `json:"parts"`
}

type StagedPart struct {
	StorageKey string
	SHA256     string
	Size       int64
	Created    bool
}

type Repository interface {
	CreateUploadSession(context.Context, Session) (Session, error)
	GetUploadSession(context.Context, string, string) (Session, error)
	RecordUploadPart(context.Context, string, string, Part) (Session, error)
	CancelUploadSession(context.Context, string, string) (Session, error)
}

type StagingStore interface {
	PutUploadPart(context.Context, string, int, io.Reader, int64) (StagedPart, error)
	RemoveUploadPart(context.Context, string, int) error
	RemoveUploadSession(context.Context, string) error
}

type Service struct {
	repository Repository
	staging    StagingStore
	now        func() time.Time
	newID      func(time.Time) (string, error)
}

func NewService(repository Repository, staging StagingStore) *Service {
	return &Service{
		repository: repository,
		staging:    staging,
		now:        time.Now,
		newID:      newUUIDv7,
	}
}

func (s *Service) Create(ctx context.Context, actorSubjectID string, request CreateRequest) (Session, error) {
	if s == nil || s.repository == nil || s.staging == nil {
		return Session{}, ErrUnavailable
	}

	actorSubjectID = strings.TrimSpace(actorSubjectID)
	request.LibraryID = strings.ToLower(strings.TrimSpace(request.LibraryID))
	request.OriginalFilename = strings.TrimSpace(request.OriginalFilename)
	request.MediaType = strings.TrimSpace(request.MediaType)
	request.ExpectedSHA256 = strings.ToLower(strings.TrimSpace(request.ExpectedSHA256))
	request.DeviceID = strings.TrimSpace(request.DeviceID)

	if actorSubjectID == "" ||
		!uuidPattern.MatchString(request.LibraryID) ||
		request.OriginalFilename == "" || len(request.OriginalFilename) > 1024 ||
		request.MediaType == "" || len(request.MediaType) > 255 ||
		request.ExpectedSize < 0 ||
		len(request.DeviceID) > 256 {
		return Session{}, ErrInvalidRequest
	}
	if request.ExpectedSHA256 != "" && !sha256Pattern.MatchString(request.ExpectedSHA256) {
		return Session{}, ErrInvalidRequest
	}

	now := s.now().UTC()
	uploadID, err := s.newID(now)
	if err != nil {
		return Session{}, ErrUnavailable
	}

	session := Session{
		UploadID:         uploadID,
		LibraryID:        request.LibraryID,
		ActorSubjectID:   actorSubjectID,
		OriginalFilename: request.OriginalFilename,
		MediaType:        request.MediaType,
		ExpectedSize:     request.ExpectedSize,
		ExpectedSHA256:   request.ExpectedSHA256,
		PartSize:         DefaultPartSize,
		State:            "open",
		DeviceID:         request.DeviceID,
		TransferMethod:   "parts-v1",
		ExpiresAt:        now.Add(sessionLifetime),
		CreatedAt:        now,
		UpdatedAt:        now,
		Parts:            []Part{},
	}

	return s.repository.CreateUploadSession(ctx, session)
}

func (s *Service) Get(ctx context.Context, actorSubjectID, uploadID string) (Session, error) {
	if s == nil || s.repository == nil {
		return Session{}, ErrUnavailable
	}
	actorSubjectID = strings.TrimSpace(actorSubjectID)
	uploadID = strings.ToLower(strings.TrimSpace(uploadID))
	if actorSubjectID == "" || !uuidPattern.MatchString(uploadID) {
		return Session{}, ErrInvalidRequest
	}
	return s.repository.GetUploadSession(ctx, actorSubjectID, uploadID)
}

func (s *Service) PutPart(ctx context.Context, actorSubjectID, uploadID string, partNumber int, body io.Reader) (Session, error) {
	if s == nil || s.repository == nil || s.staging == nil {
		return Session{}, ErrUnavailable
	}
	if body == nil || partNumber < 1 {
		return Session{}, ErrInvalidRequest
	}

	session, err := s.Get(ctx, actorSubjectID, uploadID)
	if err != nil {
		return Session{}, err
	}
	if session.State != "open" && session.State != "receiving" {
		return Session{}, ErrInvalidState
	}
	if !session.ExpiresAt.After(s.now().UTC()) {
		return Session{}, ErrExpired
	}
	if session.ExpectedSize == 0 {
		return Session{}, ErrPartSize
	}

	partIndex := int64(partNumber - 1)
	if session.PartSize <= 0 || partIndex > math.MaxInt64/session.PartSize {
		return Session{}, ErrInvalidRequest
	}
	offset := partIndex * session.PartSize
	if offset >= session.ExpectedSize {
		return Session{}, ErrPartSize
	}

	expectedPartSize := session.PartSize
	if remaining := session.ExpectedSize - offset; remaining < expectedPartSize {
		expectedPartSize = remaining
	}

	data, err := io.ReadAll(io.LimitReader(body, expectedPartSize+1))
	if err != nil {
		return Session{}, ErrUnavailable
	}
	if int64(len(data)) != expectedPartSize {
		return Session{}, ErrPartSize
	}

	staged, err := s.staging.PutUploadPart(ctx, session.UploadID, partNumber, bytes.NewReader(data), expectedPartSize)
	if err != nil {
		return Session{}, err
	}

	part := Part{
		PartNumber:    partNumber,
		ByteOffset:    offset,
		ByteSize:      staged.Size,
		ContentSHA256: staged.SHA256,
		StorageKey:    staged.StorageKey,
		CreatedAt:     s.now().UTC(),
	}

	updated, err := s.repository.RecordUploadPart(ctx, actorSubjectID, session.UploadID, part)
	if err != nil {
		if staged.Created {
			_ = s.staging.RemoveUploadPart(context.Background(), session.UploadID, partNumber)
		}
		return Session{}, err
	}
	return updated, nil
}

func (s *Service) Cancel(ctx context.Context, actorSubjectID, uploadID string) (Session, error) {
	if s == nil || s.repository == nil || s.staging == nil {
		return Session{}, ErrUnavailable
	}

	session, err := s.repository.CancelUploadSession(ctx, strings.TrimSpace(actorSubjectID), strings.ToLower(strings.TrimSpace(uploadID)))
	if err != nil {
		return Session{}, err
	}
	if err := s.staging.RemoveUploadSession(ctx, session.UploadID); err != nil {
		return session, ErrUnavailable
	}
	return session, nil
}

func newUUIDv7(now time.Time) (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	milliseconds := now.UnixMilli()
	if milliseconds < 0 || milliseconds > (1<<48)-1 {
		return "", fmt.Errorf("timestamp outside UUIDv7 range")
	}

	value[0] = byte(milliseconds >> 40)
	value[1] = byte(milliseconds >> 32)
	value[2] = byte(milliseconds >> 24)
	value[3] = byte(milliseconds >> 16)
	value[4] = byte(milliseconds >> 8)
	value[5] = byte(milliseconds)
	value[6] = (value[6] & 0x0f) | 0x70
	value[8] = (value[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(value[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}

func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
