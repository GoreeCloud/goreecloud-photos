package sync

import (
	"errors"
	"math"
	"time"
)

type TransferNetwork string

const (
	TransferNetworkUnmetered TransferNetwork = "unmetered"
	TransferNetworkMetered   TransferNetwork = "metered"
)

type BandwidthPolicy struct {
	BytesPerSecond int64
	MaxChunkBytes  int64
	AllowMetered   bool
}

// ChunkLimit returns the maximum bytes that may be planned during one policy
// window. A denied metered network or a sub-byte window returns no budget
// without manufacturing transfer progress. This policy does not sleep, meter
// traffic, inspect the operating system network, or execute transport.
func (p BandwidthPolicy) ChunkLimit(network TransferNetwork, window time.Duration) (int64, bool, error) {
	if p.BytesPerSecond <= 0 || p.MaxChunkBytes <= 0 {
		return 0, false, errors.New("invalid bandwidth policy")
	}
	if window <= 0 {
		return 0, false, errors.New("bandwidth policy window must be positive")
	}
	switch network {
	case TransferNetworkUnmetered:
	case TransferNetworkMetered:
		if !p.AllowMetered {
			return 0, false, nil
		}
	default:
		return 0, false, errors.New("unknown transfer network")
	}

	budget := bytesForWindow(p.BytesPerSecond, window)
	if budget <= 0 {
		return 0, false, nil
	}
	if budget > p.MaxChunkBytes {
		budget = p.MaxChunkBytes
	}
	return budget, true, nil
}

// NextChunkWithBandwidth combines the validated resumable checkpoint with a
// pure bandwidth budget. Network observation and actual rate enforcement remain
// responsibilities of the future transfer executor.
func (c UploadCheckpoint) NextChunkWithBandwidth(
	policy BandwidthPolicy,
	network TransferNetwork,
	window time.Duration,
) (UploadChunk, bool, error) {
	limit, allowed, err := policy.ChunkLimit(network, window)
	if err != nil || !allowed {
		return UploadChunk{}, false, err
	}
	return c.NextChunk(limit)
}

func bytesForWindow(rate int64, window time.Duration) int64 {
	seconds := int64(window / time.Second)
	remainder := int64(window % time.Second)
	if seconds > math.MaxInt64/rate {
		return math.MaxInt64
	}
	whole := rate * seconds

	const nanosPerSecond = int64(time.Second)
	fraction := (rate/nanosPerSecond)*remainder +
		((rate%nanosPerSecond)*remainder)/nanosPerSecond
	if whole > math.MaxInt64-fraction {
		return math.MaxInt64
	}
	return whole + fraction
}
