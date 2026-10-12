import { parse, type HTMLElement } from "node-html-parser";
import { fetchHtml, type Download, type Scene, type ScrapeResult } from "@playingwithclouds/veil-sdk";
import { relatedFromRoot } from "./cards.ts";
import { BASE_URL, watchExternalId, watchIdFromHref, watchUrl } from "./site.ts";
import { attributeOf, cleanText, decodeEntities } from "./text.ts";

/** One file of a video page's player playlist. */
export interface PlaylistSource {
  /** the CDN URL; its signature expires, so it is only read on resolve */
  file: string;
  /** the resolution without its "p", e.g. "720" */
  label: string;
}

/** Scrapes a video page into its scene; [] for a URL that is no video page. */
export async function scrapeScene(url: string): Promise<ScrapeResult[]> {
  const id = watchIdFromHref(url);
  if (id === undefined) return [];
  const html = await fetchHtml(watchUrl(id));
  return [{ type: "scene", scene: sceneFromPage(id, html) }];
}

/** The scene described by a video page's HTML. */
export function sceneFromPage(id: string, html: string): Scene {
  const root = parse(html);
  const externalId = watchExternalId(id);
  const scene: Scene = {
    type: "scene",
    external_id: externalId,
    source_url: watchUrl(id),
    title: pageTitle(root),
    studio: { name: "mat6tube", source_url: BASE_URL },
    tags: pageTags(root),
    downloads: downloadsOf(id, playlistSources(html)),
    related: relatedFromRoot(root, externalId),
  };
  setMetaFields(scene, root);
  return scene;
}

/**
 * The player playlist's files (`window.playlist = {…}`), best quality first;
 * [] when the page carries none.
 */
export function playlistSources(html: string): PlaylistSource[] {
  const match = html.match(/window\.playlist\s*=\s*(\{.*\})\s*;/);
  if (!match) return [];
  let playlist: { sources?: unknown };
  try {
    playlist = JSON.parse(match[1]) as { sources?: unknown };
  } catch {
    return [];
  }
  if (!Array.isArray(playlist.sources)) return [];
  return (playlist.sources as PlaylistSource[]).filter(
    (source) => typeof source.file === "string" && typeof source.label === "string" && source.label !== ""
  );
}

/** The video's title from OpenGraph, else the document title. */
function pageTitle(root: HTMLElement): string {
  const title = metaContent(root, "og:title");
  if (title) return title;
  const heading = root.querySelector("title");
  if (!heading) return "";
  return cleanText(heading.text);
}

/** Date, runtime, views and poster from the page's meta tags. */
function setMetaFields(scene: Scene, root: HTMLElement): void {
  const date = metaContent(root, "ya:ovs:upload_date");
  if (/^\d{4}-\d{2}-\d{2}$/.test(date)) scene.date = date;
  const duration = parseInt(metaContent(root, "video:duration"), 10);
  if (duration > 0) scene.duration_seconds = duration;
  const views = parseInt(metaContent(root, "ya:ovs:views_total"), 10);
  if (views > 0) scene.view_count = views;
  const poster = metaContent(root, "og:image");
  if (poster) {
    scene.poster_path = poster;
    scene.images = [{ type: "poster", file_path: poster }];
  }
}

/** The content of the OpenGraph/video meta tag `property`, or "". */
function metaContent(root: HTMLElement, property: string): string {
  const meta = root.querySelector(`meta[property="${property}"]`);
  if (!meta) return "";
  return decodeEntities(cleanText(attributeOf(meta, "content")));
}

/** The video's tags: their English names where the page translates them, else as written. */
function pageTags(root: HTMLElement): string[] {
  const names: string[] = [];
  for (const anchor of root.querySelectorAll("a.vtag")) {
    const name = tagName(anchor);
    if (name === "" || names.some((known) => known.toLowerCase() === name.toLowerCase())) continue;
    names.push(name);
  }
  return names;
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

/**
 * One download per playlist file. Each points back at the page with the
 * quality as fragment, because the CDN links carry expiring signatures;
 * resolve re-reads the page and follows that quality's file.
 */
function downloadsOf(id: string, sources: PlaylistSource[]): Download[] {
  return sources.map((source) => {
    const quality = `${source.label}p`;
    return { label: quality, quality, url: `${watchUrl(id)}#${quality}`, format: "mp4" };
  });
}
