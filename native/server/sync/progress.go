package sync

import "errors"

type UploadProgress struct {
	TotalBytes     int64
	CommittedBytes int64
}

func NewUploadProgress(totalBytes int64) (UploadProgress, error) {
	if totalBytes <= 0 {
		return UploadProgress{}, errors.New("upload total bytes must be positive")
	}
	return UploadProgress{TotalBytes: totalBytes}, nil
}

func (p UploadProgress) Advance(committedBytes int64) (UploadProgress, error) {
	if !p.valid() {
		return UploadProgress{}, errors.New("invalid upload progress state")
	}
	if committedBytes < p.CommittedBytes {
		return UploadProgress{}, errors.New("upload progress cannot move backwards")
	}
	if committedBytes > p.TotalBytes {
		return UploadProgress{}, errors.New("upload progress cannot exceed total bytes")
	}
	p.CommittedBytes = committedBytes
	return p, nil
}

func (p UploadProgress) ResumeOffset() int64 {
	if !p.valid() {
		return 0
	}
	return p.CommittedBytes
}

func (p UploadProgress) RemainingBytes() int64 {
	if !p.valid() {
		return 0
	}
	return p.TotalBytes - p.CommittedBytes
}

func (p UploadProgress) Complete() bool {
	return p.valid() && p.CommittedBytes == p.TotalBytes
}

func (p UploadProgress) valid() bool {
	return p.TotalBytes > 0 && p.CommittedBytes >= 0 && p.CommittedBytes <= p.TotalBytes
}
