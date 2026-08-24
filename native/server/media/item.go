package media

import (
    "errors"
    "strings"
    "time"
)

type Kind string

const (
    KindPhoto Kind = "photo"
    KindVideo Kind = "video"
)

type Item struct {
    ID          string
    OwnerID     string
    Kind        Kind
    OriginalURI string
    Filename    string
    MIMEType    string
    CapturedAt  time.Time
    ImportedAt  time.Time
    SizeBytes   int64
    SHA256      string
}

func (i Item) Validate() error {
    if strings.TrimSpace(i.ID) == "" || strings.TrimSpace(i.OwnerID) == "" {
        return errors.New("id and owner id are required")
    }
    if i.Kind != KindPhoto && i.Kind != KindVideo {
        return errors.New("unsupported media kind")
    }
    if strings.TrimSpace(i.OriginalURI) == "" || strings.TrimSpace(i.Filename) == "" {
        return errors.New("original uri and filename are required")
    }
    if i.SizeBytes < 0 {
        return errors.New("size must be non-negative")
    }
    if len(i.SHA256) != 64 {
        return errors.New("sha256 must be a 64-character hex digest")
    }
    return nil
}
