import { HTMLElement, parse } from "node-html-parser";
import type { DiscoveredItem, Download, ScenePerformer, Scene, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { fetchText, parseIsoDuration, performerFromLink, studioFromLink } from "./http.ts";
import { parseCards } from "./listing.ts";
import { extractFlashvars, extractVideoObject } from "./page.ts";
import { requireSite, sceneExternalId, type Site } from "./sites.ts";

const MAX_RELATED = 40;

/** Scene pages carry the video itself, so one fetch yields one Scene. */
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const site = requireSite(url);
  const id = site.idFromUrl(url);
  if (!id) throw new Error(`aylo scrape: no video id in ${url}`);

  const html = await fetchText(site.sceneUrl(id));
  return [{ type: "scene", scene: buildScene(site, id, html) }];
}

/** Builds the Scene from a fetched scene page. */
export function buildScene(site: Site, id: string, html: string): Scene {
  const doc = parse(html);
  const videoObject = extractVideoObject(html);
  const flashvars = site.key === "pornhub" ? extractFlashvars(html) : undefined;

  const scene: Scene = {
    type: "scene",
    external_id: sceneExternalId(site, id),
    source_url: site.sceneUrl(id),
    title: pageTitle(doc, videoObject, flashvars),
  };

  const duration = pageDuration(videoObject, flashvars);
  if (duration > 0) scene.duration_seconds = duration;
  if (videoObject && videoObject.uploadDate) scene.date = videoObject.uploadDate;
  const poster = pagePoster(doc, videoObject, flashvars);
  if (poster) scene.poster_path = poster;

  const studio = studioFromLink(site, doc.querySelector(site.uploaderPageSelector));
  if (studio) scene.studio = studio;
  scene.performers = extractPerformers(site, doc);
  scene.tags = extractTags(site, doc);
  scene.related = extractRelated(site, doc, scene.external_id);
  scene.downloads = pageDownloads(site, scene.source_url);
  return scene;
}

/**
 * The scene's playback source: the page itself, which stream:resolve turns
 * into a stream URL on demand (the signed CDN URLs expire and are IP-bound).
 */
function pageDownloads(site: Site, pageUrl: string): Download[] {
  return [{ label: site.key, url: pageUrl }];
}

/** Title from the structured data, then the player config, then the og:title meta. */
function pageTitle(doc: HTMLElement, videoObject: any, flashvars: any): string {
  if (videoObject && videoObject.name) return videoObject.name;
  if (flashvars && flashvars.video_title) return flashvars.video_title;
  const ogTitle = doc.querySelector('meta[property="og:title"]');
  if (ogTitle) return (ogTitle.getAttribute("content") || "").trim();
  return "";
}

/** Runtime in seconds from the structured data, falling back to the player config. */
function pageDuration(videoObject: any, flashvars: any): number {
  const fromStructuredData = parseIsoDuration(videoObject ? videoObject.duration : undefined);
  if (fromStructuredData > 0) return fromStructuredData;
  if (flashvars && Number(flashvars.video_duration) > 0) return Number(flashvars.video_duration);
  return 0;
}

/** Poster URL from the structured data, the player config or og:image. */
function pagePoster(doc: HTMLElement, videoObject: any, flashvars: any): string | undefined {
  if (videoObject && videoObject.thumbnailUrl) return videoObject.thumbnailUrl;
  if (flashvars && flashvars.image_url) return flashvars.image_url;
  const ogImage = doc.querySelector('meta[property="og:image"]');
  if (ogImage) return ogImage.getAttribute("content") || undefined;
  return undefined;
}

/** Performer credits in page order, one per distinct pornstar page. */
export function extractPerformers(site: Site, doc: HTMLElement): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const anchor of doc.querySelectorAll(site.performerCreditSelector)) {
    const performer = performerFromLink(site, anchor.getAttribute("href") || "", anchor.text);
    if (!performer || seen.has(performer.source_url!)) continue;
    seen.add(performer.source_url!);
    performers.push(performer);
  }
  return performers;
}

/** Category and tag names in page order, without duplicates or site-navigation links. */
export function extractTags(site: Site, doc: HTMLElement): string[] {
  const tags: string[] = [];
  const seen = new Set<string>();
  for (const anchor of doc.querySelectorAll(site.tagLinkSelector)) {
    const href = anchor.getAttribute("href") || "";
    if (href.includes("inyourlanguage") || href.startsWith("/categories/")) continue;
    // YouPorn/Tube8 list the credited pornstars among the tag chips.
    if (href.includes("/pornstar/")) continue;
    const name = anchor.text.trim();
    const key = name.toLowerCase();
    if (name.length < 2 || seen.has(key)) continue;
    seen.add(key);
    tags.push(name);
  }
  return tags;
}

/** The related-videos grid, as listing cards so external ids match search results. */
export function extractRelated(site: Site, doc: HTMLElement, selfExternalId: string): DiscoveredItem[] {
  const related: DiscoveredItem[] = [];
  const seen = new Set<string>([selfExternalId]);
  for (const selector of site.relatedContainerSelectors) {
    const container = doc.querySelector(selector);
    if (!container) continue;
    for (const item of parseCards(site, container)) {
      if (seen.has(item.external_id)) continue;
      seen.add(item.external_id);
      related.push(item);
    }
  }
  return related.slice(0, MAX_RELATED);
}
