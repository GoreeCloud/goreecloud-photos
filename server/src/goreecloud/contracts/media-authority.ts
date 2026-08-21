export type GoreeCloudAssetId = string;
export type GoreeCloudUserId = string;

export type MediaAuthorityKind = 'original' | 'derived' | 'temporary';

export type MediaProvenanceSource =
  | 'device-upload'
  | 'filesystem-import'
  | 'google-photos-takeout'
  | 'ente-export'
  | 'immich-import'
  | 'camera-import'
  | 'removable-media'
  | 'goreecloud-photos-export'
  | 'other';

export interface MediaIntegrityEvidence {
  algorithm: 'sha256' | 'sha1' | 'md5' | 'other';
  digest: string;
  byteLength: number;
}

export interface MediaStorageReference {
  provider: string;
  location: string;
  storageClass?: string;
}

export interface MediaProvenance {
  source: MediaProvenanceSource;
  sourceIdentifier?: string;
  sourcePath?: string;
  importedAt: string;
  importerVersion?: string;
  notes?: readonly string[];
}

export interface AuthoritativeMediaRecord {
  assetId: GoreeCloudAssetId;
  ownerId: GoreeCloudUserId;
  authority: 'original';
  mediaType: string;
  storage: MediaStorageReference;
  integrity: MediaIntegrityEvidence;
  provenance: MediaProvenance;
  capturedAt?: string;
  createdAt: string;
  deletedAt?: string;
}

export interface DerivedMediaRecord {
  assetId: GoreeCloudAssetId;
  authority: 'derived' | 'temporary';
  derivedFromAssetId: GoreeCloudAssetId;
  purpose: string;
  storage?: MediaStorageReference;
  rebuildable: true;
}

export interface MediaAuthorityProvider {
  getAuthoritativeMedia(assetId: GoreeCloudAssetId): Promise<AuthoritativeMediaRecord | null>;
  verifyIntegrity(assetId: GoreeCloudAssetId): Promise<boolean>;
  resolveOriginal(assetId: GoreeCloudAssetId): Promise<MediaStorageReference>;
}
