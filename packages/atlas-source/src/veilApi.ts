import { createVeilClient } from "@playingwithclouds/veil-sdk/client";

export interface ListedGallery {
  id: string;
  title: string;
  coverPath?: string;
  imageCount?: number;
}

export interface ListedScene {
  id: string;
  title: string;
  posterPath?: string;
  durationSeconds?: number;
}

export interface ResolvedScene {
  title: string;
  url: string;
  headers: Record<string, string>;
  durationSeconds?: number;
}

export interface VeilApi {
  baseUrl: string;
  imageProxyUrl(remoteUrl: string): string;
  listGalleries(search: string, limit: number, offset: number): Promise<ListedGallery[]>;
  listScenes(search: string, limit: number, offset: number): Promise<ListedScene[]>;
  galleryImages(galleryId: string): Promise<{ title: string; urls: string[] }>;
  resolveScene(sceneId: string): Promise<ResolvedScene>;
}

interface StreamCandidate {
  url: string;
  kind: string;
  format?: string | null;
  mimeType?: string | null;
}

interface GalleryRecord {
  title: string;
  images: { filePath: string; position?: number | null }[];
}

const GALLERY_SELECTION = { title: true, images: { filePath: true, position: true } } as const;

/** Prefers direct streams over downloads, and mp4 over other containers. */
function streamScore(stream: StreamCandidate): number {
  let score = 0;
  if (stream.kind === "stream") {
    score += 2;
  }
  const format = (stream.format ?? "").toLowerCase();
  const mimeType = (stream.mimeType ?? "").toLowerCase();
  if (format.includes("mp4") || mimeType.includes("mp4")) {
    score += 1;
  }
  return score;
}

function pickBestStream(streams: StreamCandidate[]): StreamCandidate {
  let best = streams[0];
  for (const stream of streams) {
    if (streamScore(stream) > streamScore(best)) {
      best = stream;
    }
  }
  return best;
}

function sortedImageUrls(gallery: GalleryRecord): string[] {
  const images = [...gallery.images].sort((left, right) => (left.position ?? 0) - (right.position ?? 0));
  return images.map((image) => image.filePath).filter((filePath) => Boolean(filePath));
}

export function createVeilApi(baseUrl: string): VeilApi {
  const veil = createVeilClient({ url: `${baseUrl}/graphql` });

  async function query(selection: Record<string, unknown>): Promise<Record<string, any>> {
    return (await veil.client.query(selection as never)) as Record<string, any>;
  }

  async function mutation(selection: Record<string, unknown>): Promise<Record<string, any>> {
    return (await veil.client.mutation(selection as never)) as Record<string, any>;
  }

  // Galleries from search/browse are stubs; veil scrapes the origin page on demand.
  async function loadGallery(galleryId: string): Promise<GalleryRecord> {
    const stored = (await query({ gallery: { __args: { id: galleryId }, ...GALLERY_SELECTION } })).gallery;
    if (stored === null || stored === undefined) {
      throw new Error("gallery not found");
    }
    if (stored.images.length > 0) {
      return stored;
    }
    const ensured = (await mutation({ ensureGalleryImages: { __args: { galleryId }, ...GALLERY_SELECTION } }))
      .ensureGalleryImages;
    if (ensured === null || ensured === undefined) {
      throw new Error("gallery not found");
    }
    return ensured;
  }

  // ensureSceneStreams returns embed pages; the `stream` query turns one into a playable URL.
  async function resolveStreamUrl(embedUrl: string): Promise<{ url: string; headers: Record<string, string> }> {
    const result = (await query({ stream: { __args: { url: embedUrl }, url: true, headers: { name: true, value: true } } }))
      .stream;
    if (result === null || result === undefined || !result.url) {
      throw new Error("could not resolve a playable stream url");
    }
    const headers: Record<string, string> = {};
    for (const header of result.headers ?? []) {
      headers[header.name] = header.value;
    }
    return { url: result.url, headers };
  }

  async function sceneTitleAndDuration(sceneId: string): Promise<{ title: string; durationSeconds?: number }> {
    const scene = (await query({ scene: { __args: { id: sceneId }, title: true, durationSeconds: true } })).scene;
    if (scene === null || scene === undefined) {
      return { title: `scene ${sceneId}` };
    }
    return { title: scene.title, durationSeconds: scene.durationSeconds ?? undefined };
  }

  return {
    baseUrl,
    imageProxyUrl: (remoteUrl) => `${baseUrl}/api/img?url=${encodeURIComponent(remoteUrl)}`,
    async listGalleries(search, limit, offset) {
      const data = await query({
        galleries: {
          __args: { search: search === "" ? null : search, limit, offset },
          id: true,
          title: true,
          coverPath: true,
          imageCount: true,
        },
      });
      return data.galleries;
    },
    async listScenes(search, limit, offset) {
      const data = await query({
        scenes: {
          __args: { search: search === "" ? null : search, limit, offset },
          id: true,
          title: true,
          posterPath: true,
          durationSeconds: true,
        },
      });
      return data.scenes;
    },
    async galleryImages(galleryId) {
      const gallery = await loadGallery(galleryId);
      return { title: gallery.title, urls: sortedImageUrls(gallery) };
    },
    async resolveScene(sceneId) {
      const streams = (await mutation({
        ensureSceneStreams: { __args: { sceneId }, url: true, kind: true, format: true, mimeType: true },
      })).ensureSceneStreams as StreamCandidate[];
      if (!streams || streams.length === 0) {
        throw new Error("scene has no playable streams");
      }
      const resolved = await resolveStreamUrl(pickBestStream(streams).url);
      const details = await sceneTitleAndDuration(sceneId);
      return { ...details, ...resolved };
    },
  };
}
