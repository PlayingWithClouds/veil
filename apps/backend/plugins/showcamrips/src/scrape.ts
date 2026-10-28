import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, Download, Scene, ScenePerformer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import {
  externalId,
  fetchPage,
  isVideoUrl,
  listingPageFromUrl,
  parseClockSeconds,
  parseSiteDate,
  playUrl,
  posterUrl,
  recordingStampFromUrl,
  resolveUrl,
  videoIdFromUrl,
  type ListingKind,
} from "./http.ts";
import { parseVideoCards, sceneTitle } from "./listing.ts";

const MAX_RELATED = 40;

/** The tag every recorded show gets, matching what Veil files its own recordings under. */
const LIVE_CAM_TAG = "Live cam";

/** Name and URL of the model, cam site and category a video page links. */
interface PageCredits {
  model?: { name: string; url: string };
  site?: string;
  category?: string;
}

/** Scrapes a video page into one Scene; other URLs are not scraped. */
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  if (!isVideoUrl(target)) throw new Error(`showcamrips: unsupported URL ${target}`);
  const html = await fetchPage(target);
  if (!html) throw new Error(`showcamrips: empty page ${target}`);
  const root = parse(html);
  return [{ type: "scene", scene: buildScene(root, target) }];
}

/** The Scene a video page describes. */
export function buildScene(root: HTMLElement, pageUrl: string): Scene {
  const videoId = videoIdFromUrl(pageUrl);
  const credits = readCredits(root);

  const scene: Scene = {
    type: "scene",
    external_id: externalId(videoId),
    source_url: pageUrl,
    title: pageTitle(root, credits, pageUrl),
    poster_path: posterUrl(videoId),
    images: [{ type: "poster", file_path: posterUrl(videoId) }],
    tags: sceneTags(credits),
    performers: scenePerformers(credits),
    downloads: sceneDownloads(videoId),
    related: relatedItems(root, externalId(videoId)),
  };
  if (credits.site) scene.studio = { name: credits.site };
  const duration = pageDuration(root);
  if (duration > 0) scene.duration_seconds = duration;
  const date = parseSiteDate(recordingStampFromUrl(pageUrl) || pageHeadingText(root, "Date"));
  if (date) scene.date = date;
  return scene;
}

/**
 * The model, cam site and category of the video. The links under the
 * player carry `class="color_2"` and are the only ones of their kind
 * outside the sidebar menu and the card lists.
 */
function readCredits(root: HTMLElement): PageCredits {
  const credits: PageCredits = {};
  for (const anchor of root.querySelectorAll("a.color_2")) {
    const kind = linkKind(anchor.getAttribute("href") || "");
    const name = anchor.text.trim();
    if (!kind || !name) continue;
    if (kind === "model" && !credits.model) {
      credits.model = { name, url: resolveUrl(anchor.getAttribute("href") || "") };
    }
    if (kind === "site" && !credits.site) credits.site = name;
    if (kind === "cat" && !credits.category) credits.category = name;
  }
  return credits;
}

/** Which listing a link leads to, or undefined for any other link. */
function linkKind(href: string): ListingKind | undefined {
  const page = listingPageFromUrl(href);
  if (!page) return undefined;
  return page.kind;
}

/** The listing-style title, else the page heading without its "CamRips" tail. */
function pageTitle(root: HTMLElement, credits: PageCredits, pageUrl: string): string {
  let modelName = "";
  if (credits.model) modelName = credits.model.name;
  let siteName = "";
  if (credits.site) siteName = credits.site;
  const title = sceneTitle(modelName, siteName, pageUrl);
  if (title) return title;
  const heading = root.querySelector("h1")?.text || "";
  return heading.replace(/\s*CamRips\s*$/i, "").trim();
}

/** The cam show's category (Female, Couple, Trans) plus the live-cam tag. */
function sceneTags(credits: PageCredits): string[] {
  const tags = [LIVE_CAM_TAG];
  if (credits.category) tags.push(credits.category);
  return tags;
}

/** The model as the scene's performer, identified by her page. */
function scenePerformers(credits: PageCredits): ScenePerformer[] {
  if (!credits.model) return [];
  const page = listingPageFromUrl(credits.model.url);
  const performer: ScenePerformer = { name: credits.model.name, order: 0, source_url: credits.model.url };
  if (page) performer.external_id = `showcamrips-model-${page.name}`;
  return [performer];
}

/**
 * The one mp4 behind the player. Its URL is only known after the play page
 * is fetched, so the download points at that page and `resolve` follows it.
 */
function sceneDownloads(videoId: string): Download[] {
  return [{ label: "mp4", url: playUrl(videoId), format: "mp4" }];
}

/** The runtime from the "Duration : 00:13:13" line under the player. */
function pageDuration(root: HTMLElement): number {
  return parseClockSeconds(pageHeadingText(root, "Duration"));
}

/** The value of the `<h3>Label : value</h3>` line under the player, or "". */
function pageHeadingText(root: HTMLElement, label: string): string {
  for (const heading of root.querySelectorAll("span.tl h3")) {
    const text = heading.text.trim();
    if (!text.startsWith(label)) continue;
    return text.slice(label.length).replace(/^\s*:\s*/, "");
  }
  return "";
}

/**
 * The page's related videos in site order: the "Related" list first, then
 * the first list of the model's other shows ("More by"), without the video
 * itself.
 */
export function relatedItems(root: HTMLElement, selfExternalId: string): DiscoveredItem[] {
  const related = root.querySelector("#myListmoi1");
  const lists: HTMLElement[] = [];
  if (related) lists.push(related);
  const moreByModel = root.querySelectorAll("ul.content").find((list) => list.id !== "myListmoi1");
  if (moreByModel) lists.push(moreByModel);

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>([selfExternalId]);
  for (const list of lists) {
    for (const item of parseVideoCards(list)) {
      if (seen.has(item.external_id)) continue;
      seen.add(item.external_id);
      items.push(item);
    }
  }
  return items.slice(0, MAX_RELATED);
}
