package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrAlreadyExists    = errors.New("storage object already exists")
	ErrChecksumMismatch = errors.New("storage checksum mismatch")
	ErrInvalidKey       = errors.New("invalid storage object key")
)

type Object struct {
	Key    string
	SHA256 string
	Size   int64
}

type OriginalStore interface {
	PutImmutable(ctx context.Context, key string, source io.Reader, expectedSHA256 string) (Object, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Probe(ctx context.Context) error
}
