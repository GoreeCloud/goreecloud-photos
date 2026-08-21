import type { GoreeCloudAssetId, GoreeCloudUserId, MediaIntegrityEvidence, MediaProvenance } from './media-authority.js';

export interface PortableAssetManifestEntry {
  assetId: GoreeCloudAssetId;
  ownerId: GoreeCloudUserId;
  relativePath: string;
  mediaType: string;
  integrity: MediaIntegrityEvidence;
  provenance: MediaProvenance;
  capturedAt?: string;
  description?: string;
  favorite?: boolean;
  albumIds?: readonly string[];
}

export interface PortableExportManifest {
  format: 'goreecloud-photos-portable-export';
  version: '0.1';
  createdAt: string;
  assets: readonly PortableAssetManifestEntry[];
  albums?: readonly PortableAlbumManifestEntry[];
  warnings?: readonly string[];
}

export interface PortableAlbumManifestEntry {
  albumId: string;
  name: string;
  ownerId: GoreeCloudUserId;
  assetIds: readonly GoreeCloudAssetId[];
}

export interface ImportCandidate {
  sourceType: MediaProvenance['source'];
  sourceIdentifier?: string;
  sourcePath: string;
  mediaType?: string;
  integrity?: MediaIntegrityEvidence;
  capturedAt?: string;
  sidecars?: readonly string[];
  sourceAlbumNames?: readonly string[];
}

export interface ImportDecision {
  candidate: ImportCandidate;
  disposition: 'import' | 'duplicate' | 'skip' | 'reject';
  existingAssetId?: GoreeCloudAssetId;
  warnings?: readonly string[];
}

export interface PortableTransferProvider {
  inspectImport(candidate: ImportCandidate): Promise<ImportDecision>;
  exportLibrary(ownerId: GoreeCloudUserId): Promise<PortableExportManifest>;
}
