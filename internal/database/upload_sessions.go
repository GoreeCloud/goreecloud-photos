package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidUploadSession  = errors.New("invalid upload session")
	ErrUploadSessionNotFound = errors.New("upload session not found")
	ErrUploadSessionClosed   = errors.New("upload session is not writable")
	ErrUploadSessionExpired  = errors.New("upload session expired")
	ErrInvalidUploadPart     = errors.New("invalid upload part")
	ErrUploadPartConflict    = errors.New("upload part conflicts with committed evidence")
)

var (
	uuidPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type UploadSession struct {
	UploadID       string
	ActorSubjectID string
	LibraryID      string
	ExpectedSize   int64
	ExpectedSHA256 *string
	ReceivedBytes  int64
	PartSize       int64
	State          string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Parts          []UploadPart
}

type UploadPart struct {
	PartNumber int32
	ByteSize   int64
	SHA256     string
	CreatedAt  time.Time
}

type CreateUploadSessionParams struct {
	UploadID       string
	ActorSubjectID string
	LibraryID      string
	ExpectedSize   int64
	ExpectedSHA256 *string
	PartSize       int64
	ExpiresAt      time.Time
}

type RecordUploadPartParams struct {
	UploadID   string
	PartNumber int32
	ByteSize   int64
	SHA256     string
}

func (p *PostgreSQL) CreateUploadSession(ctx context.Context, params CreateUploadSessionParams) (UploadSession, error) {
	if p == nil || p.pool == nil {
		return UploadSession{}, ErrNotInitialized
	}
	if err := validateCreateUploadSession(params); err != nil {
		return UploadSession{}, err
	}

	var expected sql.NullString
	if params.ExpectedSHA256 != nil {
		expected.String = *params.ExpectedSHA256
		expected.Valid = true
	}

	var session UploadSession
	var storedExpected sql.NullString
	err := p.pool.QueryRow(ctx, `
		INSERT INTO upload_sessions (
			upload_id,
			actor_subject_id,
			library_id,
			expected_size,
			expected_sha256,
			received_bytes,
			part_size,
			state,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, 0, $6, 'open', $7)
		RETURNING
			upload_id::text,
			actor_subject_id,
			library_id::text,
			expected_size,
			expected_sha256,
			received_bytes,
			part_size,
			state,
			expires_at,
			created_at,
			updated_at
	`,
		params.UploadID,
		params.ActorSubjectID,
		params.LibraryID,
		params.ExpectedSize,
		expected,
		params.PartSize,
		params.ExpiresAt.UTC(),
	).Scan(
		&session.UploadID,
		&session.ActorSubjectID,
		&session.LibraryID,
		&session.ExpectedSize,
		&storedExpected,
		&session.ReceivedBytes,
		&session.PartSize,
		&session.State,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return UploadSession{}, fmt.Errorf("create upload session: %w", err)
	}
	session.ExpectedSHA256 = nullableStringPointer(storedExpected)
	session.Parts = []UploadPart{}
	return session, nil
}

func (p *PostgreSQL) GetUploadSession(ctx context.Context, uploadID string) (UploadSession, error) {
	if p == nil || p.pool == nil {
		return UploadSession{}, ErrNotInitialized
	}
	if !uuidPattern.MatchString(uploadID) {
		return UploadSession{}, ErrInvalidUploadSession
	}

	return loadUploadSession(ctx, p.pool, uploadID)
}

func (p *PostgreSQL) RecordUploadPart(ctx context.Context, params RecordUploadPartParams) (UploadSession, error) {
	if p == nil || p.pool == nil {
		return UploadSession{}, ErrNotInitialized
	}
	if err := validateRecordUploadPart(params); err != nil {
		return UploadSession{}, err
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UploadSession{}, fmt.Errorf("begin upload part transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	var expectedSize int64
	var partSize int64
	var state string
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT expected_size, part_size, state, expires_at
		FROM upload_sessions
		WHERE upload_id = $1
		FOR UPDATE
	`, params.UploadID).Scan(&expectedSize, &partSize, &state, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadSession{}, ErrUploadSessionNotFound
	}
	if err != nil {
		return UploadSession{}, fmt.Errorf("lock upload session: %w", err)
	}

	if (state == "open" || state == "receiving") && !expiresAt.After(time.Now()) {
		if _, err := tx.Exec(ctx, `
			UPDATE upload_sessions
			SET state = 'expired', updated_at = NOW()
			WHERE upload_id = $1
		`, params.UploadID); err != nil {
			return UploadSession{}, fmt.Errorf("expire upload session: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return UploadSession{}, fmt.Errorf("commit expired upload session: %w", err)
		}
		return UploadSession{}, ErrUploadSessionExpired
	}
	if state != "open" && state != "receiving" {
		return UploadSession{}, ErrUploadSessionClosed
	}

	if err := validatePartBounds(params.PartNumber, params.ByteSize, expectedSize, partSize); err != nil {
		return UploadSession{}, err
	}

	var existingSize int64
	var existingSHA256 string
	err = tx.QueryRow(ctx, `
		SELECT byte_size, content_sha256
		FROM upload_parts
		WHERE upload_id = $1 AND part_number = $2
	`, params.UploadID, params.PartNumber).Scan(&existingSize, &existingSHA256)
	switch {
	case err == nil:
		if existingSize != params.ByteSize || existingSHA256 != params.SHA256 {
			return UploadSession{}, ErrUploadPartConflict
		}
		session, err := loadUploadSession(ctx, tx, params.UploadID)
		if err != nil {
			return UploadSession{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return UploadSession{}, fmt.Errorf("commit idempotent upload part: %w", err)
		}
		return session, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return UploadSession{}, fmt.Errorf("inspect upload part: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO upload_parts (upload_id, part_number, byte_size, content_sha256)
		VALUES ($1, $2, $3, $4)
	`, params.UploadID, params.PartNumber, params.ByteSize, params.SHA256); err != nil {
		return UploadSession{}, fmt.Errorf("record upload part: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE upload_sessions
		SET
			received_bytes = (
				SELECT COALESCE(SUM(byte_size), 0)
				FROM upload_parts
				WHERE upload_id = $1
			),
			state = CASE WHEN state = 'open' THEN 'receiving' ELSE state END,
			updated_at = NOW()
		WHERE upload_id = $1
	`, params.UploadID); err != nil {
		return UploadSession{}, fmt.Errorf("update upload session progress: %w", err)
	}

	session, err := loadUploadSession(ctx, tx, params.UploadID)
	if err != nil {
		return UploadSession{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UploadSession{}, fmt.Errorf("commit upload part: %w", err)
	}
	return session, nil
}

type uploadSessionQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func loadUploadSession(ctx context.Context, querier uploadSessionQuerier, uploadID string) (UploadSession, error) {
	var session UploadSession
	var expected sql.NullString
	err := querier.QueryRow(ctx, `
		SELECT
			upload_id::text,
			actor_subject_id,
			library_id::text,
			expected_size,
			expected_sha256,
			received_bytes,
			part_size,
			state,
			expires_at,
			created_at,
			updated_at
		FROM upload_sessions
		WHERE upload_id = $1
	`, uploadID).Scan(
		&session.UploadID,
		&session.ActorSubjectID,
		&session.LibraryID,
		&session.ExpectedSize,
		&expected,
		&session.ReceivedBytes,
		&session.PartSize,
		&session.State,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadSession{}, ErrUploadSessionNotFound
	}
	if err != nil {
		return UploadSession{}, fmt.Errorf("load upload session: %w", err)
	}
	session.ExpectedSHA256 = nullableStringPointer(expected)

	rows, err := querier.Query(ctx, `
		SELECT part_number, byte_size, content_sha256, created_at
		FROM upload_parts
		WHERE upload_id = $1
		ORDER BY part_number
	`, uploadID)
	if err != nil {
		return UploadSession{}, fmt.Errorf("load upload parts: %w", err)
	}
	defer rows.Close()

	session.Parts = make([]UploadPart, 0)
	for rows.Next() {
		var part UploadPart
		if err := rows.Scan(&part.PartNumber, &part.ByteSize, &part.SHA256, &part.CreatedAt); err != nil {
			return UploadSession{}, fmt.Errorf("scan upload part: %w", err)
		}
		session.Parts = append(session.Parts, part)
	}
	if err := rows.Err(); err != nil {
		return UploadSession{}, fmt.Errorf("iterate upload parts: %w", err)
	}

	return session, nil
}

func validateCreateUploadSession(params CreateUploadSessionParams) error {
	if !uuidPattern.MatchString(params.UploadID) ||
		!uuidPattern.MatchString(params.LibraryID) ||
		strings.TrimSpace(params.ActorSubjectID) == "" ||
		params.ExpectedSize < 0 ||
		params.PartSize <= 0 ||
		!params.ExpiresAt.After(time.Now()) {
		return ErrInvalidUploadSession
	}
	if params.ExpectedSHA256 != nil && !sha256HexPattern.MatchString(*params.ExpectedSHA256) {
		return ErrInvalidUploadSession
	}
	return nil
}

func validateRecordUploadPart(params RecordUploadPartParams) error {
	if !uuidPattern.MatchString(params.UploadID) ||
		params.PartNumber < 1 ||
		params.ByteSize <= 0 ||
		!sha256HexPattern.MatchString(params.SHA256) {
		return ErrInvalidUploadPart
	}
	return nil
}

func validatePartBounds(partNumber int32, byteSize int64, expectedSize int64, partSize int64) error {
	if expectedSize <= 0 || partSize <= 0 || byteSize > partSize {
		return ErrInvalidUploadPart
	}

	index := int64(partNumber - 1)
	if index > math.MaxInt64/partSize {
		return ErrInvalidUploadPart
	}
	offset := index * partSize
	if offset >= expectedSize || byteSize > expectedSize-offset {
		return ErrInvalidUploadPart
	}
	if byteSize != partSize && offset+byteSize != expectedSize {
		return ErrInvalidUploadPart
	}
	return nil
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
