import { HTMLElement } from "node-html-parser";
import { resolveUrl, videoCodeFromUrl, decodeEntities, parseClockDuration } from "./http.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

// The catalog grid renders one `div.thumbnail.group` per video. Each card links
// to /en/<code>, carries the poster in `img[data-src]` (a fourhoi.com cover), a
// short hover clip in `video[data-src]`, and the display title in the caption
// link `a.text-secondary`.
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const card of root.querySelectorAll(".thumbnail.group")) {
    const link = cardLink(card);
    const href = link?.getAttribute("href") ?? "";
    const code = videoCodeFromUrl(href);
    if (!code || seen.has(code)) continue;
    seen.add(code);

    const title = link ? decodeEntities(link.text.trim()) : "";
    if (!title) continue;

    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: resolveUrl(href),
      external_id: `missav-${code}`,
      poster_path: cardPoster(card),
      preview_video: cardPreview(card),
    };
    const duration = cardDuration(card);
    if (duration > 0) item.duration_seconds = duration;
    items.push(item);
  }

  return items;
}

function cardLink(card: HTMLElement): HTMLElement | null {
  return card.querySelector("a.text-secondary") ?? card.querySelector('a[href*="/en/"]');
}

function cardPoster(card: HTMLElement): string | undefined {
  const src = card.querySelector("img[data-src]")?.getAttribute("data-src");
  return src ? resolveUrl(src) : undefined;
}

// The runtime renders as an overlay span with clock-formatted text ("2:07:59").
// Match on the text shape rather than utility classes, which churn.
function cardDuration(card: HTMLElement): number {
  for (const span of card.querySelectorAll("span")) {
    const seconds = parseClockDuration(span.text);
    if (seconds > 0) return seconds;
  }
  return 0;
}

function cardPreview(card: HTMLElement): string | undefined {
  const src = card.querySelector("video[data-src]")?.getAttribute("data-src");
  return src ? resolveUrl(src) : undefined;
}
