import { HTMLElement } from "node-html-parser";
import { resolveUrl, videoIdFromUrl, parseDurationSeconds } from "./http.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

// Browse and search share the same video-card grid: each card is a
// `h3.meta-data-title > a` linking to /hdporn/<id>-<slug>.html, with the poster
// carried by the sibling `img` inside the card's `a.image`. Each card also ships
// a `preload_<id>` script array of 10 landscape frames used for the hover strip.
export function parseVideoCards(root: HTMLElement, html: string): DiscoveredItem[] {
  const framesById = parsePreloadFrames(html);
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const heading of root.querySelectorAll("h3.meta-data-title")) {
    const link = heading.querySelector('a[href*="/hdporn/"]');
    const href = link?.getAttribute("href") ?? "";
    const id = videoIdFromUrl(href);
    if (!id || seen.has(id)) continue;
    seen.add(id);

    const title = link?.text.trim();
    if (!title) continue;

    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: resolveUrl(href),
      external_id: `hqporner-${id}`,
      poster_path: findPoster(heading, id),
      preview_images: framesById.get(id) ?? [],
    };
    const duration = findDurationSeconds(heading);
    if (duration > 0) item.duration_seconds = duration;
    items.push(item);
  }

  return items;
}

// The listing embeds one `var preload_<id>=["//host/..._1.jpg", ...];` array per
// card. Collect those frames keyed by video id for the hover thumbnail strip.
function parsePreloadFrames(html: string): Map<string, string[]> {
  const framesById = new Map<string, string[]>();
  const arrayPattern = /var\s+preload_(\d+)\s*=\s*\[([^\]]*)\]/g;

  for (const match of html.matchAll(arrayPattern)) {
    const id = match[1];
    const urls = Array.from(match[2].matchAll(/"([^"]+)"/g)).map((m) => resolveUrl(m[1]));
    if (urls.length > 0) framesById.set(id, urls);
  }

  return framesById;
}

// Each card shows its runtime in a `<span class="fa-clock-o meta-data">` badge.
function findDurationSeconds(heading: HTMLElement): number {
  const card = heading.closest("section");
  const badge = card?.querySelector("span.fa-clock-o");
  if (!badge) return 0;
  return parseDurationSeconds(badge.text);
}

// The poster lives in the card's image anchor, which precedes the title
// heading. Walk up to the enclosing card section, then read `img#cover_<id>`.
function findPoster(heading: HTMLElement, id: string): string | undefined {
  const card = heading.closest("section");
  const img = card?.querySelector(`img#cover_${id}, a.image img`);
  const src = img?.getAttribute("src");
  return src ? resolveUrl(src) : undefined;
}
