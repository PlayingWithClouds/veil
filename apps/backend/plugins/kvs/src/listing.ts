import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem, StudioRef } from "@playingwithclouds/veil-sdk";
import type { SiteConfig } from "./sites.ts";
import {
  absoluteUrl,
  canonicalUrl,
  entityExternalId,
  entitySlug,
  entityUrl,
  videoExternalId,
  videoKey,
} from "./urls.ts";
import { cleanText, entityName, parseDurationText } from "./text.ts";

// KVS themes differ in card markup, so cards are found structurally rather
// than by class: every anchor to a video page marks a card, and the card is
// the largest ancestor of that anchor holding no other video's link.

const MAX_CARD_DEPTH = 6;

/** An element text that is nothing but a runtime: "5:59", "1:02:03", "5m:59s". */
const RUNTIME_ONLY = /^(?:\d{1,2}:)?\d{1,3}:\d{2}$|^\d{1,3}\s*m\s*:?\s*\d{1,2}\s*s$/i;

/** Image attributes carrying the real thumbnail, lazy-load ones first. */
const IMAGE_ATTRIBUTES = ["data-original", "data-src", "data-jpg", "data-webp", "src"];

/** Attributes on card elements pointing at a short hover clip. */
const PREVIEW_ATTRIBUTES = ["data-preview-custom", "data-preview", "data-trailer", "vthumb", "data-mp4"];

/**
 * The video cards of a listing page. KVS renders each list as a block whose
 * items sit in `#[custom_]list_videos_<name>_items`; when the page has such
 * blocks (other than related lists), only their cards count, so sidebars and
 * teasers stay out. Themes without them fall back to the whole page.
 */
export function parseListingCards(site: SiteConfig, root: HTMLElement): DiscoveredItem[] {
  const blocks = root
    .querySelectorAll('[id*="list_videos_"][id$="_items"]')
    .filter((block) => !block.id.includes("related"));
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const block of blocks) {
    for (const item of parseVideoCards(site, block)) {
      if (seen.has(item.external_id)) continue;
      seen.add(item.external_id);
      items.push(item);
    }
  }
  if (items.length > 0) return items;
  return parseVideoCards(site, root);
}

/** One discovered scene per distinct video linked from `root`, in page order. */
export function parseVideoCards(site: SiteConfig, root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const anchor of root.querySelectorAll("a[href]")) {
    const key = videoKey(site, anchor.getAttribute("href") ?? "");
    if (!key || seen.has(key)) continue;
    seen.add(key);

    const item = cardItem(site, cardOf(site, root, anchor, key), anchor, key);
    if (item) items.push(item);
  }
  return items;
}

/** The scene stub a card describes, or undefined when it shows no title. */
function cardItem(site: SiteConfig, card: HTMLElement, anchor: HTMLElement, key: string): DiscoveredItem | undefined {
  const title = cardTitle(site, card, key);
  if (!title) return undefined;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: canonicalUrl(site, anchor.getAttribute("href") ?? ""),
    external_id: videoExternalId(site, key),
  };
  const poster = cardPoster(site, card, key);
  if (poster) item.poster_path = poster;
  const preview = cardPreview(site, card);
  if (preview) item.preview_video = preview;
  const duration = cardDuration(card);
  if (duration > 0) item.duration_seconds = duration;
  const studio = cardStudio(site, card);
  if (studio) item.studio = studio;
  return item;
}

/**
 * Walks up from a video anchor while the parent still belongs to this video
 * alone, so the card spans the thumbnail, title and badges around the link.
 * The walk never leaves `root`, the block being parsed.
 */
function cardOf(site: SiteConfig, root: HTMLElement, anchor: HTMLElement, key: string): HTMLElement {
  let card = anchor;
  for (let depth = 0; depth < MAX_CARD_DEPTH && card !== root; depth++) {
    const parent = card.parentNode as HTMLElement | null;
    if (!parent || !parent.tagName) break;
    if (linksOtherVideo(site, parent, key)) break;
    card = parent;
  }
  return card;
}

/** Whether `element` links a video other than `key`. */
function linksOtherVideo(site: SiteConfig, element: HTMLElement, key: string): boolean {
  for (const anchor of element.querySelectorAll("a[href]")) {
    const otherKey = videoKey(site, anchor.getAttribute("href") ?? "");
    if (otherKey && otherKey !== key) return true;
  }
  return false;
}

