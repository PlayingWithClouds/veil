import { HTMLElement } from "node-html-parser";
import { idFromUrl } from "./http.ts";
import { absoluteUrl, imageSrc } from "./html.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

// The site's HTML video grid renders one `div.mb` card per video: the thumbnail
// anchor wraps a lazy `img[data-src]`, `p.mbtit a` carries the title and
// `span.mbtim` the runtime. The external id matches the API's `eporner-<id>`.
export function parseVideoCards(container: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const card of container.querySelectorAll("div.mb")) {
    const link = card.querySelector("p.mbtit a");
    const href = link?.getAttribute("href") ?? "";
    const id = idFromUrl(href);
    if (!id || seen.has(id)) continue;
    seen.add(id);

    const title = link?.text.trim();
    if (!title) continue;

    const item: DiscoveredItem = {
      title,
      media_type: "scene",
      source_url: absoluteUrl(href),
      external_id: `eporner-${id}`,
      poster_path: imageSrc(card.querySelector(".mbimg img")),
    };
    const duration = parseClockDuration(card.querySelector("span.mbtim")?.text ?? "");
    if (duration > 0) item.duration_seconds = duration;
    items.push(item);
  }

  return items;
}

// Card runtimes render as "23:00", "120:06" or "1:02:03".
export function parseClockDuration(text: string): number {
  const parts = text.trim().split(":").map((part) => parseInt(part, 10));
  if (parts.length < 2 || parts.some((part) => !Number.isFinite(part))) return 0;
  return parts.reduce((total, part) => total * 60 + part, 0);
}
