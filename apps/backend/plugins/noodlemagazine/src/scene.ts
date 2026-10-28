import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, Scene, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { parseVideoCards } from "./cards.ts";
import { attributeOf, decodeEntities, externalIdOf, fetchWatchPage, normalizeText, videoIdFromUrl, watchUrl } from "./http.ts";

/** Fetches a watch page (from any mirror) and returns the scene as the scrape stream's root record. */
export async function scrapeScene(url: string): Promise<ScrapeResult[]> {
  const videoId = videoIdFromUrl(url);
  if (videoId === "") throw new Error(`noodlemagazine: no video in ${url}`);
  const page = await fetchWatchPage(videoId, false);
  return [{ type: "scene", scene: sceneFromWatchPage(parse(page.body), videoId) }];
}

/** The content of a meta tag, entities decoded; empty when the tag is missing. */
export function metaContent(root: HTMLElement, property: string): string {
  const meta = root.querySelector(`meta[property="${property}"]`);
  if (meta === null) return "";
  return decodeEntities(attributeOf(meta, "content"));
}

/**
 * The scene record of a watch page. Its one download is the watch page itself:
 * the real mp4 links are signed and expire within minutes, so the resolve
 * capability (routed by the mirrors' domains) fetches a fresh one when the
 * video plays.
 */
export function sceneFromWatchPage(root: HTMLElement, videoId: string): Scene {
  const scene: Scene = {
    type: "scene",
    external_id: externalIdOf(videoId),
    source_url: watchUrl(videoId),
    title: normalizeText(metaContent(root, "og:title")),
    tags: pageTags(root),
    downloads: [{ label: "Stream", url: watchUrl(videoId), format: "mp4" }],
    related: relatedFrom(root, videoId),
  };
  const date = metaContent(root, "ya:ovs:upload_date");
  if (/^\d{4}-\d{2}-\d{2}$/.test(date)) scene.date = date;
  const duration = Number(metaContent(root, "video:duration"));
  if (duration > 0) scene.duration_seconds = duration;
  const views = Number(metaContent(root, "ya:ovs:views_total"));
  if (views > 0) scene.view_count = views;
  const poster = metaContent(root, "og:image");
  if (poster !== "") scene.poster_path = poster;
  return scene;
}

/**
 * The video's tags: the tag links' English names where the page translates
 * them, else their base names; the video:tag meta list when the page has no
 * tag links.
 */
export function pageTags(root: HTMLElement): string[] {
  const names: string[] = [];
  for (const anchor of root.querySelectorAll("a.vtag")) {
    const name = normalizeText(tagName(anchor));
    if (name === "" || names.some((known) => known.toLowerCase() === name.toLowerCase())) continue;
    names.push(name);
  }
  if (names.length > 0) return names;
  return tagsFrom(metaContent(root, "video:tag"));
}

/** A tag link's name: the English translation when the page has one, else its base name. */
function tagName(anchor: HTMLElement): string {
  const english = englishTranslation(decodeEntities(attributeOf(anchor, "data-tag-translations")));
  if (english !== "") return english;
  return decodeEntities(attributeOf(anchor, "data-tag-base"));
}

/** The "en" entry of a tag translations JSON object, or "". */
function englishTranslation(json: string): string {
  if (json === "") return "";
  let translations: Record<string, unknown>;
  try {
    translations = JSON.parse(json) as Record<string, unknown>;
  } catch {
    return "";
  }
  if (typeof translations.en === "string") return translations.en;
  return "";
}

/** The comma-separated tag list of the video:tag meta tag. */
export function tagsFrom(content: string): string[] {
  return content
    .split(",")
    .map((tag) => normalizeText(tag))
    .filter((tag) => tag !== "");
}

/** The page's related videos, without the video itself. */
function relatedFrom(root: HTMLElement, videoId: string): DiscoveredItem[] {
  return parseVideoCards(root).filter((item) => videoIdFromUrl(item.source_url) !== videoId);
}
