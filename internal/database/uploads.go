package database

import (
	"context"
	"errors"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
	"github.com/jackc/pgx/v5"
)

const uploadSessionColumns = `
	upload_id::text, library_id::text, actor_subject_id,
	COALESCE(original_filename, ''), COALESCE(media_type, ''),
	expected_size, COALESCE(expected_sha256, ''), received_bytes,
	part_size, state, COALESCE(device_id, ''), expires_at, created_at, updated_at
`

func (p *PostgreSQL) CreateUploadSession(ctx context.Context, session upload.Session) (upload.Session, error) {
	if p == nil || p.pool == nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	row := p.pool.QueryRow(ctx, `
		INSERT INTO upload_sessions (
			upload_id, actor_subject_id, library_id, original_filename, media_type,
			device_id, expected_size, expected_sha256, received_bytes, part_size,
			state, expires_at, created_at, updated_at
		)
		SELECT $1::uuid, $2, l.library_id, $4, $5, NULLIF($6, ''), $7,
			NULLIF($8, ''), 0, $9, 'open', $10, $11, $11
		FROM libraries l
		WHERE l.library_id = $3::uuid
		  AND l.lifecycle_state = 'active'
		  AND (
			l.owner_subject_id = $2 OR EXISTS (
				SELECT 1 FROM library_memberships m
				WHERE m.library_id = l.library_id
				  AND m.subject_id = $2
				  AND m.role IN ('owner', 'manager', 'contributor')
			)
		  )
		RETURNING `+uploadSessionColumns,
		session.UploadID, session.ActorSubjectID, session.LibraryID,
		session.OriginalFilename, session.MediaType, session.DeviceID,
		session.ExpectedSize, session.ExpectedSHA256, session.PartSize,
		session.ExpiresAt, session.CreatedAt,
	)
	created, err := scanUploadSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return upload.Session{}, upload.ErrNotFound
	}
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	created.TransferMethod = "parts-v1"
	created.Parts = []upload.Part{}
	return created, nil
}

