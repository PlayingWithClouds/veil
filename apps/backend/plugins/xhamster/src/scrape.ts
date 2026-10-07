import { parse, HTMLElement } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import { resolveUrl, videoIdFromUrl, decodeEntities } from "./http.ts";
import { parseVideoCards, thumbsFromInitials, type VideoThumb } from "./listing.ts";
import {
  parseInitials,
  extractPerformers,
  extractTags,
  extractStudio,
  extractDetails,
  extractDate,
} from "./initials.ts";
import type {
  ScrapeResult,
  Scene,
  Download,
  Image,
  DiscoveredItem,
} from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// Every xhamster item is a single adult video, modelled as one Scene.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const id = videoIdFromUrl(target);
  const html = await fetchHtml(target);
  const root = parse(html);

  const title = extractTitle(root);
  const poster = extractPoster(root);
  const initials = parseInitials(html);

  const scene: Scene = {
    type: "scene",
    external_id: `xhamster-${id}`,
    source_url: target,
    title,
    details: extractDetails(initials),
    date: extractDate(initials),
    studio: extractStudio(initials),
    poster_path: poster,
    images: poster ? [{ type: "poster", file_path: poster } as Image] : [],
    tags: extractTags(initials),
    performers: extractPerformers(initials),
    downloads: extractDownloads(target),
    duration_seconds: extractDurationSeconds(html),
    related: extractRelated(root, `xhamster-${id}`, thumbsFromInitials(initials)),
  };

  return [{ type: "scene", scene }];
}

// The "Related videos" tab is server-rendered with the listing card markup.
// Scoped to its block so other thumb grids on the page are not mixed in; the
// page state fills in what skeleton cards lack.
export function extractRelated(
  root: HTMLElement,
  sceneExternalId: string,
  thumbs: Map<string, VideoThumb> = new Map()
): DiscoveredItem[] {
  const block = root.querySelector('[data-block="related-video"]');
  if (!block) return [];
  return parseVideoCards(block, thumbs)
    .filter((item) => item.external_id !== sceneExternalId)
    .slice(0, MAX_RELATED);
}

function extractTitle(root: HTMLElement): string {
  const og = root.querySelector('meta[property="og:title"]')?.getAttribute("content");
  if (og) return decodeEntities(og).trim();
  const heading = root.querySelector("h1")?.text.trim();
  return heading ?? "";
}

function extractPoster(root: HTMLElement): string | undefined {
  const meta = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  return meta ? resolveUrl(meta) : undefined;
}

// The player config carries the runtime in whole seconds, scoped to this video.
function extractDurationSeconds(html: string): number {
  const match = html.match(/xplayerSettings":\{"videoId":\d+,"duration":(\d+)/);
  if (!match) return 0;
  return parseInt(match[1], 10);
}

// The playable sources on the video page are encrypted per-session, so the
// resolver re-fetches the page to extract a stream. The download therefore
// carries the video page URL as its resolvable handle.
function extractDownloads(pageUrl: string): Download[] {
  return [{ label: "xhamster", url: pageUrl }];
}
