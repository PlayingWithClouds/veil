import { API_BASE, UA, toIsoDate, toRating, parseKeywords, embedUrl } from "./http.ts";
import type { DiscoveredItem, Scene, Image } from "@playingwithclouds/veil-sdk";

// One video as returned by the eporner v2 API.
export interface EpVideo {
  id: string;
  title: string;
  keywords?: string;
  views?: number;
  rate?: string;
  url: string;
  added?: string;
  length_sec?: number;
  embed?: string;
  default_thumb?: { src?: string };
  thumbs?: { src?: string }[];
}

// eporner search orders. "latest" backs the browse feed; searches use relevance.
export type EpOrder =
  | "latest"
  | "longest"
  | "shortest"
  | "top-rated"
  | "most-popular"
  | "top-weekly"
  | "top-monthly";

async function apiGet(path: string, params: Record<string, string | number>): Promise<any> {
  const query = new URLSearchParams({
    ...Object.fromEntries(Object.entries(params).map(([k, v]) => [k, String(v)])),
    thumbsize: "big",
    format: "json",
  }).toString();
  const res = await fetch(`${API_BASE}${path}?${query}`, {
    headers: { "User-Agent": UA },
    signal: AbortSignal.timeout(15_000),
  });
  if (!res.ok) throw new Error(`eporner api HTTP ${res.status} for ${path}`);
  return res.json();
}

// video/search backs both search and browse; an empty query with order=latest is
// the newest-videos feed.
export async function apiSearch(
  query: string,
  perPage: number,
  page: number,
  order: EpOrder,
): Promise<EpVideo[]> {
  const data = await apiGet("/video/search/", {
    query,
    per_page: clampPerPage(perPage),
    page: Math.max(1, page),
    order,
    gay: 0,
    lq: 1,
  });
  return Array.isArray(data?.videos) ? data.videos : [];
}

// video/id returns a single video's full detail (top-level fields).
export async function apiVideo(id: string): Promise<EpVideo | null> {
  const data = await apiGet("/video/id/", { id });
  if (data?.video) return data.video as EpVideo;
  if (data?.id) return data as EpVideo;
  return null;
}

// The API caps per_page at 1000 and rejects tiny values oddly; keep it sane.
function clampPerPage(perPage: number): number {
  if (!Number.isFinite(perPage) || perPage <= 0) return 30;
  return Math.min(1000, Math.max(1, Math.floor(perPage)));
}

export function toDiscoveredItem(video: EpVideo): DiscoveredItem {
  return {
    title: video.title,
    media_type: "scene",
    source_url: video.url,
    external_id: `eporner-${video.id}`,
    date: toIsoDate(video.added),
    poster_path: video.default_thumb?.src,
    preview_images: previewFrames(video).slice(0, 12),
    duration_seconds: video.length_sec,
    rating: toRating(video.rate),
    view_count: video.views,
    // The embed page is a deterministic source, so the browse stub is playable
    // without a scrape (resolved to a CDN mp4 on demand). Mirrors toScene.
    downloads: [{ label: "eporner", url: video.embed || embedUrl(video.id) }],
  };
}

export function toScene(video: EpVideo): Scene {
  const poster = video.default_thumb?.src;
  const images: Image[] = poster ? [{ type: "poster", file_path: poster }] : [];
  return {
    type: "scene",
    external_id: `eporner-${video.id}`,
    source_url: video.url,
    title: video.title,
    date: toIsoDate(video.added),
    duration_seconds: video.length_sec,
    rating: toRating(video.rate),
    view_count: video.views,
    tags: parseKeywords(video.keywords, video.title),
    poster_path: poster,
    preview_images: previewFrames(video),
    images,
    // The embed page carries the hash + CDN source the resolver needs.
    downloads: [{ label: "eporner", url: video.embed || embedUrl(video.id) }],
  };
}

function previewFrames(video: EpVideo): string[] {
  return (video.thumbs ?? [])
    .map((thumb) => thumb.src)
    .filter((src): src is string => typeof src === "string" && src.length > 0);
}
