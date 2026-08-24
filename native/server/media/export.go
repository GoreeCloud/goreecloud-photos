package media

import (
	"errors"
	"strings"
)

type ExportAsset struct {
	MediaID      string   `json:"mediaId"`
	PortablePath string   `json:"portablePath"`
	SHA256       string   `json:"sha256"`
	AlbumIDs     []string `json:"albumIds,omitempty"`
}

type ExportManifest struct {
	SchemaVersion string        `json:"schemaVersion"`
	OwnerID       string        `json:"ownerId"`
	Assets        []ExportAsset `json:"assets"`
}

func (m ExportManifest) Validate() error {
	if m.SchemaVersion != "goreecloud.photos.export.v1" {
		return errors.New("unsupported export schema version")
	}
	if strings.TrimSpace(m.OwnerID) == "" {
		return errors.New("owner id is required")
	}
	seen := make(map[string]struct{}, len(m.Assets))
	for _, asset := range m.Assets {
		if strings.TrimSpace(asset.MediaID) == "" || strings.TrimSpace(asset.PortablePath) == "" {
			return errors.New("export assets require media id and portable path")
		}
		if len(asset.SHA256) != 64 {
			return errors.New("export asset sha256 must be a 64-character hex digest")
		}
		if _, exists := seen[asset.MediaID]; exists {
			return errors.New("duplicate media id in export manifest")
		}
		seen[asset.MediaID] = struct{}{}
	}
	return nil
}

type Repository interface {
	Get(ownerID, mediaID string) (Item, error)
	List(ownerID string) ([]Item, error)
}
