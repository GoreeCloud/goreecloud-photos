package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	objectKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{7,127}$`)
	sha256Pattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type Filesystem struct {
	root          string
	originals     string
	staging       string
	uploadStaging string
}

func NewFilesystem(root string) (*Filesystem, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("storage root is required")
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}

	store := &Filesystem{
		root:          absoluteRoot,
		originals:     filepath.Join(absoluteRoot, "originals"),
		staging:       filepath.Join(absoluteRoot, "staging"),
		uploadStaging: filepath.Join(absoluteRoot, "upload-staging"),
	}

	for _, path := range []string{store.root, store.originals, store.staging, store.uploadStaging} {
		if err := os.MkdirAll(path, 0o750); err != nil {
			return nil, fmt.Errorf("create storage directory %q: %w", path, err)
		}
	}

	return store, nil
}

func (s *Filesystem) PutImmutable(ctx context.Context, key string, source io.Reader, expectedSHA256 string) (Object, error) {
	if err := ctx.Err(); err != nil {
		return Object{}, err
	}
	if !objectKeyPattern.MatchString(key) {
		return Object{}, ErrInvalidKey
	}

	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	if expectedSHA256 != "" && !sha256Pattern.MatchString(expectedSHA256) {
		return Object{}, fmt.Errorf("%w: expected SHA-256 must be 64 lowercase hexadecimal characters", ErrChecksumMismatch)
	}

	finalPath := s.objectPath(key)
	if _, err := os.Lstat(finalPath); err == nil {
		return Object{}, ErrAlreadyExists
	} else if !os.IsNotExist(err) {
		return Object{}, fmt.Errorf("inspect destination: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(finalPath), 0o750); err != nil {
		return Object{}, fmt.Errorf("create object directory: %w", err)
	}

	temp, err := os.CreateTemp(s.staging, "original-*")
	if err != nil {
		return Object{}, fmt.Errorf("create staging file: %w", err)
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temp.Close()
			_ = os.Remove(tempName)
		}
	}()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(temp, hasher), &contextReader{ctx: ctx, reader: source})
	if err != nil {
		return Object{}, fmt.Errorf("write staging object: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return Object{}, fmt.Errorf("sync staging object: %w", err)
	}
	if err := temp.Close(); err != nil {
		return Object{}, fmt.Errorf("close staging object: %w", err)
	}

	actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
	if expectedSHA256 != "" && actualSHA256 != expectedSHA256 {
		return Object{}, fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expectedSHA256, actualSHA256)
	}

	if err := os.Chmod(tempName, 0o640); err != nil {
		return Object{}, fmt.Errorf("set original object permissions: %w", err)
	}

	if err := os.Link(tempName, finalPath); err != nil {
		if os.IsExist(err) {
			return Object{}, ErrAlreadyExists
		}
		return Object{}, fmt.Errorf("commit immutable object: %w", err)
	}

	if err := syncDirectory(filepath.Dir(finalPath)); err != nil {
		_ = os.Remove(finalPath)
		return Object{}, fmt.Errorf("sync object directory: %w", err)
	}

	if err := os.Remove(tempName); err != nil {
		_ = os.Remove(finalPath)
		return Object{}, fmt.Errorf("remove staging link: %w", err)
	}
	committed = true

	return Object{
		Key:    key,
		SHA256: actualSHA256,
		Size:   written,
	}, nil
}

func (s *Filesystem) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !objectKeyPattern.MatchString(key) {
		return nil, ErrInvalidKey
	}

	path := s.objectPath(key)
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("object is not a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *Filesystem) PutUploadPart(ctx context.Context, uploadID string, partNumber int32, source io.Reader, expectedSize int64) (UploadPartObject, error) {
	if err := ctx.Err(); err != nil {
		return UploadPartObject{}, err
	}
	if !objectKeyPattern.MatchString(uploadID) || partNumber < 1 {
		return UploadPartObject{}, ErrInvalidKey
	}
	if expectedSize <= 0 {
		return UploadPartObject{}, ErrSizeMismatch
	}

	finalPath := s.uploadPartPath(uploadID, partNumber)
	directory := filepath.Dir(finalPath)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return UploadPartObject{}, fmt.Errorf("create upload staging directory: %w", err)
	}

	temp, err := os.CreateTemp(directory, ".part-*")
	if err != nil {
		return UploadPartObject{}, fmt.Errorf("create upload part staging file: %w", err)
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temp.Close()
			_ = os.Remove(tempName)
		}
	}()

	hasher := sha256.New()
	limited := io.LimitReader(&contextReader{ctx: ctx, reader: source}, expectedSize+1)
	written, err := io.Copy(io.MultiWriter(temp, hasher), limited)
	if err != nil {
		return UploadPartObject{}, fmt.Errorf("write upload part: %w", err)
	}
	if written != expectedSize {
		return UploadPartObject{}, fmt.Errorf("%w: expected %d bytes, got %d", ErrSizeMismatch, expectedSize, written)
	}
	if err := temp.Sync(); err != nil {
		return UploadPartObject{}, fmt.Errorf("sync upload part: %w", err)
	}
	if err := temp.Close(); err != nil {
		return UploadPartObject{}, fmt.Errorf("close upload part: %w", err)
	}
	if err := os.Chmod(tempName, 0o640); err != nil {
		return UploadPartObject{}, fmt.Errorf("set upload part permissions: %w", err)
	}

	actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
	result := UploadPartObject{
		UploadID:   uploadID,
		PartNumber: partNumber,
		SHA256:     actualSHA256,
		Size:       written,
	}

	if existing, err := s.uploadPartEvidence(finalPath); err == nil {
		if existing.Size != written || existing.SHA256 != actualSHA256 {
			return UploadPartObject{}, ErrUploadPartConflict
		}
		result.Created = false
		return result, nil
	} else if !os.IsNotExist(err) {
		return UploadPartObject{}, err
	}

	if err := os.Link(tempName, finalPath); err != nil {
		if os.IsExist(err) {
			existing, evidenceErr := s.uploadPartEvidence(finalPath)
			if evidenceErr != nil {
				return UploadPartObject{}, evidenceErr
			}
			if existing.Size != written || existing.SHA256 != actualSHA256 {
				return UploadPartObject{}, ErrUploadPartConflict
			}
			result.Created = false
			return result, nil
		}
		return UploadPartObject{}, fmt.Errorf("commit upload part: %w", err)
	}

	if err := syncDirectory(directory); err != nil {
		_ = os.Remove(finalPath)
		return UploadPartObject{}, fmt.Errorf("sync upload part directory: %w", err)
	}
	if err := os.Remove(tempName); err != nil {
		_ = os.Remove(finalPath)
		return UploadPartObject{}, fmt.Errorf("remove upload part temporary link: %w", err)
	}
	committed = true
	result.Created = true
	return result, nil
}

func (s *Filesystem) OpenUploadPart(ctx context.Context, uploadID string, partNumber int32) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !objectKeyPattern.MatchString(uploadID) || partNumber < 1 {
		return nil, ErrInvalidKey
	}
	path := s.uploadPartPath(uploadID, partNumber)
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, ErrUploadPartNotFound
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("upload part is not a regular file")
	}
	return os.Open(path)
}

func (s *Filesystem) DeleteUploadPart(ctx context.Context, uploadID string, partNumber int32, expectedSHA256 string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	if !objectKeyPattern.MatchString(uploadID) || partNumber < 1 || !sha256Pattern.MatchString(expectedSHA256) {
		return ErrInvalidKey
	}

	path := s.uploadPartPath(uploadID, partNumber)
	evidence, err := s.uploadPartEvidence(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if evidence.SHA256 != expectedSHA256 {
		return ErrUploadPartConflict
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove upload part: %w", err)
	}
	return syncDirectory(filepath.Dir(path))
}

func (s *Filesystem) Probe(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, directory := range []string{s.staging, s.uploadStaging} {
		if err := probeDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

func (s *Filesystem) objectPath(key string) string {
	prefix := key[:2]
	return filepath.Join(s.originals, prefix, key)
}

func (s *Filesystem) uploadPartPath(uploadID string, partNumber int32) string {
	return filepath.Join(s.uploadStaging, uploadID[:2], uploadID, strconv.FormatInt(int64(partNumber), 10)+".part")
}

func (s *Filesystem) uploadPartEvidence(path string) (UploadPartObject, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return UploadPartObject{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return UploadPartObject{}, fmt.Errorf("upload part is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return UploadPartObject{}, err
	}
	defer file.Close()

	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return UploadPartObject{}, err
	}
	return UploadPartObject{
		SHA256: hex.EncodeToString(hasher.Sum(nil)),
		Size:   size,
	}, nil
}

func probeDirectory(directory string) error {
	temp, err := os.CreateTemp(directory, "probe-*")
	if err != nil {
		return fmt.Errorf("create storage probe: %w", err)
	}
	name := temp.Name()
	if _, err := temp.Write([]byte{0}); err != nil {
		_ = temp.Close()
		_ = os.Remove(name)
		return fmt.Errorf("write storage probe: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		_ = os.Remove(name)
		return fmt.Errorf("sync storage probe: %w", err)
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("close storage probe: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove storage probe: %w", err)
	}
	return syncDirectory(directory)
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}
