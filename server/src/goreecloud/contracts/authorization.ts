import type { GoreeCloudAssetId, GoreeCloudUserId } from './media-authority.js';

export type PhotosAction =
  | 'asset.read'
  | 'asset.download'
  | 'asset.upload'
  | 'asset.update'
  | 'asset.trash'
  | 'asset.purge'
  | 'album.read'
  | 'album.manage'
  | 'share.create'
  | 'share.consume'
  | 'library.manage'
  | 'admin.manage';

export interface AuthorizationContext {
  actorId: GoreeCloudUserId;
  action: PhotosAction;
  assetId?: GoreeCloudAssetId;
  libraryId?: string;
  albumId?: string;
  shareId?: string;
  serviceAuthority?: string;
  administrative: boolean;
  publicationBoundary: boolean;
}

export interface AuthorizationDecision {
  allowed: boolean;
  reason: string;
  policyId?: string;
}

export interface AuthorizationProvider {
  authorize(context: AuthorizationContext): Promise<AuthorizationDecision>;
}

export function requireAuthorized(decision: AuthorizationDecision): void {
  if (!decision.allowed) {
    throw new Error(`GoreeCloud Photos authorization denied: ${decision.reason}`);
  }
}
