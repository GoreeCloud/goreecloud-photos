package sync

import "errors"

// AcknowledgeChunk applies a transport acknowledgement only to the exact
// currently planned resumable range. Partial acknowledgement within the range
// is allowed; stale, future, empty, or out-of-range acknowledgements fail closed.
func (p UploadProgress) AcknowledgeChunk(chunk UploadChunk, committedBytes int64) (UploadProgress, error) {
	if !p.valid() {
		return UploadProgress{}, errors.New("invalid upload progress state")
	}
	if chunk.Offset != p.CommittedBytes || chunk.Length <= 0 || chunk.Offset < 0 || chunk.Offset > p.TotalBytes || chunk.Length > p.TotalBytes-chunk.Offset {
		return UploadProgress{}, errors.New("invalid upload chunk acknowledgement scope")
	}
	end := chunk.Offset + chunk.Length
	if committedBytes < chunk.Offset || committedBytes > end {
		return UploadProgress{}, errors.New("upload acknowledgement falls outside planned chunk")
	}
	return p.Advance(committedBytes)
}
