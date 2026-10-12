import { BROWSER_HEADERS, fetchHtml, type ResolveResult } from "@playingwithclouds/veil-sdk";
import { playlistSources, type PlaylistSource } from "./scene.ts";
import { BASE_URL, watchIdFromHref, watchUrl } from "./site.ts";

/**
 * Resolves a video page (optionally "#<quality>" from the scene's downloads)
 * to the mp4 of that quality. The page is re-read because its links expire.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const id = watchIdFromHref(url);
  if (id === undefined) throw new Error(`mat6tube resolver: not a video page: ${url}`);
  const sources = playlistSources(await fetchHtml(watchUrl(id)));
  if (sources.length === 0) throw new Error(`mat6tube resolver: no playable source on ${url}`);
  const source = pickSource(sources, requestedQuality(url));
  return {
    url: source.file,
    mime_type: "video/mp4",
    quality: `${source.label}p`,
    headers: { Referer: BASE_URL + "/", "User-Agent": BROWSER_HEADERS["User-Agent"] },
  };
}

/** The source of quality `quality` ("720p"), else the best source. */
function pickSource(sources: PlaylistSource[], quality: string): PlaylistSource {
  const match = sources.find((source) => `${source.label}p` === quality);
  if (match) return match;
  return sources[0];
}

/** The quality in a download URL's fragment ("…/watch/x#720p"), "" without one. */
function requestedQuality(url: string): string {
  const hashIndex = url.indexOf("#");
  if (hashIndex === -1) return "";
  return decodeURIComponent(url.slice(hashIndex + 1));
}
