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
	"strings"
)

var (
	objectKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{7,127}$`)
	sha256Pattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type Filesystem struct {
	root      string
	originals string
	staging   string
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
		root:      absoluteRoot,
		originals: filepath.Join(absoluteRoot, "originals"),
		staging:   filepath.Join(absoluteRoot, "staging"),
	}

	for _, path := range []string{store.root, store.originals, store.staging} {
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

func (s *Filesystem) Probe(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	temp, err := os.CreateTemp(s.staging, "probe-*")
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
	return syncDirectory(s.staging)
}

func (s *Filesystem) objectPath(key string) string {
	prefix := key[:2]
	return filepath.Join(s.originals, prefix, key)
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
