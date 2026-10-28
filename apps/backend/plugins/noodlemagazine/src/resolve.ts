import type { ResolveResult } from "@playingwithclouds/veil-sdk";
import { decodeEntities, fetchWatchPage, videoIdFromUrl } from "./http.ts";
import { playbackHeaders } from "./mirrors.ts";

interface PlaybackSource {
  url: string;
  height: number;
}

/**
 * Resolves a watch URL on any mirror to an mp4: the quality named in its
 * fragment ("#720p", as downloads stored by the former mat6tube plugin carry),
 * else the best. The page's player playlist lists one signed source per
 * quality; the signatures expire, so this runs at play time.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const videoId = videoIdFromUrl(url);
  if (videoId === "") throw new Error(`noodlemagazine: no video in ${url}`);
  const page = await fetchWatchPage(videoId, true);
  const source = chooseSource(page.body, requestedHeight(url));
  if (source === undefined) throw new Error(`noodlemagazine: no playable source for ${url}`);
  const result: ResolveResult = { url: source.url, mime_type: "video/mp4", headers: playbackHeaders(page.origin) };
  if (source.height > 0) result.quality = `${source.height}p`;
  return result;
}

/** The source of the requested height when the page has it, else the best one. */
export function chooseSource(html: string, height: number): PlaybackSource | undefined {
  if (height > 0) {
    const match = playlistSources(html).find((source) => source.height === height);
    if (match !== undefined) return match;
  }
  return bestSource(html);
}

/** The height a URL's fragment asks for ("#720p" → 720), 0 without one. */
export function requestedHeight(url: string): number {
  const match = url.match(/#(\d+)p$/);
  if (!match) return 0;
  return Number(match[1]);
}

/** The highest-quality source of the page's playlist, falling back to the player's video tag. */
export function bestSource(html: string): PlaybackSource | undefined {
  const sources = playlistSources(html);
  if (sources.length === 0) {
    const tagSource = videoTagSource(html);
    if (tagSource === undefined) return undefined;
    return { url: tagSource, height: 0 };
  }
  sources.sort((first, second) => second.height - first.height);
  return sources[0];
}

/** Every source of the window.playlist object, read from its "file" and "label" pairs. */
function playlistSources(html: string): PlaybackSource[] {
  const sources: PlaybackSource[] = [];
  for (const match of html.matchAll(/"file":"((?:[^"\\]|\\.)*)","label":"(\d+)"/g)) {
    const url = JSON.parse(`"${match[1]}"`) as string;
    sources.push({ url, height: Number(match[2]) });
  }
  return sources;
}

/** The src of the player's <video> element when the playlist is missing. */
function videoTagSource(html: string): string | undefined {
  const match = html.match(/<video[^>]*\ssrc="([^"]+\.mp4[^"]*)"/);
  if (!match) return undefined;
  return decodeEntities(match[1]);
}
