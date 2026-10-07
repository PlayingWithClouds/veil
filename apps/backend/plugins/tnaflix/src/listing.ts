import { HTMLElement } from "node-html-parser";
import { resolveUrl, videoIdFromUrl, decodeEntities, parseDurationSeconds } from "./http.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

// Browse and search share one video-card grid. Each card is a `div[data-vid]`
// holding an `a.video-thumb` (video href, hover trailer in `data-trailer`, the
// poster img and a `.video-duration` chip) plus a sibling `a.video-title`.
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const card of root.querySelectorAll("div[data-vid]")) {
    const thumb = card.querySelector("a.video-thumb");
    const href = thumb?.getAttribute("href") ?? "";
    const id = videoIdFromUrl(href);
    if (!id || seen.has(id)) continue;
    seen.add(id);

    const title = cardTitle(card);
    if (!title) continue;

    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: resolveUrl(href),
      external_id: `tnaflix-${id}`,
      poster_path: cardPoster(card),
      preview_video: cardTrailer(thumb),
    };

    const duration = cardDuration(card);
    if (duration > 0) item.duration_seconds = duration;

    items.push(item);
  }

  return items;
}

function cardTitle(card: HTMLElement): string {
  const link = card.querySelector("a.video-title");
  if (link) return decodeEntities(link.text).trim();
  const alt = card.querySelector("img")?.getAttribute("alt");
  if (alt) return decodeEntities(alt).trim();
  return "";
}

// Lazy-loaded cards ship a placeholder in `src` and the real poster in
// `data-src`; eagerly loaded cards carry it in `src` directly.
function cardPoster(card: HTMLElement): string | undefined {
  const img = card.querySelector("img");
  if (!img) return undefined;
  const lazy = img.getAttribute("data-src");
  if (lazy) return resolveUrl(lazy);
  const src = img.getAttribute("src");
  if (src && !src.includes("placeholder")) return resolveUrl(src);
  return undefined;
}

function cardDuration(card: HTMLElement): number {
  const chip = card.querySelector(".video-duration");
  if (!chip) return 0;
  return parseDurationSeconds(chip.text);
}

function cardTrailer(thumb: HTMLElement | null): string | undefined {
  const trailer = thumb?.getAttribute("data-trailer");
  if (trailer) return decodeEntities(trailer);
  return undefined;
}
