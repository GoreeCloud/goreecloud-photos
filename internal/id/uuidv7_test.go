package id

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewUUIDv7(t *testing.T) {
	before := time.Now().UnixMilli()
	value, err := NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now().UnixMilli()

	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		t.Fatalf("unexpected UUID format: %q", value)
	}
	if value[14] != '7' {
		t.Fatalf("expected version 7 UUID, got %q", value)
	}
	if !strings.ContainsRune("89ab", rune(value[19])) {
		t.Fatalf("expected RFC 9562 variant, got %q", value)
	}

	rawHex := strings.ReplaceAll(value, "-", "")
	raw, err := hex.DecodeString(rawHex)
	if err != nil {
		t.Fatal(err)
	}
	milliseconds, err := strconv.ParseUint(hex.EncodeToString(raw[:6]), 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	if int64(milliseconds) < before || int64(milliseconds) > after {
		t.Fatalf("UUID timestamp %d outside generation window [%d, %d]", milliseconds, before, after)
	}
}
