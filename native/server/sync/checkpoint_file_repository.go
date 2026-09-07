package sync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const MaxUploadCheckpointFileBytes = 128 * 1024

var (
	ErrInvalidUploadCheckpointFileRepository = errors.New("invalid upload checkpoint file repository")
	ErrUploadCheckpointAlreadyExists         = errors.New("upload checkpoint already exists")
	ErrUploadCheckpointFileBusy              = errors.New("upload checkpoint file repository is busy")
)

type uploadCheckpointFileEnvelope struct {
	Revision UploadCheckpointRevision `json:"revision"`
	Payload  []byte                   `json:"payload"`
}

// FileUploadCheckpointRepository is a durable single-process implementation of
// UploadCheckpointRepository. Each transfer checkpoint is stored as one
// canonical, bounded JSON envelope under a hashed owner/media key. Writes use a
// temporary file plus fsync+rename so an interrupted write cannot expose a
// partially written checkpoint.
//
// The in-process mutex provides optimistic-CAS correctness for one GoreeCloud
// Photos server process. This adapter deliberately does not claim distributed
// or multi-process locking; deployments that run concurrent server processes
// must use a repository with an external transactional authority.
type FileUploadCheckpointRepository struct {
	root string
	mu   sync.Mutex
}

func NewFileUploadCheckpointRepository(root string) (*FileUploadCheckpointRepository, error) {
	if strings.TrimSpace(root) == "" {
		return nil, ErrInvalidUploadCheckpointFileRepository
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, ErrInvalidUploadCheckpointFileRepository
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return nil, fmt.Errorf("create upload checkpoint directory: %w", err)
	}
	if err := os.Chmod(absolute, 0o700); err != nil {
		return nil, fmt.Errorf("protect upload checkpoint directory: %w", err)
	}
	return &FileUploadCheckpointRepository{root: absolute}, nil
}

func (r *FileUploadCheckpointRepository) Load(
	ctx context.Context,
	ownerID, mediaID string,
) (UploadCheckpointRecord, bool, error) {
	if r == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(mediaID) == "" {
		return UploadCheckpointRecord{}, false, ErrInvalidUploadCheckpointFileRepository
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadLocked(ctx, ownerID, mediaID)
}

func (r *FileUploadCheckpointRepository) Create(
	ctx context.Context,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	if r == nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointFileRepository
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, err
	}
	record, err := NewUploadCheckpointRecord(checkpoint)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	identity, err := checkpointFileIdentity(checkpoint)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, err
	}
	path := r.recordPath(identity.ownerID, identity.mediaID)
	encoded, err := encodeUploadCheckpointFileEnvelope(record)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return UploadCheckpointRecord{}, ErrUploadCheckpointAlreadyExists
	}
	if err != nil {
		return UploadCheckpointRecord{}, fmt.Errorf("create upload checkpoint file: %w", err)
	}
	writeErr := writeAndSyncCheckpointFile(file, encoded)
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(path)
		return UploadCheckpointRecord{}, writeErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return UploadCheckpointRecord{}, fmt.Errorf("close upload checkpoint file: %w", closeErr)
	}
	if err := syncDirectory(r.root); err != nil {
		return UploadCheckpointRecord{}, err
	}
	return record, nil
}

func (r *FileUploadCheckpointRepository) Save(
	ctx context.Context,
	expected UploadCheckpointRevision,
	checkpoint UploadCheckpoint,
) (UploadCheckpointRecord, error) {
	if r == nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointFileRepository
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, err
	}
	identity, err := checkpointFileIdentity(checkpoint)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	current, found, err := r.loadLocked(ctx, identity.ownerID, identity.mediaID)
	if err != nil {
		return UploadCheckpointRecord{}, err
	}
	if !found {
		return UploadCheckpointRecord{}, ErrUploadCheckpointRecordNotFound
	}
	if current.Revision() != expected {
		return current, ErrStaleUploadCheckpointRevision
	}
	next, err := current.Update(expected, checkpoint)
	if err != nil {
		return current, err
	}
	encoded, err := encodeUploadCheckpointFileEnvelope(next)
	if err != nil {
		return current, err
	}
	if err := ctx.Err(); err != nil {
		return current, err
	}
	if err := r.replaceLocked(r.recordPath(identity.ownerID, identity.mediaID), encoded); err != nil {
		return current, err
	}
	return next, nil
}

