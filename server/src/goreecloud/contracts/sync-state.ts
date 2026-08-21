import type { GoreeCloudAssetId, GoreeCloudUserId } from './media-authority.js';

export type SyncDisposition = 'pending' | 'uploading' | 'synced' | 'conflict' | 'failed' | 'tombstoned';

export interface SyncStateRecord {
  ownerId: GoreeCloudUserId;
  deviceId: string;
  clientAssetId: string;
  serverAssetId?: GoreeCloudAssetId;
  disposition: SyncDisposition;
  checksum?: string;
  lastReconciledAt?: string;
  retryCount: number;
  errorCode?: string;
  deletionRequestedAt?: string;
}

export interface SyncReconciliationResult {
  state: SyncStateRecord;
  destructiveActionAuthorized: boolean;
  reason: string;
}

export interface SyncStateProvider {
  read(ownerId: GoreeCloudUserId, deviceId: string, clientAssetId: string): Promise<SyncStateRecord | null>;
  reconcile(state: SyncStateRecord): Promise<SyncReconciliationResult>;
}
