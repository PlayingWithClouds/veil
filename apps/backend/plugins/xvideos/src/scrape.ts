import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, Image, Scene, ScenePerformer, ScrapeResult, StudioRef } from "@playingwithclouds/veil-sdk";
import { fetchComments } from "./comments.ts";
import { externalId, jsonItem, jsonVideoList, studioFromLink } from "./cards.ts";
import { canonicalUrl, decodeEntities, fetchText, ratingFromPercent, siteOf, type Site } from "./http.ts";

const MAX_RELATED = 40;

/** What a video page tells about the video, before comments (which need another request). */
export interface VideoPage {
  scene: Scene;
  /** The site's numeric id, which the comments endpoint takes. */
  numericId?: string;
}

/** Fetches a video page of either site and returns it as one scene, with its related videos and comments. */
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const site = siteOf(url);
  if (!site) throw new Error(`xvideos: not an xvideos or xnxx page: ${url}`);
  const target = canonicalUrl(url);
  const page = parseVideoPage(await fetchText(target), target, site);
  if (page.numericId) {
    const comments = await fetchComments(site, page.numericId, page.scene.source_url);
    if (comments.length > 0) page.scene.comments = comments;
  }
  return [{ type: "scene", scene: page.scene }];
}

/** The first-argument strings of the page's html5player.setX('…') calls, by setter name. */
export function playerSettings(html: string): Map<string, string> {
  const settings = new Map<string, string>();
  for (const match of html.matchAll(/html5player\.(set\w+)\('((?:\\.|[^'\\])*)'/g)) {
    if (settings.has(match[1])) continue;
    settings.set(match[1], match[2].replace(/\\(['"\\/])/g, "$1"));
  }
  return settings;
}

/** The page's schema.org VideoObject, or an empty object when it has none or it doesn't parse. */
function videoObject(root: HTMLElement): Record<string, any> {
  for (const script of root.querySelectorAll('script[type="application/ld+json"]')) {
    try {
      const data = JSON.parse(script.rawText);
      if (data && data["@type"] === "VideoObject") return data;
    } catch {
      continue;
    }
  }
  return {};
}

/** "PT00H05M44S" in seconds; 0 when unreadable. */
export function secondsFromIsoDuration(text: string | undefined): number {
  const match = text?.match(/^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$/);
  if (!match) return 0;
  return parseInt(match[1] || "0", 10) * 3600 + parseInt(match[2] || "0", 10) * 60 + parseInt(match[3] || "0", 10);
}

/** Parses a video page of either site into a scene; target is the page's canonical URL. */
export function parseVideoPage(html: string, target: string, site: Site): VideoPage {
  const root = parse(html);
  const player = playerSettings(html);
  const video = videoObject(root);

  const encodedId = player.get("setEncodedIdVideo") || "";
  if (encodedId === "") throw new Error(`xvideos: no video on ${target}`);
  let sourceUrl = target;
  const pagePath = player.get("setVideoURL");
  if (pagePath) sourceUrl = canonicalUrl(pagePath, site);

  const scene: Scene = {
    type: "scene",
    external_id: externalId(site, encodedId),
    source_url: sourceUrl,
    title: decodeEntities(player.get("setVideoTitle") || video.name || "").trim(),
    tags: tagsOf(root),
    performers: performersOf(root, site),
    downloads: [{ label: site, url: sourceUrl }],
    related: relatedOf(html, site, externalId(site, encodedId)),
  };
  const description = typeof video.description === "string" ? decodeEntities(video.description).trim() : "";
  if (description !== "" && description !== scene.title) scene.details = description;
  if (typeof video.uploadDate === "string") scene.date = new Date(video.uploadDate).toISOString();
  const duration = secondsFromIsoDuration(video.duration);
  if (duration > 0) scene.duration_seconds = duration;
  const views = video.interactionStatistic?.userInteractionCount;
  if (typeof views === "number") scene.view_count = views;
  const rating = ratingFromPercent(root.querySelector(".rating-good-perc, .rating-box")?.text);
  if (rating !== undefined) scene.rating = rating;
  const studio = studioOf(root, player, site);
  if (studio) scene.studio = studio;

  const poster = posterOf(video, player);
  if (poster) {
    scene.poster_path = poster;
    scene.images = [{ type: "poster", file_path: poster } as Image];
  }
  const slides = player.get("setThumbSlide");
  if (slides && slides.endsWith("/mozaique_listing.jpg")) {
    scene.preview_video = slides.replace(/mozaique_listing\.jpg$/, "preview.mp4");
  }

  const numericId = html.match(/"id_video":(\d+)/)?.[1];
  return { scene, numericId };
}

/** The poster: the VideoObject's jpg thumbnail, else the player's (xnxx's 16:9 one is avif, which not every client shows). */
function posterOf(video: Record<string, any>, player: Map<string, string>): string | undefined {
  const thumbnails = video.thumbnailUrl;
  if (Array.isArray(thumbnails) && typeof thumbnails[0] === "string") return thumbnails[0];
  if (typeof thumbnails === "string") return thumbnails;
  return player.get("setThumbUrl");
}

/** The video's tags, in page order, without repeats. */
function tagsOf(root: HTMLElement): string[] {
  const tags: string[] = [];
  for (const link of root.querySelectorAll("a.is-keyword")) {
    const tag = decodeEntities(link.text).trim();
    if (tag !== "" && !tags.includes(tag)) tags.push(tag);
  }
  return tags;
}

/**
 * The credited performers. xvideos lists them as model labels linking their
 * profile ("/pornstars/nicole_aniston"); xnxx as search links, so only the
 * name is known there.
 */
function performersOf(root: HTMLElement, site: Site): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  for (const link of root.querySelectorAll("li.model a, a.is-pornstar")) {
    const name = titleCase(decodeEntities(link.querySelector(".name")?.text || link.text).trim());
    if (name === "" || performers.some((known) => known.name === name)) continue;
    const performer: ScenePerformer = { name, order: performers.length };
    const path = link.getAttribute("href");
    if (path && /^\/(pornstars|models|profiles)\//.test(path)) {
      performer.source_url = canonicalUrl(path, site);
      performer.external_id = `${site}-${path.replace(/^\/+/, "").replace(/\//g, "-")}`;
    }
    performers.push(performer);
  }
  return performers;
}

/** xnxx writes performer names in lower case ("octavia red"); stage names are capitalized. */
function titleCase(name: string): string {
  if (name !== name.toLowerCase()) return name;
  return name.replace(/(^|[\s-])([a-z])/g, (_, lead: string, letter: string) => lead + letter.toUpperCase());
}

/** The uploader: xvideos links its channel or profile; xnxx only names the account. */
function studioOf(root: HTMLElement, player: Map<string, string>, site: Site): StudioRef | undefined {
  const uploader = root.querySelector(".main-uploader a");
  if (uploader) return studioFromLink(uploader.querySelector(".name")?.text, uploader.getAttribute("href"), site);
  return studioFromLink(player.get("setUploaderName"), undefined, site);
}

/** The related videos the page embeds as JSON (var video_related=[…]), in site order. */
function relatedOf(html: string, site: Site, ownId: string): DiscoveredItem[] {
  const match = html.match(/var video_related=(\[[\s\S]*?\]);\s*window\./);
  if (!match) return [];
  let videos: unknown;
  try {
    videos = JSON.parse(match[1]);
  } catch {
    return [];
  }
  const items: DiscoveredItem[] = [];
  for (const video of jsonVideoList(videos)) {
    const item = jsonItem(video, site);
    if (item && item.external_id !== ownId) items.push(item);
  }
  return items.slice(0, MAX_RELATED);
}
