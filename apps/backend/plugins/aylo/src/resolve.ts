import type { ResolveResult } from "@playingwithclouds/veil-sdk";
import { UA, fetchPageWithSession, streamHeaders } from "./http.ts";
import { extractFlashvars, extractMediaDefinitions } from "./page.ts";
import { requireSite, type Site } from "./sites.ts";

interface Source {
  url: string;
  format: "hls" | "mp4";
  height: number;
}

// Pornhub randomly serves some page loads stream URLs on a CDN variant (hv-h.phncdn.com)
// that answers 410 to anything but a browser; a fresh page load usually gets the working one.
const MAX_PORNHUB_ATTEMPTS = 5;

/**
 * Resolves a scene page to its best playable stream. The page lists quality
 * definitions inline (Pornhub) or as "remote" entries whose URL returns the
 * quality list as JSON (the other three); the tallest wins, HLS over MP4 on a
 * tie. The signed URLs are bound to the requesting IP, so the backend that
 * resolves must also fetch the stream.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const site = requireSite(url);
  const id = site.idFromUrl(url);
  if (!id) throw new Error(`aylo resolve: no video id in ${url}`);

  const pageUrl = site.sceneUrl(id);
  let attempts = 1;
  if (site.key === "pornhub") attempts = MAX_PORNHUB_ATTEMPTS;

  for (let attempt = 1; attempt <= attempts; attempt++) {
    const result = await resolveOnce(site, pageUrl);
    if (attempts === 1 || (await isReachable(result))) return result;
  }
  throw new Error(`aylo resolve: no reachable stream for ${pageUrl}`);
}

/** One page load and pick of the best source. */
async function resolveOnce(site: Site, pageUrl: string): Promise<ResolveResult> {
  const { html, sessionCookie } = await fetchPageWithSession(pageUrl);
  assertAvailable(site, html, pageUrl);

  const sources = await collectSources(site, pageDefinitions(site, html), pageUrl, sessionCookie);
  const best = pickBest(preferReliable(site, sources));
  if (!best) throw new Error(`aylo resolve: no playable source for ${pageUrl}`);

  let mimeType: ResolveResult["mime_type"] = "video/mp4";
  if (best.format === "hls") mimeType = "application/x-mpegURL";
  const result: ResolveResult = { url: best.url, mime_type: mimeType, headers: streamHeaders(site) };
  if (best.height > 0) result.quality = `${best.height}p`;
  return result;
}

/** True when the stream URL answers with a success status (a ranged GET keeps mp4 checks cheap). */
async function isReachable(result: ResolveResult): Promise<boolean> {
  try {
    const response = await fetch(result.url, {
      headers: { ...result.headers, Range: "bytes=0-0" },
      signal: AbortSignal.timeout(10_000),
    });
    return response.ok;
  } catch {
    return false;
  }
}

/** Throws a readable error when the site marks the video removed or blocked for this country. */
function assertAvailable(site: Site, html: string, pageUrl: string): void {
  if (site.key !== "pornhub") return;
  const flashvars = extractFlashvars(html);
  if (!flashvars) return;
  if (flashvars.video_unavailable_country === "true") {
    throw new Error(`aylo resolve: ${pageUrl} is not available in this country`);
  }
  if (flashvars.video_unavailable === "true") {
    throw new Error(`aylo resolve: ${pageUrl} is unavailable`);
  }
}

/** The raw mediaDefinitions as embedded in the page. */
export function pageDefinitions(site: Site, html: string): any[] {
  if (site.key === "pornhub") {
    const flashvars = extractFlashvars(html);
    if (flashvars && Array.isArray(flashvars.mediaDefinitions)) return flashvars.mediaDefinitions;
    return [];
  }
  return extractMediaDefinitions(html);
}

/**
 * Pornhub's inline HLS URLs are randomly served from a CDN variant that
 * answers 410 to non-browsers, while the MP4 list behind the session-bound
 * get_media call is always playable. Use MP4 there whenever it is available.
 */
function preferReliable(site: Site, sources: Source[]): Source[] {
  if (site.key !== "pornhub") return sources;
  const mp4Sources = sources.filter((source) => source.format === "mp4");
  if (mp4Sources.length > 0) return mp4Sources;
  return sources;
}

/** Flattens definitions into sources, fetching the quality list behind each remote entry. */
async function collectSources(
  site: Site,
  definitions: any[],
  pageUrl: string,
  sessionCookie: string | undefined,
): Promise<Source[]> {
  const lists = await Promise.all(
    definitions.map((definition) => sourcesOf(site, definition, pageUrl, sessionCookie)),
  );
  return lists.flat();
}

/** The sources one definition stands for: itself, or the list its remote URL returns. */
async function sourcesOf(
  site: Site,
  definition: any,
  pageUrl: string,
  sessionCookie: string | undefined,
): Promise<Source[]> {
  if (!definition || typeof definition.videoUrl !== "string" || !definition.videoUrl) return [];
  const remoteUrl = new URL(definition.videoUrl, `https://${site.host}`).toString();
  if (!definition.remote) return toSources([definition]);

  try {
    const headers: Record<string, string> = { "User-Agent": UA, Referer: pageUrl, "X-Requested-With": "XMLHttpRequest" };
    if (sessionCookie) headers.Cookie = sessionCookie;
    const response = await fetch(remoteUrl, {
      headers,
      signal: AbortSignal.timeout(15_000),
    });
    if (!response.ok) return [];
    const list = await response.json();
    if (!Array.isArray(list)) return [];
    return toSources(list);
  } catch {
    return [];
  }
}

/** Keeps the entries that are real hls/mp4 URLs and reads their height. */
function toSources(entries: any[]): Source[] {
  const sources: Source[] = [];
  for (const entry of entries) {
    if (!entry || typeof entry.videoUrl !== "string" || !entry.videoUrl.startsWith("http")) continue;
    if (entry.format !== "hls" && entry.format !== "mp4") continue;
    const height = parseInt(String(entry.quality), 10);
    sources.push({ url: entry.videoUrl, format: entry.format, height: Number.isFinite(height) ? height : 0 });
  }
  return sources;
}

/** Tallest source; at equal height HLS is preferred since it seeks and adapts better. */
export function pickBest(sources: Source[]): Source | undefined {
  let best: Source | undefined = undefined;
  for (const source of sources) {
    if (!best || source.height > best.height) {
      best = source;
    } else if (source.height === best.height && source.format === "hls" && best.format !== "hls") {
      best = source;
    }
  }
  return best;
}
