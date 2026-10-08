import { baseUrl, videoPageUrl } from "./http.ts";
import type { DiscoveredItem, StudioRef } from "@playingwithclouds/veil-sdk";

// One entry in the videos2 API `videos` array. Only the fields the plugin reads
// are typed; the payload carries many more.
export interface ApiVideo {
  video_id: string;
  title: string;
  dir: string;
  duration?: string;
  post_date?: string;
  scr?: string | null;
  pv?: string | null;
  content_source_name?: string | null;
  user_id?: string | null;
  display_name?: string | null;
}

export function videoToDiscoveredItem(host: string, video: ApiVideo): DiscoveredItem | null {
  const id = String(video.video_id ?? "").trim();
  const title = (video.title ?? "").trim();
  if (!id || !title) return null;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: videoUrl(host, id, video.dir),
    external_id: `txxx-${id}`,
    date: normalizeDate(video.post_date),
    poster_path: withScheme(video.scr),
    preview_video: withScheme(video.pv),
  };

  const durationSeconds = parseDurationSeconds(video.duration);
  if (durationSeconds > 0) item.duration_seconds = durationSeconds;

  const studio = listingStudio(host, video);
  if (studio) item.studio = studio;

  return item;
}

/**
 * The publisher a listing entry names: its content source, else the uploading
 * member (same identity the video page credits them with). Listings carry no
 * channel, so a channel video's stub is re-credited when its page is visited.
 */
export function listingStudio(host: string, video: ApiVideo): StudioRef | undefined {
  const sourceName = (video.content_source_name ?? "").trim();
  if (sourceName) return { name: sourceName };

  const userId = String(video.user_id ?? "").trim();
  const displayName = (video.display_name ?? "").trim();
  if (!userId || !displayName) return undefined;
  return {
    name: displayName,
    external_id: hostScopedId(host, "member", userId),
    source_url: `${baseUrl(host)}/members/${userId}/`,
  };
}

/**
 * An external id for a site-scoped entity. Channel and member ids are shared
 * network-wide, but each site gets its own studio (parented to that site), so
 * the host is part of the identity.
 */
export function hostScopedId(host: string, kind: string, id: string): string {
  return `txxx-${kind}-${id}@${host}`;
}

// Durations render as "MM:SS" or "HH:MM:SS".
export function parseDurationSeconds(duration: string | undefined): number {
  if (!duration) return 0;
  const parts = duration.split(":").map((part) => parseInt(part, 10));
  if (parts.some((part) => !Number.isFinite(part))) return 0;
  return parts.reduce((total, part) => total * 60 + part, 0);
}

export function videoUrl(host: string, id: string, dir?: string): string {
  return videoPageUrl(host, id, dir);
}

// Preview and screenshot URLs arrive scheme-relative (e.g. "vp1.txxx.com/...").
function withScheme(url: string | null | undefined): string | undefined {
  if (!url) return undefined;
  if (url.startsWith("http")) return url;
  if (url.startsWith("//")) return "https:" + url;
  return "https://" + url;
}

// post_date is "YYYY-MM-DD HH:MM:SS"; keep the date portion as ISO 8601.
function normalizeDate(postDate: string | undefined): string | undefined {
  if (!postDate) return undefined;
  const match = postDate.match(/^(\d{4}-\d{2}-\d{2})/);
  return match ? match[1] : undefined;
}
