package media

import (
	"strings"
	"testing"
)

func TestExportManifestValidate(t *testing.T) {
	manifest := ExportManifest{
		SchemaVersion: "goreecloud.photos.export.v1",
		OwnerID:       "user-1",
		Assets: []ExportAsset{{
			MediaID: "media-1", PortablePath: "photos/2026/photo.jpg", SHA256: strings.Repeat("a", 64), AlbumIDs: []string{"album-1"},
		}},
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("expected valid export manifest: %v", err)
	}
}

func TestExportManifestRejectsDuplicateMedia(t *testing.T) {
	digest := strings.Repeat("b", 64)
	manifest := ExportManifest{
		SchemaVersion: "goreecloud.photos.export.v1",
		OwnerID:       "user-1",
		Assets: []ExportAsset{
			{MediaID: "media-1", PortablePath: "photos/one.jpg", SHA256: digest},
			{MediaID: "media-1", PortablePath: "photos/two.jpg", SHA256: digest},
		},
	}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected duplicate media id to be rejected")
	}
}
