import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { absoluteUrl, watchExternalId, watchIdFromHref, watchUrl } from "./site.ts";
import { attributeOf, clockSeconds, cleanText, countText } from "./text.ts";

/** Most related videos kept from a video page. */
const MAX_RELATED = 40;

/** The scene stubs of the video cards under `root`, in page order, one per video. */
export function cardsFromRoot(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll("a.item_link")) {
    const item = cardItem(anchor);
    if (!item || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/** The video cards of a video page's related list, without the video itself. */
export function relatedFromRoot(root: HTMLElement, ownExternalId: string): DiscoveredItem[] {
  return cardsFromRoot(root)
    .filter((item) => item.external_id !== ownExternalId)
    .slice(0, MAX_RELATED);
}

/** The scene stub one card describes, or undefined when it links no video or shows no title. */
function cardItem(anchor: HTMLElement): DiscoveredItem | undefined {
  const id = watchIdFromHref(attributeOf(anchor, "href"));
  if (id === undefined) return undefined;
  const title = cardTitle(anchor);
  if (!title) return undefined;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: watchUrl(id),
    external_id: watchExternalId(id),
  };
  const poster = cardPoster(anchor);
  if (poster) item.poster_path = poster;
  const preview = cardPreview(anchor);
  if (preview) item.preview_video = preview;
  const duration = clockSeconds(elementText(anchor, ".m_time"));
  if (duration > 0) item.duration_seconds = duration;
  const views = countText(elementText(anchor, ".m_views"));
  if (views > 0) item.view_count = views;
  return item;
}

/** The card's title, else its thumbnail's alt text. */
function cardTitle(anchor: HTMLElement): string {
  const title = cleanText(elementText(anchor, ".title"));
  if (title) return title;
  const image = anchor.querySelector("img");
  if (!image) return "";
  return cleanText(attributeOf(image, "alt"));
}

/** The thumbnail URL, from its lazy-load attribute when it has one; "" for placeholders. */
function cardPoster(anchor: HTMLElement): string {
  const image = anchor.querySelector("img");
  if (!image) return "";
  let source = attributeOf(image, "data-src");
  if (source === "") source = attributeOf(image, "src");
  if (source === "" || source.startsWith("data:")) return "";
  return absoluteUrl(source);
}

/** The short hover clip the card names in its data-trailer_url, or "". */
function cardPreview(anchor: HTMLElement): string {
  const element = anchor.querySelector("[data-trailer_url]");
  if (!element) return "";
  return absoluteUrl(attributeOf(element, "data-trailer_url"));
}

/** The text of the first element matching `selector` under `element`, or "". */
function elementText(element: HTMLElement, selector: string): string {
  const found = element.querySelector(selector);
  if (!found) return "";
  return found.text;
}
