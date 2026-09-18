package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrAlreadyExists      = errors.New("storage object already exists")
	ErrChecksumMismatch   = errors.New("storage checksum mismatch")
	ErrInvalidKey         = errors.New("invalid storage object key")
	ErrSizeMismatch       = errors.New("storage object size mismatch")
	ErrUploadPartConflict = errors.New("upload part conflicts with staged bytes")
	ErrUploadPartNotFound = errors.New("upload part not found")
)

type Object struct {
	Key    string
	SHA256 string
	Size   int64
}

type UploadPartObject struct {
	UploadID   string
	PartNumber int32
	SHA256     string
	Size       int64
	Created    bool
}

type OriginalStore interface {
	PutImmutable(ctx context.Context, key string, source io.Reader, expectedSHA256 string) (Object, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Probe(ctx context.Context) error
}

type UploadStagingStore interface {
	PutUploadPart(ctx context.Context, uploadID string, partNumber int32, source io.Reader, expectedSize int64) (UploadPartObject, error)
	OpenUploadPart(ctx context.Context, uploadID string, partNumber int32) (io.ReadCloser, error)
	DeleteUploadPart(ctx context.Context, uploadID string, partNumber int32, expectedSHA256 string) error
}
