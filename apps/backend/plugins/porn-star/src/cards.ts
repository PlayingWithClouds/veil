import { parse } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, durationSeconds, resolveUrl } from "./http.ts";

type Root = ReturnType<typeof parse>;
type Element = NonNullable<ReturnType<Root["querySelector"]>>;

// A video page path: /video/<site>/<id>/, the id part may contain a dash.
const VIDEO_PATH_PATTERN = /^\/video\/([^/]+)\/([^/]+)\/?$/;
// The thumbnail alt text ends with "(Video duration: 37:06)" on the catalog.
const DURATION_PATTERN = /Video duration:\s*([\d:]+)\)/;

/** The stable external id of a video, from its /video/<site>/<id>/ path. */
export function externalIdFromPath(pathname: string): string | undefined {
  const match = pathname.match(VIDEO_PATH_PATTERN);
  if (match === null) return undefined;
  return `porn-star-${match[1]}-${match[2]}`;
}

/**
 * The video cards of a page (catalog, or a video's related list) as stubs, in page
 * order. A query keeps the cards whose performer names or thumbnail description contain it.
 */
export function cardsFromRoot(root: Root, query?: string): DiscoveredItem[] {
  let needle = "";
  if (query !== undefined) needle = query.trim().toLowerCase();
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll(".thumbnails-figures-videos a")) {
    const item = stubFromAnchor(anchor);
    if (item === undefined || seen.has(item.external_id)) continue;
    if (needle !== "" && !matchesNeedle(anchor, item, needle)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/** Whether a card's performer names or thumbnail description contain the (lower-case) needle. */
function matchesNeedle(anchor: Element, item: DiscoveredItem, needle: string): boolean {
  let description = "";
  const image = anchor.querySelector("img");
  if (image !== null) description = attributeOf(image, "alt");
  return `${item.title} ${description}`.toLowerCase().includes(needle);
}

/** The runtime in seconds named in a card's thumbnail alt text, if any. */
export function durationFromAlt(altText: string): number | undefined {
  const match = altText.match(DURATION_PATTERN);
  if (match === null) return undefined;
  return durationSeconds(match[1]);
}

/** The stub for one video card anchor, or undefined when the anchor is not a video link. */
function stubFromAnchor(anchor: ReturnType<Root["querySelector"]>): DiscoveredItem | undefined {
  if (anchor === null) return undefined;
  const pathname = new URL(resolveUrl(attributeOf(anchor, "href"))).pathname;
  const externalId = externalIdFromPath(pathname);
  if (externalId === undefined) return undefined;

  const item: DiscoveredItem = {
    title: decodeEntities(attributeOf(anchor, "title")),
    media_type: "scene",
    source_url: resolveUrl(pathname),
    external_id: externalId,
  };
  const image = anchor.querySelector("img");
  if (image !== null) addImageFields(item, image);
  return item;
}

/** Copies the poster and runtime of a card's thumbnail onto its stub. */
function addImageFields(item: DiscoveredItem, image: NonNullable<ReturnType<Root["querySelector"]>>): void {
  const poster = attributeOf(image, "src");
  if (poster !== "") item.poster_path = resolveUrl(poster);
  const duration = durationFromAlt(attributeOf(image, "alt"));
  if (duration !== undefined) item.duration_seconds = duration;
}
