import { parse, HTMLElement } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import { resolveUrl, videoIdFromUrl, decodeEntities } from "./http.ts";
import { extractSources } from "./streams.ts";
import { parseVideoCards } from "./listing.ts";
import type {
  ScrapeResult,
  Scene,
  ScenePerformer,
  StudioRef,
  Download,
  Image,
  DiscoveredItem,
} from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// Every tnaflix item is a single adult video, modelled as one Scene.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const id = videoIdFromUrl(target);
  const html = await fetchHtml(target);
  const root = parse(html);

  const title = extractTitle(root);
  const poster = extractPoster(root);
  const badges = root.querySelector("div.video-detail-badges");

  const scene: Scene = {
    type: "scene",
    external_id: `tnaflix-${id}`,
    source_url: target,
    title,
    details: extractDetails(root),
    date: extractDate(html),
    studio: extractStudio(badges),
    poster_path: poster,
    images: poster ? [{ type: "poster", file_path: poster } as Image] : [],
    tags: extractTags(badges),
    performers: extractPerformers(badges),
    downloads: extractDownloads(root, target),
    duration_seconds: extractDurationSeconds(root),
    related: extractRelated(root, `tnaflix-${id}`),
  };

  return [{ type: "scene", scene }];
}

// The "Related content" section below the player reuses the listing card grid.
export function extractRelated(root: HTMLElement, sceneExternalId: string): DiscoveredItem[] {
  return parseVideoCards(root)
    .filter((item) => item.external_id !== sceneExternalId)
    .slice(0, MAX_RELATED);
}

function extractTitle(root: HTMLElement): string {
  const og = root.querySelector('meta[property="og:title"]')?.getAttribute("content");
  if (og) return decodeEntities(og).trim();
  const heading = root.querySelector("h1")?.text.trim();
  if (heading) return heading;
  return "";
}

function extractPoster(root: HTMLElement): string | undefined {
  const meta = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  if (meta) return resolveUrl(meta);
  return undefined;
}

// The player element carries the runtime in whole seconds.
function extractDurationSeconds(root: HTMLElement): number {
  const attr = root.querySelector("video#video-player")?.getAttribute("data-duration") ?? "";
  const seconds = parseInt(attr, 10);
  if (Number.isFinite(seconds) && seconds > 0) return seconds;
  return 0;
}

// The VideoObject JSON-LD block carries the ISO publish timestamp.
function extractDate(html: string): string | undefined {
  const match = html.match(/"uploadDate":"([^"]+)"/);
  if (match) return match[1];
  return undefined;
}

// The badge strip mixes the uploader, credited performers, category chips
// (plain site-section paths) and search-term chips (/search?what=<term>). The
// entity links share badge-video-info; performers are marked badge-kiss,
// the uploader badge-verified (/channel/…) or badge-unverified (/profile/…).
export function extractPerformers(badges: HTMLElement | null): ScenePerformer[] {
  if (!badges) return [];
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();

  for (const anchor of badges.querySelectorAll('a.badge-kiss[href*="/profile/"]')) {
    const name = decodeEntities(anchor.text).trim();
    if (!name || seen.has(name)) continue;
    seen.add(name);
    performers.push({
      name,
      order: performers.length,
      source_url: resolveUrl(anchor.getAttribute("href") ?? ""),
    });
  }

  return performers;
}

/**
 * The verified channel badge as the scene's studio, keyed by its channel URL.
 * Unverified uploads come from throwaway member accounts (random handles like
 * "halvorson_patience") that would each become a junk studio, so they credit none.
 */
export function extractStudio(badges: HTMLElement | null): StudioRef | undefined {
  if (!badges) return undefined;
  const anchor = badges.querySelector("a.badge-verified");
  if (!anchor) return undefined;
  const name = decodeEntities(anchor.text).trim();
  const href = anchor.getAttribute("href");
  if (!name || !href) return undefined;
  return { name, source_url: resolveUrl(href) };
}

/**
 * Category chip names. Entity badges (uploader, performers) are skipped, and so
 * are the /search?what= chips: the site derives those from the title's words,
 * so they carry fragments like "more" or "impact" rather than real tags.
 */
export function extractTags(badges: HTMLElement | null): string[] {
  if (!badges) return [];
  const tags: string[] = [];
  const seen = new Set<string>();

  for (const anchor of badges.querySelectorAll("a")) {
    if (!isCategoryChip(anchor)) continue;
    const name = decodeEntities(anchor.text).trim();
    if (!name || seen.has(name.toLowerCase())) continue;
    seen.add(name.toLowerCase());
    tags.push(name);
  }

  return tags;
}

/** Whether a badge anchor is a category chip: a plain site-section link, not an entity or search link. */
function isCategoryChip(anchor: HTMLElement): boolean {
  if (anchor.classList.contains("badge-video-info")) return false;
  const href = anchor.getAttribute("href");
  if (!href) return false;
  if (href.includes("/search")) return false;
  return !isVideoLink(href);
}

/** Whether the link points at a video page. */
function isVideoLink(href: string): boolean {
  return videoIdFromUrl(href) !== "";
}

/**
 * The uploader's description from the detail paragraph, else the JSON-LD
 * VideoObject. Placeholders and the site's generated SEO blurbs are dropped.
 */
export function extractDetails(root: HTMLElement): string | undefined {
  let description = descriptionParagraph(root);
  if (!description) {
    description = jsonLdDescription(root);
  }
  if (!description || isPlaceholderDescription(description)) return undefined;
  return description;
}

/** The p.video-detail-description text without its "Description:" label. */
function descriptionParagraph(root: HTMLElement): string {
  const paragraph = root.querySelector("p.video-detail-description");
  if (!paragraph) return "";
  return paragraph.text.replace(/^\s*Description\s*:/i, "").trim();
}

/** The description field of the page's VideoObject JSON-LD block, if any. */
function jsonLdDescription(root: HTMLElement): string {
  for (const script of root.querySelectorAll('script[type="application/ld+json"]')) {
    const description = parseJsonLdDescription(script.rawText);
    if (description) return description;
  }
  return "";
}

/** Read `description` from one JSON-LD payload; empty when absent or malformed. */
function parseJsonLdDescription(json: string): string {
  try {
    const data = JSON.parse(json) as { description?: unknown };
    if (typeof data.description !== "string") return "";
    return decodeEntities(data.description).trim();
  } catch {
    return "";
  }
}

/**
 * Whether the text is not a real description: the "No description provided"
 * placeholder, or a generated blurb like "Watch <title> on com, the best
 * hardcore porn site…" / "Watch <title> on now! - …".
 */
function isPlaceholderDescription(text: string): boolean {
  if (/^no description provided$/i.test(text)) return true;
  return /^Watch .* on .*(the best hardcore porn site|now! - )/i.test(text);
}

// Every inlined quality is offered so the UI can present them all for download;
// each is also directly streamable. The mp4 URLs are signed with an expiring
// token, so the resolver re-fetches the page for a fresh stream when playing.
// Falls back to the page URL as a resolvable handle if no sources are inlined.
function extractDownloads(root: HTMLElement, pageUrl: string): Download[] {
  const sources = extractSources(root);
  if (sources.length === 0) return [{ label: "tnaflix", url: pageUrl }];
  return sources.map((source) => ({
    label: source.quality,
    url: source.url,
    quality: source.quality,
    format: source.format,
  }));
}