/** The anchors inside `card` that link video `key`. */
function videoAnchors(site: SiteConfig, card: HTMLElement, key: string): HTMLElement[] {
  const anchors: HTMLElement[] = [];
  if (card.tagName === "A") anchors.push(card);
  for (const anchor of card.querySelectorAll("a[href]")) {
    if (videoKey(site, anchor.getAttribute("href") ?? "") === key) anchors.push(anchor);
  }
  return anchors;
}

/**
 * The card's title: the site's title element when configured, else a video
 * anchor's title attribute, the alt text of the thumbnail inside it (title
 * elements are often truncated with "..."), a title-classed element, and
 * finally the anchor text.
 */
function cardTitle(site: SiteConfig, card: HTMLElement, key: string): string {
  if (site.cardTitleSelector) {
    const configured = cleanText(card.querySelector(site.cardTitleSelector)?.text ?? "");
    if (configured) return configured;
  }
  const anchors = videoAnchors(site, card, key);
  for (const anchor of anchors) {
    const attribute = cleanText(anchor.getAttribute("title") ?? "");
    if (attribute) return attribute;
  }
  for (const anchor of anchors) {
    const alt = cleanText(anchor.querySelector("img[alt]")?.getAttribute("alt") ?? "");
    if (alt) return alt;
  }
  for (const element of card.querySelectorAll('[class*="title"]')) {
    const text = cleanText(element.text);
    if (text) return text;
  }
  for (const anchor of anchors) {
    const text = cleanText(anchor.text);
    if (text) return text;
  }
  return "";
}

/**
 * The thumbnail: the first image inside a link to the video (skipping model
 * avatars elsewhere in the card), else any card image or video poster.
 */
function cardPoster(site: SiteConfig, card: HTMLElement, key: string): string {
  for (const anchor of videoAnchors(site, card, key)) {
    for (const image of anchor.querySelectorAll("img")) {
      const source = imageSource(site, image);
      if (source) return source;
    }
  }
  for (const image of card.querySelectorAll("img")) {
    const source = imageSource(site, image);
    if (source) return source;
  }
  const poster = card.querySelector("video[poster]")?.getAttribute("poster");
  if (poster) return absoluteUrl(site, poster);
  return "";
}

/** The real image URL of an <img>, skipping inline placeholders. */
export function imageSource(site: SiteConfig, image: HTMLElement): string {
  for (const attribute of IMAGE_ATTRIBUTES) {
    const value = (image.getAttribute(attribute) ?? "").trim();
    if (!value || value.startsWith("data:") || /placeholder|blank\.gif/i.test(value)) continue;
    return absoluteUrl(site, value);
  }
  return "";
}

/** The hover clip a card element points at, when it is an mp4. */
function cardPreview(site: SiteConfig, card: HTMLElement): string {
  const elements = [card, ...card.querySelectorAll("*")];
  for (const element of elements) {
    for (const attribute of PREVIEW_ATTRIBUTES) {
      const value = (element.getAttribute(attribute) ?? "").trim();
      if (value && /\.mp4/i.test(value)) return absoluteUrl(site, value);
    }
  }
  return "";
}

/**
 * The runtime badge: a duration/time-classed element, else the first element
 * whose whole text is a runtime ("5:59"), else any runtime in the card text.
 */
function cardDuration(card: HTMLElement): number {
  for (const element of card.querySelectorAll('[class*="duration"], [class*="time"]')) {
    const seconds = parseDurationText(element.text);
    if (seconds > 0) return seconds;
  }
  for (const element of card.querySelectorAll("span, div, em, strong, small, b, i")) {
    const text = element.text.trim();
    if (!RUNTIME_ONLY.test(text)) continue;
    const seconds = parseDurationText(text);
    if (seconds > 0) return seconds;
  }
  return parseDurationText(card.text);
}

/** The channel a card credits, when it links one. */
function cardStudio(site: SiteConfig, card: HTMLElement): StudioRef | undefined {
  for (const prefix of site.channelPrefixes) {
    for (const anchor of card.querySelectorAll("a[href]")) {
      const slug = entitySlug(site, anchor.getAttribute("href") ?? "", [prefix]);
      const name = entityName(anchor);
      if (!slug || !name) continue;
      return {
        name,
        external_id: entityExternalId(site, "channel", slug),
        source_url: entityUrl(site, prefix, slug),
      };
    }
  }
  return undefined;
}
