import { apiVideo, toScene } from "./api.ts";
import { idFromUrl } from "./http.ts";
import { fetchDoc, absoluteUrl } from "./html.ts";
import { parseVideoCards } from "./listing.ts";
import { HTMLElement } from "node-html-parser";
import type { ScrapeResult, ScenePerformer, DiscoveredItem, StudioRef } from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// Every eporner item is a single video → one Scene, built from the API detail
// and enriched from the video page with the pornstar credits, the channel or
// uploader, the site's categories/tags and the related videos list (the API has
// none of these).
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const id = idFromUrl(url);
  if (!id) throw new Error(`eporner scrape: no video id in ${url}`);

  const video = await apiVideo(id);
  if (!video) throw new Error(`eporner scrape: no video for id ${id}`);

  const scene = toScene(video);
  const doc = await fetchDoc(video.url).catch(() => null);
  if (!doc) return [{ type: "scene", scene }];

  const performers = extractPerformerCredits(doc);
  if (performers.length > 0) scene.performers = performers;
  const studio = extractStudio(doc);
  if (studio) scene.studio = studio;
  scene.tags = mergeTags(scene.tags ?? [], extractPageTags(doc));
  scene.related = extractRelated(doc, scene.external_id);

  return [{ type: "scene", scene }];
}

/**
 * Reads who published the video: a real channel (`li.vit-channel`, the same
 * /channel/ page studio:find scrapes) wins over the uploader's /profile/ page.
 */
export function extractStudio(doc: HTMLElement): StudioRef | undefined {
  const channel = studioFromLink(doc.querySelector('#video-info-tags li.vit-channel a[href*="/channel/"]'));
  if (channel) {
    // Same external_id scheme as studios.ts so ingest merges it with studio:find.
    const slug = channel.source_url!.match(/\/channel\/([^/]+)/);
    if (slug) channel.external_id = `eporner-ch-${slug[1]}`;
    return channel;
  }
  return studioFromLink(doc.querySelector('#video-info-tags li.vit-uploader a[href*="/profile/"]'));
}

/** Builds a StudioRef from a channel/profile link, its page URL as identity. */
function studioFromLink(anchor: HTMLElement | null): StudioRef | undefined {
  if (!anchor) return undefined;
  const href = anchor.getAttribute("href");
  const name = anchor.text.trim();
  if (!href || !name) return undefined;
  return { name, source_url: absoluteUrl(href) };
}

/** Reads the video page's category (/cat/) and tag (/tag/) link texts, in page order. */
export function extractPageTags(doc: HTMLElement): string[] {
  const selector = '#video-info-tags li.vit-category a[href*="/cat/"], #video-info-tags li.vit-tag a[href*="/tag/"]';
  return doc
    .querySelectorAll(selector)
    .map((anchor) => anchor.text.trim())
    .filter((name) => name.length > 0);
}

/** Appends `extra` to `base`, dropping names already present case-insensitively. */
export function mergeTags(base: string[], extra: string[]): string[] {
  const merged: string[] = [];
  const seen = new Set<string>();
  for (const tag of [...base, ...extra]) {
    const key = tag.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    merged.push(tag);
  }
  return merged;
}

// The "Videos related to …" block under the player reuses the site's video grid.
export function extractRelated(doc: HTMLElement, selfExternalId: string): DiscoveredItem[] {
  const block = doc.querySelector("#relateddiv");
  if (!block) return [];
  return parseVideoCards(block)
    .filter((item) => item.external_id !== selfExternalId)
    .slice(0, MAX_RELATED);
}

// Reads the video page's `vit-pornstar` credits into ScenePerformer refs. Each
// carries a source_url so ingest matches or creates the performer record.
export function extractPerformerCredits(doc: HTMLElement): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const item of doc.querySelectorAll("#video-info-tags li.vit-pornstar a")) {
    const href = item.getAttribute("href");
    const name = item.text.trim();
    if (!href || !href.includes("/pornstar/") || !name) continue;
    const sourceUrl = absoluteUrl(href);
    if (seen.has(sourceUrl)) continue;
    seen.add(sourceUrl);
    performers.push({ name, source_url: sourceUrl });
  }
  return performers;
}
