import { HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { absoluteUrl, parseClockDuration, studioFromLink } from "./http.ts";
import { sceneExternalId, type Site } from "./sites.ts";

/**
 * Parses the video cards under `root` into DiscoveredItems, in page order.
 * Cards in navigation chrome (dropdown menus) and duplicates are skipped.
 */
export function parseCards(site: Site, root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const card of root.querySelectorAll(site.cardSelector)) {
    if (isNavigationCard(site, card)) continue;
    const item = parseCard(site, card);
    if (!item || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/** True when the card sits inside one of the site's header menus rather than a result grid. */
function isNavigationCard(site: Site, card: HTMLElement): boolean {
  for (const selector of site.cardExcludedAncestors) {
    if (card.closest(selector)) return true;
  }
  return false;
}

/** Builds one DiscoveredItem from a card, undefined when it has no usable link or title. */
function parseCard(site: Site, card: HTMLElement): DiscoveredItem | undefined {
  const link = card.querySelector(site.titleLinkSelector);
  const href = link ? link.getAttribute("href") : undefined;
  if (!link || !href) return undefined;

  const id = site.idFromUrl(absoluteUrl(site, href));
  const title = link.text.trim();
  if (!id || !title) return undefined;

  const thumbnail = card.querySelector(site.thumbnailSelector);
  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: site.sceneUrl(id),
    external_id: sceneExternalId(site, id),
  };

  const poster = thumbnailUrl(thumbnail);
  if (poster) item.poster_path = poster;
  const previewVideo = thumbnail ? thumbnail.getAttribute("data-mediabook") : undefined;
  if (previewVideo && previewVideo.startsWith("http")) item.preview_video = previewVideo;

  const durationElement = card.querySelector(site.durationSelector);
  if (durationElement) {
    const duration = parseClockDuration(durationElement.text);
    if (duration > 0) item.duration_seconds = duration;
  }

  const studio = studioFromLink(site, card.querySelector(site.uploaderLinkSelector));
  if (studio) item.studio = studio;
  return item;
}

/** The real thumbnail URL: lazy-load attributes win over the 1x1 base64 placeholder in `src`. */
function thumbnailUrl(image: HTMLElement | null): string | undefined {
  if (!image) return undefined;
  for (const attribute of ["data-image", "data-src", "data-poster", "src"]) {
    const value = image.getAttribute(attribute);
    if (value && value.startsWith("http")) return value;
  }
  return undefined;
}
