import { createHash } from 'node:crypto';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { AssetRepository } from 'src/repositories/asset.repository';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ImmichMediaAuthorityAdapter } from './immich-media-authority.adapter';

const assetId = '00000000-0000-4000-8000-000000000001';
const ownerId = '00000000-0000-4000-8000-000000000002';

const sha1 = (value: Buffer | string) => createHash('sha1').update(value).digest();

describe(ImmichMediaAuthorityAdapter.name, () => {
  let directory: string;
  let originalPath: string;
  let getById: ReturnType<typeof vi.fn>;
  let adapter: ImmichMediaAuthorityAdapter;

  beforeEach(async () => {
    directory = await mkdtemp(join(tmpdir(), 'goreecloud-photos-media-authority-'));
    originalPath = join(directory, 'original.jpg');
    getById = vi.fn();
    adapter = new ImmichMediaAuthorityAdapter({ getById } as unknown as AssetRepository);
  });

  afterEach(async () => {
    await rm(directory, { recursive: true, force: true });
  });

  const asset = (overrides: Record<string, unknown> = {}) => ({
    id: assetId,
    ownerId,
    libraryId: null,
    originalPath,
    checksum: sha1('photo-bytes'),
    checksumAlgorithm: 'sha1',
    type: 'IMAGE',
    fileCreatedAt: new Date('2026-08-20T12:00:00.000Z'),
    createdAt: new Date('2026-08-20T12:05:00.000Z'),
    deletedAt: null,
    ...overrides,
  });

  it('maps inherited Immich metadata into the GoreeCloud authoritative-media contract', async () => {
    await writeFile(originalPath, 'photo-bytes');
    getById.mockResolvedValue(asset());

    await expect(adapter.getAuthoritativeMedia(assetId)).resolves.toEqual({
      assetId,
      ownerId,
      authority: 'original',
      mediaType: 'IMAGE',
      storage: {
        provider: 'immich-inherited-filesystem',
        location: originalPath,
      },
      integrity: {
        algorithm: 'sha1',
        digest: sha1('photo-bytes').toString('hex'),
        byteLength: Buffer.byteLength('photo-bytes'),
      },
      provenance: {
        source: 'immich-import',
        sourcePath: originalPath,
        importedAt: '2026-08-20T12:05:00.000Z',
        notes: ['Read through the inherited Immich asset repository; provenance is transitional.'],
      },
      capturedAt: '2026-08-20T12:00:00.000Z',
      createdAt: '2026-08-20T12:05:00.000Z',
      deletedAt: undefined,
    });

    expect(getById).toHaveBeenCalledWith(assetId);
  });

  it('marks library-backed assets as filesystem imports', async () => {
    await writeFile(originalPath, 'photo-bytes');
    getById.mockResolvedValue(asset({ libraryId: 'library-1' }));

    const record = await adapter.getAuthoritativeMedia(assetId);

    expect(record?.provenance.source).toBe('filesystem-import');
  });

  it('returns null when authoritative media is missing', async () => {
    getById.mockResolvedValue(null);

    await expect(adapter.getAuthoritativeMedia(assetId)).resolves.toBeNull();
  });

  it('verifies the inherited SHA-1 checksum without modifying the source file', async () => {
    await writeFile(originalPath, 'photo-bytes');
    getById.mockResolvedValue(asset());

    await expect(adapter.verifyIntegrity(assetId)).resolves.toBe(true);
  });

  it('rejects checksum mismatches', async () => {
    await writeFile(originalPath, 'different-bytes');
    getById.mockResolvedValue(asset());

    await expect(adapter.verifyIntegrity(assetId)).resolves.toBe(false);
  });

  it('fails closed for missing assets, unsupported checksum algorithms, and filesystem errors', async () => {
    getById.mockResolvedValueOnce(null);
    await expect(adapter.verifyIntegrity(assetId)).resolves.toBe(false);

    getById.mockResolvedValueOnce(asset({ checksumAlgorithm: 'sha256' }));
    await expect(adapter.verifyIntegrity(assetId)).resolves.toBe(false);

    getById.mockResolvedValueOnce(asset());
    await expect(adapter.verifyIntegrity(assetId)).resolves.toBe(false);
  });

  it('resolves only the inherited original-file storage reference', async () => {
    getById.mockResolvedValue(asset());

    await expect(adapter.resolveOriginal(assetId)).resolves.toEqual({
      provider: 'immich-inherited-filesystem',
      location: originalPath,
    });
  });

  it('throws a bounded error when an original cannot be resolved', async () => {
    getById.mockResolvedValue(null);

    await expect(adapter.resolveOriginal(assetId)).rejects.toThrow(`Authoritative media not found: ${assetId}`);
  });
});