func (p *PostgreSQL) GetUploadSession(ctx context.Context, actorSubjectID, uploadID string) (upload.Session, error) {
	if p == nil || p.pool == nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	session, err := scanUploadSession(p.pool.QueryRow(ctx, `
		SELECT `+uploadSessionColumns+`
		FROM upload_sessions
		WHERE upload_id = $1::uuid AND actor_subject_id = $2
	`, uploadID, actorSubjectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return upload.Session{}, upload.ErrNotFound
	}
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	parts, err := p.listUploadParts(ctx, uploadID)
	if err != nil {
		return upload.Session{}, err
	}
	session.TransferMethod = "parts-v1"
	session.Parts = parts
	return session, nil
}

func (p *PostgreSQL) RecordUploadPart(ctx context.Context, actorSubjectID, uploadID string, part upload.Part) (upload.Session, error) {
	if p == nil || p.pool == nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	defer tx.Rollback(ctx)

	session, err := scanUploadSession(tx.QueryRow(ctx, `
		SELECT `+uploadSessionColumns+`
		FROM upload_sessions
		WHERE upload_id = $1::uuid AND actor_subject_id = $2
		FOR UPDATE
	`, uploadID, actorSubjectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return upload.Session{}, upload.ErrNotFound
	}
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	if !session.ExpiresAt.After(time.Now().UTC()) {
		_, _ = tx.Exec(ctx, `UPDATE upload_sessions SET state = 'expired', updated_at = NOW() WHERE upload_id = $1::uuid`, uploadID)
		if tx.Commit(ctx) != nil {
			return upload.Session{}, upload.ErrUnavailable
		}
		return upload.Session{}, upload.ErrExpired
	}
	if session.State != "open" && session.State != "receiving" {
		return upload.Session{}, upload.ErrInvalidState
	}
	if part.PartNumber < 1 {
		return upload.Session{}, upload.ErrPartSize
	}
	expectedOffset := int64(part.PartNumber-1) * session.PartSize
	if expectedOffset < 0 || expectedOffset >= session.ExpectedSize || part.ByteOffset != expectedOffset {
		return upload.Session{}, upload.ErrPartSize
	}
	expectedSize := session.PartSize
	if remaining := session.ExpectedSize - expectedOffset; remaining < expectedSize {
		expectedSize = remaining
	}
	if part.ByteSize != expectedSize {
		return upload.Session{}, upload.ErrPartSize
	}

	var existing upload.Part
	err = tx.QueryRow(ctx, `
		SELECT part_number, byte_offset, byte_size, content_sha256, storage_key, created_at
		FROM upload_parts WHERE upload_id = $1::uuid AND part_number = $2
	`, uploadID, part.PartNumber).Scan(
		&existing.PartNumber, &existing.ByteOffset, &existing.ByteSize,
		&existing.ContentSHA256, &existing.StorageKey, &existing.CreatedAt,
	)
	if err == nil {
		if existing.ByteOffset != part.ByteOffset || existing.ByteSize != part.ByteSize ||
			existing.ContentSHA256 != part.ContentSHA256 || existing.StorageKey != part.StorageKey {
			return upload.Session{}, upload.ErrPartConflict
		}
		if err := tx.Commit(ctx); err != nil {
			return upload.Session{}, upload.ErrUnavailable
		}
		return p.GetUploadSession(ctx, actorSubjectID, uploadID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return upload.Session{}, upload.ErrUnavailable
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO upload_parts (
			upload_id, part_number, byte_offset, byte_size, content_sha256, storage_key, created_at
		) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
	`, uploadID, part.PartNumber, part.ByteOffset, part.ByteSize, part.ContentSHA256, part.StorageKey, part.CreatedAt); err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	if _, err := tx.Exec(ctx, `
		UPDATE upload_sessions
		SET received_bytes = received_bytes + $2, state = 'receiving', updated_at = NOW()
		WHERE upload_id = $1::uuid AND received_bytes + $2 <= expected_size
	`, uploadID, part.ByteSize); err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	return p.GetUploadSession(ctx, actorSubjectID, uploadID)
}

func (p *PostgreSQL) CancelUploadSession(ctx context.Context, actorSubjectID, uploadID string) (upload.Session, error) {
	if p == nil || p.pool == nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	defer tx.Rollback(ctx)

	session, err := scanUploadSession(tx.QueryRow(ctx, `
		SELECT `+uploadSessionColumns+`
		FROM upload_sessions
		WHERE upload_id = $1::uuid AND actor_subject_id = $2
		FOR UPDATE
	`, uploadID, actorSubjectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return upload.Session{}, upload.ErrNotFound
	}
	if err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	if session.State == "complete" {
		return upload.Session{}, upload.ErrInvalidState
	}
	if session.State != "cancelled" {
		if _, err := tx.Exec(ctx, `
			UPDATE upload_sessions SET state = 'cancelled', updated_at = NOW()
			WHERE upload_id = $1::uuid
		`, uploadID); err != nil {
			return upload.Session{}, upload.ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return upload.Session{}, upload.ErrUnavailable
	}
	return p.GetUploadSession(ctx, actorSubjectID, uploadID)
}

func (p *PostgreSQL) listUploadParts(ctx context.Context, uploadID string) ([]upload.Part, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT part_number, byte_offset, byte_size, content_sha256, storage_key, created_at
		FROM upload_parts WHERE upload_id = $1::uuid ORDER BY part_number
	`, uploadID)
	if err != nil {
		return nil, upload.ErrUnavailable
	}
	defer rows.Close()

	parts := make([]upload.Part, 0)
	for rows.Next() {
		var part upload.Part
		if err := rows.Scan(&part.PartNumber, &part.ByteOffset, &part.ByteSize, &part.ContentSHA256, &part.StorageKey, &part.CreatedAt); err != nil {
			return nil, upload.ErrUnavailable
		}
		parts = append(parts, part)
	}
	if rows.Err() != nil {
		return nil, upload.ErrUnavailable
	}
	return parts, nil
}

func scanUploadSession(row pgx.Row) (upload.Session, error) {
	var session upload.Session
	err := row.Scan(
		&session.UploadID, &session.LibraryID, &session.ActorSubjectID,
		&session.OriginalFilename, &session.MediaType, &session.ExpectedSize,
		&session.ExpectedSHA256, &session.ReceivedBytes, &session.PartSize,
		&session.State, &session.DeviceID, &session.ExpiresAt,
		&session.CreatedAt, &session.UpdatedAt,
	)
	return session, err
}
