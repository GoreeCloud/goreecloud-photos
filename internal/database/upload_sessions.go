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
	ErrInvalidIdempotency    = errors.New("invalid idempotency request")
	ErrIdempotencyConflict   = errors.New("idempotency key conflicts with an earlier request")
)

var (
	uuidPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type UploadSession struct {
	UploadID         string
	ActorSubjectID   string
	LibraryID        string
	OriginalFilename string
	MediaType        string
	DeviceID         string
	CaptureTime      *time.Time
	CaptureTimeZone  string
	ExpectedSize     int64
	ExpectedSHA256   *string
	ReceivedBytes    int64
	PartSize         int64
	State            string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Parts            []UploadPart
}

type UploadPart struct {
	PartNumber int32
	ByteSize   int64
	SHA256     string
	CreatedAt  time.Time
}

type CreateUploadSessionParams struct {
	UploadID         string
	ActorSubjectID   string
	LibraryID        string
	OriginalFilename string
	MediaType        string
	DeviceID         string
	CaptureTime      *time.Time
	CaptureTimeZone  string
	ExpectedSize     int64
	ExpectedSHA256   *string
	PartSize         int64
	ExpiresAt        time.Time
}

type CreateUploadSessionIdempotentParams struct {
	IdempotencyKey string
	RequestSHA256  string
	Session        CreateUploadSessionParams
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
	return insertUploadSession(ctx, p.pool, params)
}

func (p *PostgreSQL) CreateUploadSessionIdempotent(ctx context.Context, params CreateUploadSessionIdempotentParams) (UploadSession, bool, error) {
	if p == nil || p.pool == nil {
		return UploadSession{}, false, ErrNotInitialized
	}
	if err := validateCreateUploadSession(params.Session); err != nil {
		return UploadSession{}, false, err
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" || len(params.IdempotencyKey) > 128 || !sha256HexPattern.MatchString(params.RequestSHA256) {
		return UploadSession{}, false, ErrInvalidIdempotency
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UploadSession{}, false, fmt.Errorf("begin idempotent upload transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_records (
			actor_subject_id,
			idempotency_key,
			request_sha256,
			response_status,
			response_body,
			expires_at
		)
		VALUES ($1, $2, $3, 102, NULL, $4)
		ON CONFLICT (actor_subject_id, idempotency_key) DO NOTHING
	`,
		params.Session.ActorSubjectID,
		params.IdempotencyKey,
		params.RequestSHA256,
		params.Session.ExpiresAt.UTC(),
	)
	if err != nil {
		return UploadSession{}, false, fmt.Errorf("reserve idempotency key: %w", err)
	}

	if tag.RowsAffected() == 0 {
		var storedSHA256 string
		var status int
		var uploadID string
		if err := tx.QueryRow(ctx, `
			SELECT
				request_sha256,
				response_status,
				COALESCE(response_body->>'upload_id', '')
			FROM idempotency_records
			WHERE actor_subject_id = $1 AND idempotency_key = $2
			FOR UPDATE
		`, params.Session.ActorSubjectID, params.IdempotencyKey).Scan(&storedSHA256, &status, &uploadID); err != nil {
			return UploadSession{}, false, fmt.Errorf("load idempotency record: %w", err)
		}
		if storedSHA256 != params.RequestSHA256 {
			return UploadSession{}, false, ErrIdempotencyConflict
		}
		if status != 201 || uploadID == "" {
			return UploadSession{}, false, fmt.Errorf("idempotency record is incomplete")
		}
		session, err := loadUploadSession(ctx, tx, uploadID)
		if err != nil {
			return UploadSession{}, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return UploadSession{}, false, fmt.Errorf("commit idempotent replay: %w", err)
		}
		return session, true, nil
	}

	session, err := insertUploadSession(ctx, tx, params.Session)
	if err != nil {
		return UploadSession{}, false, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE idempotency_records
		SET response_status = 201,
			response_body = jsonb_build_object('upload_id', $3)
		WHERE actor_subject_id = $1 AND idempotency_key = $2
	`, params.Session.ActorSubjectID, params.IdempotencyKey, session.UploadID); err != nil {
		return UploadSession{}, false, fmt.Errorf("complete idempotency record: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return UploadSession{}, false, fmt.Errorf("commit idempotent upload creation: %w", err)
	}
	return session, false, nil
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

func (p *PostgreSQL) PrepareUploadPart(ctx context.Context, uploadID string, partNumber int32) (UploadSession, int64, error) {
	if p == nil || p.pool == nil {
		return UploadSession{}, 0, ErrNotInitialized
	}
	if !uuidPattern.MatchString(uploadID) || partNumber < 1 {
		return UploadSession{}, 0, ErrInvalidUploadPart
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UploadSession{}, 0, fmt.Errorf("begin upload preparation transaction: %w", err)
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
	`, uploadID).Scan(&expectedSize, &partSize, &state, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadSession{}, 0, ErrUploadSessionNotFound
	}
	if err != nil {
		return UploadSession{}, 0, fmt.Errorf("lock upload session: %w", err)
	}

	if (state == "open" || state == "receiving") && !expiresAt.After(time.Now()) {
		if _, err := tx.Exec(ctx, `
			UPDATE upload_sessions
			SET state = 'expired', updated_at = NOW()
			WHERE upload_id = $1
		`, uploadID); err != nil {
			return UploadSession{}, 0, fmt.Errorf("expire upload session: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return UploadSession{}, 0, fmt.Errorf("commit expired upload session: %w", err)
		}
		return UploadSession{}, 0, ErrUploadSessionExpired
	}
	if state != "open" && state != "receiving" {
		return UploadSession{}, 0, ErrUploadSessionClosed
	}

	expectedPartSize, err := uploadPartExpectedSize(partNumber, expectedSize, partSize)
	if err != nil {
		return UploadSession{}, 0, err
	}
	session, err := loadUploadSession(ctx, tx, uploadID)
	if err != nil {
		return UploadSession{}, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UploadSession{}, 0, fmt.Errorf("commit upload preparation: %w", err)
	}
	return session, expectedPartSize, nil
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

	expectedPartSize, err := uploadPartExpectedSize(params.PartNumber, expectedSize, partSize)
	if err != nil || params.ByteSize != expectedPartSize {
		return UploadSession{}, ErrInvalidUploadPart
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

type uploadSessionInserter interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func insertUploadSession(ctx context.Context, inserter uploadSessionInserter, params CreateUploadSessionParams) (UploadSession, error) {
	var expected sql.NullString
	if params.ExpectedSHA256 != nil {
		expected.String = *params.ExpectedSHA256
		expected.Valid = true
	}
	var captureTime sql.NullTime
	if params.CaptureTime != nil {
		captureTime.Time = params.CaptureTime.UTC()
		captureTime.Valid = true
	}

	var session UploadSession
	var storedExpected sql.NullString
	var storedFilename sql.NullString
	var storedMediaType sql.NullString
	var storedDeviceID sql.NullString
	var storedCaptureTime sql.NullTime
	var storedCaptureTimeZone sql.NullString
	err := inserter.QueryRow(ctx, `
		INSERT INTO upload_sessions (
			upload_id,
			actor_subject_id,
			library_id,
			original_filename,
			media_type,
			device_id,
			capture_time,
			capture_time_zone,
			expected_size,
			expected_sha256,
			received_bytes,
			part_size,
			state,
			expires_at
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''), $9, $10, 0, $11, 'open', $12)
		RETURNING
			upload_id::text,
			actor_subject_id,
			library_id::text,
			original_filename,
			media_type,
			device_id,
			capture_time,
			capture_time_zone,
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
		params.OriginalFilename,
		params.MediaType,
		params.DeviceID,
		captureTime,
		params.CaptureTimeZone,
		params.ExpectedSize,
		expected,
		params.PartSize,
		params.ExpiresAt.UTC(),
	).Scan(
		&session.UploadID,
		&session.ActorSubjectID,
		&session.LibraryID,
		&storedFilename,
		&storedMediaType,
		&storedDeviceID,
		&storedCaptureTime,
		&storedCaptureTimeZone,
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
	applyOptionalUploadSessionFields(&session, storedFilename, storedMediaType, storedDeviceID, storedCaptureTime, storedCaptureTimeZone, storedExpected)
	session.Parts = []UploadPart{}
	return session, nil
}

func loadUploadSession(ctx context.Context, querier uploadSessionQuerier, uploadID string) (UploadSession, error) {
	var session UploadSession
	var expected sql.NullString
	var filename sql.NullString
	var mediaType sql.NullString
	var deviceID sql.NullString
	var captureTime sql.NullTime
	var captureTimeZone sql.NullString
	err := querier.QueryRow(ctx, `
		SELECT
			upload_id::text,
			actor_subject_id,
			library_id::text,
			original_filename,
			media_type,
			device_id,
			capture_time,
			capture_time_zone,
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
		&filename,
		&mediaType,
		&deviceID,
		&captureTime,
		&captureTimeZone,
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
	applyOptionalUploadSessionFields(&session, filename, mediaType, deviceID, captureTime, captureTimeZone, expected)

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

func applyOptionalUploadSessionFields(
	session *UploadSession,
	filename sql.NullString,
	mediaType sql.NullString,
	deviceID sql.NullString,
	captureTime sql.NullTime,
	captureTimeZone sql.NullString,
	expected sql.NullString,
) {
	session.OriginalFilename = filename.String
	session.MediaType = mediaType.String
	session.DeviceID = deviceID.String
	session.CaptureTimeZone = captureTimeZone.String
	if captureTime.Valid {
		value := captureTime.Time
		session.CaptureTime = &value
	}
	session.ExpectedSHA256 = nullableStringPointer(expected)
}

func validateCreateUploadSession(params CreateUploadSessionParams) error {
	if !uuidPattern.MatchString(params.UploadID) ||
		!uuidPattern.MatchString(params.LibraryID) ||
		strings.TrimSpace(params.ActorSubjectID) == "" ||
		params.ExpectedSize <= 0 ||
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

func uploadPartExpectedSize(partNumber int32, expectedSize int64, partSize int64) (int64, error) {
	if partNumber < 1 || expectedSize <= 0 || partSize <= 0 {
		return 0, ErrInvalidUploadPart
	}

	index := int64(partNumber - 1)
	if index > math.MaxInt64/partSize {
		return 0, ErrInvalidUploadPart
	}
	offset := index * partSize
	if offset >= expectedSize {
		return 0, ErrInvalidUploadPart
	}
	remaining := expectedSize - offset
	if remaining < partSize {
		return remaining, nil
	}
	return partSize, nil
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