type uploadCheckpointFileIdentityValue struct {
	ownerID string
	mediaID string
}

func checkpointFileIdentity(checkpoint UploadCheckpoint) (uploadCheckpointFileIdentityValue, error) {
	validated, err := RestoreUploadCheckpoint(checkpoint.record, checkpoint.progress)
	if err != nil {
		return uploadCheckpointFileIdentityValue{}, ErrInvalidUploadCheckpointRecord
	}
	return uploadCheckpointFileIdentityValue{
		ownerID: validated.record.OwnerID,
		mediaID: validated.record.MediaID,
	}, nil
}

func (r *FileUploadCheckpointRepository) loadLocked(
	ctx context.Context,
	ownerID, mediaID string,
) (UploadCheckpointRecord, bool, error) {
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, false, err
	}
	path := r.recordPath(ownerID, mediaID)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return UploadCheckpointRecord{}, false, nil
	}
	if err != nil {
		return UploadCheckpointRecord{}, false, fmt.Errorf("open upload checkpoint file: %w", err)
	}
	defer file.Close()
	limited := io.LimitReader(file, MaxUploadCheckpointFileBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return UploadCheckpointRecord{}, false, fmt.Errorf("read upload checkpoint file: %w", err)
	}
	if len(payload) == 0 || len(payload) > MaxUploadCheckpointFileBytes {
		return UploadCheckpointRecord{}, false, ErrInvalidUploadCheckpointRecord
	}
	record, err := decodeUploadCheckpointFileEnvelope(payload)
	if err != nil {
		return UploadCheckpointRecord{}, false, err
	}
	checkpoint, err := record.Restore()
	if err != nil || checkpoint.record.OwnerID != ownerID || checkpoint.record.MediaID != mediaID {
		return UploadCheckpointRecord{}, false, ErrInvalidUploadCheckpointRecord
	}
	if err := ctx.Err(); err != nil {
		return UploadCheckpointRecord{}, false, err
	}
	return record, true, nil
}

func (r *FileUploadCheckpointRepository) replaceLocked(path string, encoded []byte) error {
	temporary, err := os.CreateTemp(r.root, ".checkpoint-*.tmp")
	if err != nil {
		return fmt.Errorf("create upload checkpoint temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("protect upload checkpoint temporary file: %w", err)
	}
	if err := writeAndSyncCheckpointFile(temporary, encoded); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close upload checkpoint temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace upload checkpoint file: %w", err)
	}
	removeTemporary = false
	return syncDirectory(r.root)
}

func (r *FileUploadCheckpointRepository) recordPath(ownerID, mediaID string) string {
	digest := sha256.Sum256([]byte(ownerID + "\x00" + mediaID))
	return filepath.Join(r.root, hex.EncodeToString(digest[:])+".json")
}

func encodeUploadCheckpointFileEnvelope(record UploadCheckpointRecord) ([]byte, error) {
	if _, err := record.Restore(); err != nil {
		return nil, ErrInvalidUploadCheckpointRecord
	}
	payload, err := json.Marshal(uploadCheckpointFileEnvelope{
		Revision: record.Revision(),
		Payload:  record.Payload(),
	})
	if err != nil || len(payload) == 0 || len(payload) > MaxUploadCheckpointFileBytes {
		return nil, ErrInvalidUploadCheckpointRecord
	}
	return payload, nil
}

func decodeUploadCheckpointFileEnvelope(payload []byte) (UploadCheckpointRecord, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var envelope uploadCheckpointFileEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	record, err := RestoreUploadCheckpointRecord(envelope.Revision, envelope.Payload)
	if err != nil {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	canonical, err := json.Marshal(uploadCheckpointFileEnvelope{
		Revision: record.Revision(),
		Payload:  record.Payload(),
	})
	if err != nil || !bytes.Equal(payload, canonical) {
		return UploadCheckpointRecord{}, ErrInvalidUploadCheckpointRecord
	}
	return record, nil
}

func writeAndSyncCheckpointFile(file *os.File, payload []byte) error {
	if len(payload) == 0 || len(payload) > MaxUploadCheckpointFileBytes {
		return ErrInvalidUploadCheckpointRecord
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate upload checkpoint file: %w", err)
	}
	if _, err := file.Write(payload); err != nil {
		return fmt.Errorf("write upload checkpoint file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync upload checkpoint file: %w", err)
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open upload checkpoint directory: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync upload checkpoint directory: %w", err)
	}
	return nil
}
