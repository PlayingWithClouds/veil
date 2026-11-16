import { PLAYBACK_HEADERS, playbackIdFromUrl, playUrl } from "./http.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

const FETCH_TIMEOUT_MS = 15_000;

/**
 * Resolves a play page (`play.php?idd=<id>`), the player iframe or a video
 * page to the video's mp4. The mp4 sits on a separate host that only serves
 * requests carrying a showcamrips referer, so the result names the headers
 * to play it with.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const videoId = playbackIdFromUrl(url);
  if (!videoId) throw new Error(`showcamrips resolver: no video id in ${url}`);
  const html = await fetchPlayPage(videoId);
  const source = videoSource(html);
  if (!source) throw new Error(`showcamrips resolver: no video on the play page of ${videoId}`);
  return { url: source, mime_type: "video/mp4", headers: { ...PLAYBACK_HEADERS } };
}

/** The play page's HTML, fetched with the referer the site checks. */
async function fetchPlayPage(videoId: string): Promise<string> {
  const target = playUrl(videoId);
  const response = await fetch(target, {
    headers: { ...PLAYBACK_HEADERS, Accept: "text/html,application/xhtml+xml" },
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${target}`);
  return response.text();
}

const VIDEO_TAG_SOURCE = /<video[^>]*\ssrc="(https?:\/\/[^"]+\.mp4[^"]*)"/i;
const SCRIPT_MP4_SOURCE = /\.src\s*=\s*['"](https?:\/\/[^'"]+\.mp4(?:\?[^'"]*)?)['"]/i;

/**
 * The mp4 URL of the page's player. The `<video>` tag ships with an empty
 * `src`; an inline script assigns it, with a decoy `.mp6` URL in the branch
 * for an 800x600 Chrome and the real `.mp4` in the other, so only an mp4
 * assignment counts. A `<video src>` filled in the markup is still honoured.
 * The ad-injection and "loading_video" variants carry neither, so they
 * resolve to nothing.
 */
export function videoSource(html: string): string {
  const tagMatch = html.match(VIDEO_TAG_SOURCE);
  if (tagMatch) return tagMatch[1];
  const scriptMatch = html.match(SCRIPT_MP4_SOURCE);
  if (scriptMatch) return scriptMatch[1];
  return "";
}
