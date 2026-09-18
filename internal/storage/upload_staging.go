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

	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
)

var uploadIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (s *Filesystem) PutUploadPart(ctx context.Context, uploadID string, partNumber int, source io.Reader, expectedSize int64) (upload.StagedPart, error) {
	if err := ctx.Err(); err != nil {
		return upload.StagedPart{}, err
	}
	uploadID = strings.ToLower(strings.TrimSpace(uploadID))
	if !uploadIDPattern.MatchString(uploadID) || partNumber < 1 || expectedSize <= 0 {
		return upload.StagedPart{}, upload.ErrInvalidRequest
	}

	partsDir, err := s.ensureUploadPartsDirectory(uploadID)
	if err != nil {
		return upload.StagedPart{}, err
	}
	finalPath := filepath.Join(partsDir, fmt.Sprintf("%08d.part", partNumber))
	temp, err := os.CreateTemp(partsDir, ".incoming-*")
	if err != nil {
		return upload.StagedPart{}, fmt.Errorf("create upload staging file: %w", err)
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
	written, err := io.Copy(io.MultiWriter(temp, hasher), &contextReader{ctx: ctx, reader: io.LimitReader(source, expectedSize+1)})
	if err != nil {
		return upload.StagedPart{}, fmt.Errorf("write upload part: %w", err)
	}
	if written != expectedSize {
		return upload.StagedPart{}, upload.ErrPartSize
	}
	if err := temp.Sync(); err != nil {
		return upload.StagedPart{}, fmt.Errorf("sync upload part: %w", err)
	}
	if err := temp.Close(); err != nil {
		return upload.StagedPart{}, fmt.Errorf("close upload part: %w", err)
	}
	if err := os.Chmod(tempName, 0o640); err != nil {
		return upload.StagedPart{}, fmt.Errorf("set upload part permissions: %w", err)
	}

	actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
	if err := os.Link(tempName, finalPath); err != nil {
		if os.IsExist(err) {
			existingSHA256, existingSize, inspectErr := hashRegularFile(finalPath)
			if inspectErr != nil {
				return upload.StagedPart{}, inspectErr
			}
			if existingSize != expectedSize || existingSHA256 != actualSHA256 {
				return upload.StagedPart{}, upload.ErrPartConflict
			}
			_ = os.Remove(tempName)
			committed = true
			return upload.StagedPart{StorageKey: uploadPartStorageKey(uploadID, partNumber), SHA256: actualSHA256, Size: written}, nil
		}
		return upload.StagedPart{}, fmt.Errorf("commit upload part: %w", err)
	}

	if err := syncDirectory(partsDir); err != nil {
		_ = os.Remove(finalPath)
		return upload.StagedPart{}, fmt.Errorf("sync upload part directory: %w", err)
	}
	if err := os.Remove(tempName); err != nil {
		_ = os.Remove(finalPath)
		return upload.StagedPart{}, fmt.Errorf("remove upload staging link: %w", err)
	}
	committed = true
	return upload.StagedPart{StorageKey: uploadPartStorageKey(uploadID, partNumber), SHA256: actualSHA256, Size: written, Created: true}, nil
}

func (s *Filesystem) RemoveUploadPart(ctx context.Context, uploadID string, partNumber int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	uploadID = strings.ToLower(strings.TrimSpace(uploadID))
	if !uploadIDPattern.MatchString(uploadID) || partNumber < 1 {
		return upload.ErrInvalidRequest
	}
	partsDir := filepath.Join(s.staging, "uploads", uploadID, "parts")
	path := filepath.Join(partsDir, fmt.Sprintf("%08d.part", partNumber))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(partsDir); err == nil {
		return syncDirectory(partsDir)
	}
	return nil
}

func (s *Filesystem) RemoveUploadSession(ctx context.Context, uploadID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	uploadID = strings.ToLower(strings.TrimSpace(uploadID))
	if !uploadIDPattern.MatchString(uploadID) {
		return upload.ErrInvalidRequest
	}
	uploadsDir := filepath.Join(s.staging, "uploads")
	if err := os.RemoveAll(filepath.Join(uploadsDir, uploadID)); err != nil {
		return err
	}
	if _, err := os.Stat(uploadsDir); err == nil {
		return syncDirectory(uploadsDir)
	}
	return nil
}

func (s *Filesystem) ensureUploadPartsDirectory(uploadID string) (string, error) {
	uploadsDir := filepath.Join(s.staging, "uploads")
	sessionDir := filepath.Join(uploadsDir, uploadID)
	partsDir := filepath.Join(sessionDir, "parts")
	for _, path := range []string{uploadsDir, sessionDir, partsDir} {
		if err := ensureDirectory(path); err != nil {
			return "", err
		}
	}
	return partsDir, nil
}

func ensureDirectory(path string) error {
	if err := os.Mkdir(path, 0o750); err == nil {
		return nil
	} else if !os.IsExist(err) {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("upload staging path is not a directory")
	}
	return nil
}

func hashRegularFile(path string) (string, int64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", 0, fmt.Errorf("upload part is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

func uploadPartStorageKey(uploadID string, partNumber int) string {
	return fmt.Sprintf("staging/uploads/%s/parts/%08d.part", uploadID, partNumber)
}
