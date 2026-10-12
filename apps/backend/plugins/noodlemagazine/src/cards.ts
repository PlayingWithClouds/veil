import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { attributeOf, normalizeText, videoIdFromUrl, watchUrl } from "./http.ts";
import { clockSeconds, countFromText } from "./values.ts";

/**
 * The video cards on a page (listings, search results and the related videos
 * of a watch page), one DiscoveredItem per distinct video.
 */
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const card of root.querySelectorAll("div.item")) {
    const item = itemFromCard(card);
    if (item === undefined || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/** One card: the watch link gives the identity, the title and poster come from the card. */
function itemFromCard(card: HTMLElement): DiscoveredItem | undefined {
  const link = card.querySelector("a.item_link");
  if (link === null) return undefined;
  const videoId = videoIdFromUrl(attributeOf(link, "href"));
  if (videoId === "") return undefined;
  const titleElement = card.querySelector(".i_info .title");
  if (titleElement === null) return undefined;
  const title = normalizeText(titleElement.text);
  if (title === "") return undefined;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: watchUrl(videoId),
    external_id: `noodlemagazine-${videoId}`,
  };
  const image = card.querySelector("img");
  if (image !== null) {
    const poster = posterOf(image);
    if (poster !== "") item.poster_path = poster;
  }
  const durationElement = card.querySelector(".m_time");
  if (durationElement !== null) {
    const seconds = clockSeconds(durationElement.text);
    if (seconds > 0) item.duration_seconds = seconds;
  }
  const viewsElement = card.querySelector(".m_views");
  if (viewsElement !== null) item.view_count = countFromText(viewsElement.text);
  return item;
}

/**
 * The real poster of a card. Posters are lazy-loaded: src holds a gif placeholder
 * and data-src the image, so the placeholder is never used.
 */
function posterOf(image: HTMLElement): string {
  const lazySource = attributeOf(image, "data-src");
  if (lazySource !== "") return lazySource;
  const source = attributeOf(image, "src");
  if (source.startsWith("data:")) return "";
  return source;
}
