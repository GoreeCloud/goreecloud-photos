package id

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewUUIDv7() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	milliseconds := uint64(time.Now().UnixMilli())
	value[0] = byte(milliseconds >> 40)
	value[1] = byte(milliseconds >> 32)
	value[2] = byte(milliseconds >> 24)
	value[3] = byte(milliseconds >> 16)
	value[4] = byte(milliseconds >> 8)
	value[5] = byte(milliseconds)
	value[6] = (value[6] & 0x0f) | 0x70
	value[8] = (value[8] & 0x3f) | 0x80

	encoded := make([]byte, 32)
	hex.Encode(encoded, value[:])

	result := make([]byte, 36)
	copy(result[0:8], encoded[0:8])
	result[8] = '-'
	copy(result[9:13], encoded[8:12])
	result[13] = '-'
	copy(result[14:18], encoded[12:16])
	result[18] = '-'
	copy(result[19:23], encoded[16:20])
	result[23] = '-'
	copy(result[24:36], encoded[20:32])

	return string(result), nil
}
