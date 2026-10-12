import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, externalIdOf, normalizeText, postIdFromUrl, postUrl, resolveUrl } from "./http.ts";
import { clockSeconds, viewsFromText } from "./values.ts";

// Anchors that carry hashtags or outside links; the title keeps them out.
const ANCHOR_TO_DROP = /<a\b(?:(?!<\/a>)[\s\S])*?\b(?:hash_link|extlink)\b(?:(?!<\/a>)[\s\S])*?<\/a>/gi;
// Badges and markers in the description: the NEW label, the category, the "+" buttons.
const ELEMENT_TO_DROP =
  /<(b|span)\b[^>]*\bclass=["'][^"']*\b(?:post_b_text_el|sub_cat_s|sub_el)\b[^"']*["'][^>]*>[\s\S]*?<\/\1>/gi;

/**
 * The post cards of a page (listings, search results and a post page's related
 * videos), one DiscoveredItem per distinct post. The main post of a post page is
 * not a card.
 */
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const card of root.querySelectorAll("div.post_el_small")) {
    if (card.classList.contains("post_el_post")) continue;
    const item = itemFromCard(card);
    if (item === undefined || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/**
 * The title of a post from its description HTML: the text without hashtags,
 * outside links, badges, categories and bare URLs.
 */
export function titleFromPostText(postTextHtml: string): string {
  const withoutMarkup = postTextHtml
    .replace(ANCHOR_TO_DROP, " ")
    .replace(ELEMENT_TO_DROP, " ")
    .replace(/<[^>]+>/g, " ");
  return normalizeText(decodeEntities(withoutMarkup).replace(/https?:\/\/\S+/g, " "));
}

/** One card: the post link gives the identity, the title and the media come from the card. */
function itemFromCard(card: HTMLElement): DiscoveredItem | undefined {
  const link = card.querySelector("a.js-pop");
  if (link === null) return undefined;
  const postId = postIdFromUrl(attributeOf(link, "href"));
  if (postId === "") return undefined;
  const postText = card.querySelector(".post_text");
  if (postText === null) return undefined;
  const title = titleFromPostText(postText.innerHTML);
  if (title === "") return undefined;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: postUrl(postId),
    external_id: externalIdOf(postId),
  };
  const poster = posterOf(card);
  if (poster !== "") item.poster_path = poster;
  const preview = previewOf(card);
  if (preview !== "") item.preview_video = preview;
  const durationElement = card.querySelector(".duration_small");
  if (durationElement !== null) {
    const seconds = clockSeconds(durationElement.text);
    if (seconds > 0) item.duration_seconds = seconds;
  }
  const timeElement = card.querySelector(".post_control_time");
  if (timeElement !== null) {
    const views = viewsFromText(timeElement.text);
    if (views !== undefined) item.view_count = views;
  }
  return item;
}

/** The card's still: posters are lazy-loaded, so the image is in data-src. */
function posterOf(card: HTMLElement): string {
  const image = card.querySelector("img.mini_post_vid_thumb");
  if (image === null) return "";
  const source = attributeOf(image, "data-src");
  if (source === "") return "";
  return resolveUrl(source);
}

/** The card's short preview clip, which the site plays on hover. */
function previewOf(card: HTMLElement): string {
  const video = card.querySelector("video");
  if (video === null) return "";
  const source = attributeOf(video, "src");
  if (source === "") return "";
  return resolveUrl(source);
}
