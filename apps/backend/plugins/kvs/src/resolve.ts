import { BROWSER_HEADERS, type ResolveResult } from "@playingwithclouds/veil-sdk";
import { siteForUrl, type SiteConfig } from "./sites.ts";
import { siteReferer } from "./http.ts";
import { extractSources, isMediaUrl, type MediaSource } from "./media.ts";
import { loadVideoPage } from "./scrape.ts";

const MAX_REDIRECTS = 8;
const HOP_TIMEOUT_MS = 15_000;

/** Where a /get_file/ link ended up after its redirects. */
interface FollowedMedia {
  url: string;
  mimeType: ResolveResult["mime_type"];
}

/**
 * Resolves a video page (optionally "#<quality>" from the scene's downloads)
 * or a /get_file/ link to the CDN file it redirects to. The page is re-read
 * each time because KVS signs its media links per request.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const site = siteForUrl(url);
  if (!site) throw new Error(`kvs resolver: unsupported site ${url}`);

  let source: MediaSource = { url, quality: "", height: 0 };
  if (!isMediaUrl(url)) source = await pickSource(url);

  const followed = await followMedia(site, source.url);
  const result: ResolveResult = {
    url: followed.url,
    mime_type: followed.mimeType,
    headers: { Referer: siteReferer(site), "User-Agent": BROWSER_HEADERS["User-Agent"] },
  };
  if (source.quality) result.quality = source.quality;
  return result;
}

/** The page's source of the quality named in the URL fragment, else its best one. */
async function pickSource(url: string): Promise<MediaSource> {
  const page = await loadVideoPage(url);
  const sources = extractSources(page.site, page.root, page.html);
  if (sources.length === 0) throw new Error(`kvs resolver: no playable source on ${url}`);

  const requested = requestedQuality(url);
  const match = sources.find((source) => source.quality === requested);
  if (match) return match;
  return sources[0];
}

/** The quality in a download URL's fragment ("…/#720p"), "" without one. */
export function requestedQuality(url: string): string {
  const hashIndex = url.indexOf("#");
  if (hashIndex === -1) return "";
  try {
    return decodeURIComponent(url.slice(hashIndex + 1));
  } catch {
    return "";
  }
}

/**
 * Follows a media link's redirects with HEAD requests (no body is
 * transferred) to the CDN URL that serves it, so HLS masters are returned at
 * their final address — their segment paths are relative to it. A hop the
 * server refuses ends the walk at the last URL reached.
 */
async function followMedia(site: SiteConfig, mediaUrl: string): Promise<FollowedMedia> {
  let current = mediaUrl;
  for (let hop = 0; hop < MAX_REDIRECTS; hop++) {
    const response = await headRequest(site, current);
    if (!response) break;
    const location = response.headers.get("location");
    if (response.status >= 300 && response.status < 400 && location) {
      current = new URL(location, current).toString();
      continue;
    }
    if (!response.ok) {
      if (hop === 0) throw new Error(`kvs resolver: HTTP ${response.status} for ${mediaUrl}`);
      break;
    }
    return { url: current, mimeType: mimeTypeOf(current, response.headers.get("content-type") ?? "") };
  }
  return { url: current, mimeType: mimeTypeOf(current, "") };
}

/** One HEAD request without following redirects; undefined when it fails outright. */
async function headRequest(site: SiteConfig, url: string): Promise<Response | undefined> {
  try {
    return await fetch(url, {
      method: "HEAD",
      redirect: "manual",
      headers: { "User-Agent": BROWSER_HEADERS["User-Agent"], Referer: siteReferer(site) },
      signal: AbortSignal.timeout(HOP_TIMEOUT_MS),
    });
  } catch {
    return undefined;
  }
}

/** HLS when the server says mpegurl or the URL looks like a playlist, else mp4. */
export function mimeTypeOf(url: string, contentType: string): ResolveResult["mime_type"] {
  if (/mpegurl/i.test(contentType)) return "application/x-mpegURL";
  if (/\.m3u8|\/hls\//i.test(url.split("?")[0])) return "application/x-mpegURL";
  return "video/mp4";
}
