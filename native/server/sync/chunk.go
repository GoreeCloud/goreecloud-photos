package sync

import "errors"

type UploadChunk struct {
	Offset int64
	Length int64
}

// NextChunk plans the next bounded transfer range from the last committed
// resumable offset. Planning does not advance progress; only a later confirmed
// transfer acknowledgement may call Advance.
func (p UploadProgress) NextChunk(maxBytes int64) (UploadChunk, bool, error) {
	if !p.valid() {
		return UploadChunk{}, false, errors.New("invalid upload progress state")
	}
	if maxBytes <= 0 {
		return UploadChunk{}, false, errors.New("upload chunk size must be positive")
	}
	if p.Complete() {
		return UploadChunk{}, false, nil
	}

	length := p.RemainingBytes()
	if length > maxBytes {
		length = maxBytes
	}
	return UploadChunk{Offset: p.CommittedBytes, Length: length}, true, nil
}

func (c UploadChunk) EndOffset() int64 {
	return c.Offset + c.Length
}
