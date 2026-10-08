import { HttpError } from "@atlas/contracts/server";
import type { MediaLocation, ResolvedSource, SourceKind, SourceListing, SourceProvider } from "@atlas/contracts/server";
import { imageRef, parseVeilRef, sceneRef } from "./refs";
import type { VeilApi } from "./veilApi";

const STREAM_CACHE_MILLISECONDS = 10 * 60 * 1000;
const DEFAULT_RANDOM_COUNT = 20;
const IMAGES_PER_GALLERY = 4;
const SAMPLE_GALLERY_POOL = 100;
const SAMPLE_OFFSET_RANGE = 300;

const KINDS: SourceKind[] = [
  { id: "gallery", label: "Galleries", itemNoun: "gallery", browsable: true, ui: { layout: "grid", thumbnails: true } },
  { id: "scene", label: "Scenes", itemNoun: "scene", browsable: true, ui: { layout: "grid", thumbnails: true } },
  { id: "random", label: "Random", itemNoun: "image", browsable: false, ui: { sample: true } },
];

interface CachedStream {
  location: Extract<MediaLocation, { kind: "url" }>;
  expiresAt: number;
}

/** Veil proxies HLS through /api/stream/manifest, whose URL does not end in .m3u8. */
function isHlsPlaylist(url: string): boolean {
  const pathname = new URL(url).pathname.toLowerCase();
  return pathname.endsWith("/api/stream/manifest") || pathname.endsWith(".m3u8");
}

function shuffle<T>(values: T[]): T[] {
  const copy = [...values];
  for (let index = copy.length - 1; index > 0; index--) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    [copy[index], copy[swapIndex]] = [copy[swapIndex], copy[index]];
  }
  return copy;
}

function idParam(params: Record<string, unknown>): string {
  if (typeof params.id !== "string" || params.id === "") {
    throw new HttpError(400, "id required");
  }
  return params.id;
}

function countParam(params: Record<string, unknown>): number {
  const count = Number(params.count ?? DEFAULT_RANDOM_COUNT);
  if (!Number.isInteger(count) || count < 1) {
    throw new HttpError(400, "count must be a positive integer");
  }
  return count;
}

export function createVeilProvider(api: VeilApi, now: () => number = Date.now): SourceProvider {
  const streamCache = new Map<string, CachedStream>();

  async function listKind(kind: string, query: { search: string; limit: number; offset: number }): Promise<SourceListing> {
    if (kind === "gallery") {
      const galleries = await api.listGalleries(query.search, query.limit, query.offset);
      return {
        items: galleries.map((gallery) => ({
          id: gallery.id,
          title: gallery.title,
          thumbnail: gallery.coverPath ? api.imageProxyUrl(gallery.coverPath) : undefined,
          meta: { imageCount: gallery.imageCount },
        })),
      };
    }
    if (kind === "scene") {
      const scenes = await api.listScenes(query.search, query.limit, query.offset);
      return {
        items: scenes.map((scene) => ({
          id: scene.id,
          title: scene.title,
          thumbnail: scene.posterPath ? api.imageProxyUrl(scene.posterPath) : undefined,
          meta: { duration: scene.durationSeconds },
        })),
      };
    }
    throw new HttpError(400, `kind "${kind}" is not browsable`);
  }

  async function resolveGallery(params: Record<string, unknown>): Promise<ResolvedSource> {
    const { title, urls } = await api.galleryImages(idParam(params));
    return { label: title, items: urls.map((url) => ({ ref: imageRef(url), mediaKind: "image" })) };
  }

  async function resolveScene(params: Record<string, unknown>): Promise<ResolvedSource> {
    const sceneId = idParam(params);
    const ref = sceneRef(sceneId);
    const scene = await api.resolveScene(sceneId);
    streamCache.set(ref, {
      location: { kind: "url", url: scene.url, headers: scene.headers },
      expiresAt: now() + STREAM_CACHE_MILLISECONDS,
    });
    const sessionMeta: Record<string, unknown> = {};
    if (scene.durationSeconds !== undefined) {
      sessionMeta.duration = scene.durationSeconds;
    }
    return { label: scene.title, items: [{ ref, mediaKind: "video" }], sessionMeta };
  }

  async function sampleImageUrls(count: number): Promise<string[]> {
    const offset = Math.floor(Math.random() * SAMPLE_OFFSET_RANGE);
    let galleries = await api.listGalleries("", SAMPLE_GALLERY_POOL, offset);
    if (galleries.length === 0) {
      galleries = await api.listGalleries("", SAMPLE_GALLERY_POOL, 0);
    }
    const urls: string[] = [];
    for (const gallery of shuffle(galleries)) {
      if (urls.length >= count) {
        break;
      }
      urls.push(...(await sampleGalleryUrls(gallery.id)));
    }
    return shuffle(urls).slice(0, count);
  }

  async function sampleGalleryUrls(galleryId: string): Promise<string[]> {
    try {
      const { urls } = await api.galleryImages(galleryId);
      return shuffle(urls).slice(0, IMAGES_PER_GALLERY);
    } catch (error) {
      // A gallery that fails to scrape is skipped; sampling continues with the rest.
      return [];
    }
  }

  async function resolveRandom(params: Record<string, unknown>): Promise<ResolvedSource> {
    const urls = await sampleImageUrls(countParam(params));
    if (urls.length === 0) {
      throw new HttpError(404, "no images available");
    }
    return { label: `Random ${urls.length}`, items: urls.map((url) => ({ ref: imageRef(url), mediaKind: "image" })) };
  }

  async function locateScene(ref: string, sceneId: string): Promise<MediaLocation> {
    const cached = streamCache.get(ref);
    if (cached !== undefined && cached.expiresAt > now()) {
      return cached.location;
    }
    const scene = await api.resolveScene(sceneId);
    const location: CachedStream["location"] = { kind: "url", url: scene.url, headers: scene.headers };
    if (isHlsPlaylist(scene.url)) {
      location.format = "hls";
    }
    streamCache.set(ref, { location, expiresAt: now() + STREAM_CACHE_MILLISECONDS });
    return location;
  }

  return {
    id: "veil",
    kinds: () => KINDS,
    list: listKind,
    async resolve(kind, params) {
      if (kind === "gallery") {
        return resolveGallery(params);
      }
      if (kind === "scene") {
        return resolveScene(params);
      }
      if (kind === "random") {
        return resolveRandom(params);
      }
      throw new HttpError(400, `unknown kind: ${kind}`);
    },
    async locate(ref) {
      const parsed = parseVeilRef(ref);
      if (parsed === undefined) {
        throw new HttpError(400, `not a veil ref: ${ref}`);
      }
      if (parsed.type === "image") {
        return { kind: "url", url: api.imageProxyUrl(parsed.remoteUrl) };
      }
      return locateScene(ref, parsed.sceneId);
    },
  };
}
