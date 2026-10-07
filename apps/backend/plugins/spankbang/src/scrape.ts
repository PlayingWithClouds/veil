import { parse, HTMLElement } from "node-html-parser";
import { fetchPage, resolveUrl, videoCodeFromUrl, decodeEntities } from "./http.ts";
import { extractSources } from "./streams.ts";
import { parseCardsIn } from "./listing.ts";
import type {
  ScrapeResult,
  Scene,
  Download,
  Image,
  DiscoveredItem,
  ScenePerformer,
  StudioRef,
} from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// The video page shows related videos under the player (with a "show more"
// half that is already in the markup) and in the right-hand sidebar.
const RELATED_CONTAINERS = [".js-related-videos-bottom", ".js-related-videos-right"];

// Every spankbang item is a single adult video, modelled as one Scene.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const code = videoCodeFromUrl(target);
  const html = await fetchPage(target);
  const root = parse(html);

  const poster = extractPoster(root);

  const scene: Scene = {
    type: "scene",
    external_id: `spankbang-${code}`,
    source_url: target,
    title: extractTitle(root),
    date: extractDate(root),
    duration_seconds: extractDurationSeconds(root),
    poster_path: poster,
    images: poster ? [{ type: "poster", file_path: poster } as Image] : [],
    tags: extractTags(root),
    performers: extractPerformers(root),
    downloads: extractDownloads(html),
    related: extractRelated(root, `spankbang-${code}`),
  };
  const studio = extractStudio(root);
  if (studio) scene.studio = studio;

  return [{ type: "scene", scene }];
}

/** Reads the pornstar credits (`/<id>/pornstar/<name>/` links in the video-tags block), in page order. */
export function extractPerformers(root: HTMLElement): ScenePerformer[] {
  const block = root.querySelector('[data-testid="video-tags"]');
  if (!block) return [];

  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const anchor of block.querySelectorAll('a[href*="/pornstar/"]')) {
    const performer = creditFromLink(anchor);
    if (!performer || seen.has(performer.source_url)) continue;
    seen.add(performer.source_url);
    performers.push(performer);
  }
  return performers;
}

/**
 * Reads who published the video from the video-tags block: the studio channel
 * (`/channel/`) when credited, else the uploading creator's profile (`/creator/`).
 */
export function extractStudio(root: HTMLElement): StudioRef | undefined {
  const block = root.querySelector('[data-testid="video-tags"]');
  if (!block) return undefined;

  const channel = creditFromLink(block.querySelector('a[href*="/channel/"]'));
  if (channel) return channel;
  return creditFromLink(block.querySelector('a[href*="/creator/"]'));
}

/** Turns a credit link into a name plus its absolute page URL, the credit's identity. */
function creditFromLink(anchor: HTMLElement | null): { name: string; source_url: string } | undefined {
  if (!anchor) return undefined;
  const href = anchor.getAttribute("href");
  const name = decodeEntities(anchor.text).trim();
  if (!href || !name) return undefined;
  return { name, source_url: resolveUrl(href) };
}

export function extractRelated(root: HTMLElement, selfExternalId: string): DiscoveredItem[] {
  const related: DiscoveredItem[] = [];
  const seen = new Set<string>([selfExternalId]);

  for (const selector of RELATED_CONTAINERS) {
    const container = root.querySelector(selector);
    if (!container) continue;
    for (const item of parseCardsIn(container)) {
      if (seen.has(item.external_id)) continue;
      seen.add(item.external_id);
      related.push(item);
    }
  }

  return related.slice(0, MAX_RELATED);
}

// Every quality is offered so the UI can present them all for download; each is
// also directly streamable. The adaptive HLS master is included as "auto".
function extractDownloads(html: string): Download[] {
  return extractSources(html).map((source) => ({
    label: source.quality === "auto" ? "Auto (HLS)" : source.quality,
    url: source.url,
    quality: source.quality,
    format: source.format,
  }));
}

function extractTitle(root: HTMLElement): string {
  const heading = root.querySelector("h1")?.text.trim();
  if (heading) return decodeEntities(heading);
  const og = root.querySelector('meta[property="og:title"]')?.getAttribute("content") ?? "";
  return decodeEntities(og.replace(/\s*-\s*SpankBang\s*$/i, "")).trim();
}

function extractPoster(root: HTMLElement): string | undefined {
  const meta = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  return meta ? resolveUrl(meta) : undefined;
}

// The player exposes the runtime in whole seconds via og:video:duration.
function extractDurationSeconds(root: HTMLElement): number {
  const meta = root.querySelector('meta[property="og:video:duration"]')?.getAttribute("content");
  const seconds = parseInt(meta ?? "", 10);
  return Number.isFinite(seconds) && seconds > 0 ? seconds : 0;
}

// The upload timestamp sits on the first <time datetime="…"> in the page.
function extractDate(root: HTMLElement): string | undefined {
  const iso = root.querySelector("time[datetime]")?.getAttribute("datetime");
  if (!iso) return undefined;
  const match = iso.match(/^(\d{4}-\d{2}-\d{2})/);
  return match ? match[1] : undefined;
}

// Tags live in the video-tags block as /s/<tag>/ links; fall back to the
// keywords meta when the block is absent.
function extractTags(root: HTMLElement): string[] {
  const block = root.querySelector('[data-testid="video-tags"]');
  const names: string[] = [];
  const seen = new Set<string>();

  if (block) {
    for (const anchor of block.querySelectorAll('a[href^="/s/"]')) {
      pushName(decodeEntities(anchor.text), names, seen);
    }
  }

  if (names.length === 0) {
    const keywords = root.querySelector('meta[name="keywords"]')?.getAttribute("content") ?? "";
    for (const part of keywords.split(",")) pushName(part, names, seen);
  }

  return names;
}

function pushName(raw: string, names: string[], seen: Set<string>): void {
  const name = raw.trim();
  if (!name || seen.has(name.toLowerCase())) return;
  seen.add(name.toLowerCase());
  names.push(name);
}
