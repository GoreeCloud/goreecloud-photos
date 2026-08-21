import { createHash } from 'node:crypto';
import { createReadStream } from 'node:fs';
import { stat } from 'node:fs/promises';
import { AssetRepository } from 'src/repositories/asset.repository';
import type {
  AuthoritativeMediaRecord,
  GoreeCloudAssetId,
  MediaAuthorityProvider,
  MediaIntegrityEvidence,
  MediaProvenance,
  MediaStorageReference,
} from '../contracts/media-authority';

const fileDigest = async (path: string, algorithm: 'sha1' | 'sha256') =>
  new Promise<string>((resolve, reject) => {
    const hash = createHash(algorithm);
    const stream = createReadStream(path);
    stream.on('error', reject);
    stream.on('data', (chunk) => hash.update(chunk));
    stream.on('end', () => resolve(hash.digest('hex')));
  });

/**
 * Read-only compatibility boundary over the inherited Immich asset repository.
 *
 * This adapter deliberately performs no writes, migrations, path changes, or
 * deletion operations. It lets GoreeCloud-owned code consume the
 * MediaAuthorityProvider contract while inherited storage remains authoritative
 * during the fork-to-native transition.
 */
export class ImmichMediaAuthorityAdapter implements MediaAuthorityProvider {
  constructor(private readonly assets: AssetRepository) {}

  private storage(location: string): MediaStorageReference {
    return { provider: 'immich-inherited-filesystem', location };
  }

  private provenance(asset: { libraryId: string | null; originalPath: string; createdAt: Date }): MediaProvenance {
    return {
      source: asset.libraryId ? 'filesystem-import' : 'immich-import',
      sourcePath: asset.originalPath,
      importedAt: asset.createdAt.toISOString(),
      notes: ['Read through the inherited Immich asset repository; provenance is transitional.'],
    };
  }

  private async integrity(asset: { checksum: Buffer; checksumAlgorithm: string; originalPath: string }) {
    const { size } = await stat(asset.originalPath);
    const algorithm = asset.checksumAlgorithm === 'sha1' ? 'sha1' : 'other';
    const evidence: MediaIntegrityEvidence = {
      algorithm,
      digest: asset.checksum.toString('hex'),
      byteLength: size,
    };
    return evidence;
  }

  async getAuthoritativeMedia(assetId: GoreeCloudAssetId): Promise<AuthoritativeMediaRecord | null> {
    const asset = await this.assets.getById(assetId);
    if (!asset) {
      return null;
    }

    return {
      assetId: asset.id,
      ownerId: asset.ownerId,
      authority: 'original',
      mediaType: asset.type,
      storage: this.storage(asset.originalPath),
      integrity: await this.integrity(asset),
      provenance: this.provenance(asset),
      capturedAt: asset.fileCreatedAt.toISOString(),
      createdAt: asset.createdAt.toISOString(),
      deletedAt: asset.deletedAt?.toISOString(),
    };
  }

  async verifyIntegrity(assetId: GoreeCloudAssetId): Promise<boolean> {
    const asset = await this.assets.getById(assetId);
    if (!asset || asset.checksumAlgorithm !== 'sha1') {
      return false;
    }

    try {
      return (await fileDigest(asset.originalPath, 'sha1')) === asset.checksum.toString('hex');
    } catch {
      return false;
    }
  }

  async resolveOriginal(assetId: GoreeCloudAssetId): Promise<MediaStorageReference> {
    const asset = await this.assets.getById(assetId);
    if (!asset) {
      throw new Error(`Authoritative media not found: ${assetId}`);
    }

    return this.storage(asset.originalPath);
  }
}
