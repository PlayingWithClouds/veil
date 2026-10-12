import type { ResolveResult } from "@playingwithclouds/veil-sdk";
import { decodeEntities, fetchPage, videoIdFromUrl, watchUrl } from "./http.ts";

interface PlaybackSource {
  url: string;
  height: number;
}

/**
 * Resolves a watch URL to its best mp4. The page's player playlist lists one
 * signed source per quality; the signatures expire, so this runs at play time.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const videoId = videoIdFromUrl(url);
  if (videoId === "") throw new Error(`ukdevilz: no video in ${url}`);
  const source = bestSource(await fetchPage(watchUrl(videoId)));
  if (source === undefined) throw new Error(`ukdevilz: no playable source for ${url}`);
  return { url: source.url, mime_type: "video/mp4", quality: `${source.height}p` };
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
