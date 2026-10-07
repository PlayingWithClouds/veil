import { HTMLElement } from "node-html-parser";
import { resolveUrl, videoIdFromUrl, decodeEntities, durationSecondsFromLabel } from "./http.ts";
import type { DiscoveredItem, StudioRef } from "@playingwithclouds/veil-sdk";

/** The channel or user a listed video links to (`landing` in the page state). */
interface ThumbLanding {
  type?: string;
  name?: string;
  logo?: string | null;
  link?: string;
}

/** One video card as the page's `window.initials` state describes it. */
export interface VideoThumb {
  pageURL?: string;
  thumbURL?: string;
  duration?: number;
  created?: number;
  landing?: ThumbLanding;
}

// The /newest listing and /search results share one video-card grid. Each card's
// anchor is `a.video-thumb__image-container` carrying the video href, the title
// in `aria-label`, the poster in the child `img[src]`, and a short preview clip
// in `data-previewvideo`. Parse straight off that anchor — the separate title
// link is only present for a subset of cards.
//
// Only the first cards are fully server-rendered; the rest are skeletons with no
// image. The page state (`thumbs`, keyed by video id) fills in the poster,
// runtime, upload date and channel the markup lacks.
export function parseVideoCards(root: HTMLElement, thumbs: Map<string, VideoThumb> = new Map()): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const anchor of root.querySelectorAll("a.video-thumb__image-container")) {
    const href = anchor.getAttribute("href") ?? "";
    if (!href.includes("/videos/")) continue;

    const id = videoIdFromUrl(href);
    if (!id || seen.has(id)) continue;
    seen.add(id);

    const title = cardTitle(anchor);
    if (!title) continue;

    const thumb = thumbs.get(id);
    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: resolveUrl(href),
      external_id: `xhamster-${id}`,
      poster_path: cardPoster(anchor) ?? thumb?.thumbURL,
      preview_video: anchor.getAttribute("data-previewvideo") ?? undefined,
    };
    const duration = cardDurationSeconds(anchor) || thumb?.duration || 0;
    if (duration > 0) item.duration_seconds = duration;
    const date = thumbDate(thumb);
    if (date) item.date = date;
    const studio = landingStudio(thumb?.landing);
    if (studio) item.studio = studio;
    items.push(item);
  }

  return items;
}

/**
 * Every video card in the page state, keyed by video id. The state nests its
 * lists differently per page type, so this walks it for objects linking to a
 * /videos/ page rather than hard-coding paths.
 */
export function thumbsFromInitials(initials: unknown): Map<string, VideoThumb> {
  const thumbs = new Map<string, VideoThumb>();
  collectThumbs(initials, thumbs);
  return thumbs;
}

/** Recursive step of thumbsFromInitials. */
function collectThumbs(node: unknown, thumbs: Map<string, VideoThumb>): void {
  if (Array.isArray(node)) {
    for (const child of node) collectThumbs(child, thumbs);
    return;
  }
  if (!node || typeof node !== "object") return;
  const candidate = node as VideoThumb;
  if (typeof candidate.pageURL === "string" && candidate.pageURL.includes("/videos/")) {
    const id = videoIdFromUrl(candidate.pageURL);
    if (id && !thumbs.has(id)) thumbs.set(id, candidate);
    return;
  }
  for (const value of Object.values(node)) collectThumbs(value, thumbs);
}

/** The upload date as ISO YYYY-MM-DD, from the unix-seconds `created` timestamp. */
function thumbDate(thumb: VideoThumb | undefined): string | undefined {
  if (!thumb || typeof thumb.created !== "number" || thumb.created <= 0) return undefined;
  return new Date(thumb.created * 1000).toISOString().slice(0, 10);
}

/**
 * The card's channel or uploader as a StudioRef, keyed by the same page URL
 * the video page's uploader chip uses (listing links point at
 * /users/profiles/<name>/videos, the chip at /users/<name>).
 */
export function landingStudio(landing: ThumbLanding | undefined): StudioRef | undefined {
  if (!landing || !landing.name || !landing.link) return undefined;
  const studio: StudioRef = { name: landing.name, source_url: canonicalLandingUrl(landing.link) };
  if (landing.logo) studio.image_path = landing.logo;
  return studio;
}

/** Rewrites a user's listing link to their profile URL; channel links pass through. */
function canonicalLandingUrl(link: string): string {
  const userMatch = link.match(/^(https?:\/\/[^/]+)\/users\/profiles\/([^/?#]+)/);
  if (userMatch) return `${userMatch[1]}/users/${userMatch[2]}`;
  return link;
}

function cardTitle(anchor: HTMLElement): string {
  const label = anchor.getAttribute("aria-label");
  if (label) return decodeEntities(label).trim();
  const alt = anchor.querySelector("img")?.getAttribute("alt");
  return alt ? decodeEntities(alt).trim() : "";
}

// The runtime badge sits inside the anchor as `[data-role="video-duration"]`
// with "MM:SS" / "H:MM:SS" text.
function cardDurationSeconds(anchor: HTMLElement): number {
  const badge = anchor.querySelector('[data-role="video-duration"]');
  if (!badge) return 0;
  return durationSecondsFromLabel(badge.text);
}

function cardPoster(anchor: HTMLElement): string | undefined {
  const img = anchor.querySelector("img.thumb-image-container__image, img");
  const src = img?.getAttribute("src");
  return src ? resolveUrl(src) : undefined;
}
