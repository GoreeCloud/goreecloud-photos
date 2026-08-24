export type MediaKind = "image" | "video";

export interface AuthoritativeMedia {
  id: string;
  ownerId: string;
  kind: MediaKind;
  originalPath: string;
  contentType: string;
  byteLength: number;
  checksum: string;
  capturedAt?: string;
  importedAt: string;
}

export interface DerivedMedia {
  id: string;
  sourceMediaId: string;
  purpose: "thumbnail" | "preview" | "embedding" | "transcode";
  rebuildable: true;
}

/**
 * Native GoreeCloud Photos authority boundary.
 * Original media is user-owned authoritative data. Derived state is reconstructable and
 * cannot silently replace, rename, move, delete, or otherwise become authoritative over an
 * original. Destructive operations require explicit authorization at a higher service layer.
 */
export interface MediaAuthority {
  getOriginal(id: string, ownerId: string): Promise<AuthoritativeMedia | null>;
  listOriginals(ownerId: string): Promise<AuthoritativeMedia[]>;
  listDerived(sourceMediaId: string, ownerId: string): Promise<DerivedMedia[]>;
}

export interface PortableExportEntry {
  media: AuthoritativeMedia;
  portableRelativePath: string;
  albumIds: string[];
}

export interface PortableExportManifest {
  schema: "goreecloud.photos.portable-export";
  version: 1;
  generatedAt: string;
  entries: PortableExportEntry[];
}
