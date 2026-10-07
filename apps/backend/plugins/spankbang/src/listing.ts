import { HTMLElement } from "node-html-parser";
import { resolveUrl, videoCodeFromUrl, decodeEntities, parseDurationSeconds } from "./http.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

// Listing and search pages both open with an 8-item promoted block (a
// `[data-testid="video-list"]` that is identical across every page), then the
// real results follow in a larger container: `[data-testid="search-result"]`
// on search pages, a second `[data-testid="video-list"]` on browse pages. Pick
// the container with the most cards so the promoted block is never mistaken for
// results.
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  return parseCardsIn(resultsContainer(root));
}

// Parses every `video-item` card inside the given container, in document order.
export function parseCardsIn(container: HTMLElement): DiscoveredItem[] {
  const cards = container.querySelectorAll('div[data-testid="video-item"]');

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const card of cards) {
    const link = card.querySelector('a[href*="/video/"]');
    const href = link?.getAttribute("href") ?? "";
    const code = videoCodeFromUrl(href);
    if (!code || seen.has(code)) continue;
    seen.add(code);

    const title = cardTitle(card);
    if (!title) continue;

    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: resolveUrl(href),
      external_id: `spankbang-${code}`,
      poster_path: cardPoster(card),
      preview_video: cardPreview(card),
    };

    const durationSeconds = cardDurationSeconds(card);
    if (durationSeconds > 0) item.duration_seconds = durationSeconds;

    items.push(item);
  }

  return items;
}

function resultsContainer(root: HTMLElement): HTMLElement {
  const candidates = [
    ...root.querySelectorAll('[data-testid="search-result"]'),
    ...root.querySelectorAll('[data-testid="video-list"]'),
  ];

  let best: HTMLElement | null = null;
  let bestCount = -1;
  for (const candidate of candidates) {
    const count = candidate.querySelectorAll('div[data-testid="video-item"]').length;
    if (count > bestCount) {
      best = candidate;
      bestCount = count;
    }
  }

  return best ?? root;
}

function cardTitle(card: HTMLElement): string {
  const alt = card.querySelector("img")?.getAttribute("alt");
  if (alt) return decodeEntities(alt).trim();
  return "";
}

function cardPoster(card: HTMLElement): string | undefined {
  const src = card.querySelector("img")?.getAttribute("src");
  return src ? resolveUrl(src) : undefined;
}

// Cards carry a runtime badge like "15m" or "1h 5m".
function cardDurationSeconds(card: HTMLElement): number {
  const badge = card.querySelector('[data-testid="video-item-length"]');
  if (!badge) return 0;
  return parseDurationSeconds(badge.text.trim());
}

function cardPreview(card: HTMLElement): string | undefined {
  const src = card.querySelector("video source")?.getAttribute("data-src");
  return src ? resolveUrl(src) : undefined;
}
