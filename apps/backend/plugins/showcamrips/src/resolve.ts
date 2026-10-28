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

/**
 * The mp4 URL of the page's `<video src>`. The ad-injection and the
 * "loading_video" variants of the page carry no video, so a page without
 * one resolves to nothing.
 */
export function videoSource(html: string): string {
  const match = html.match(/<video[^>]*\ssrc="(https?:\/\/[^"]+\.mp4[^"]*)"/i);
  if (!match) return "";
  return match[1];
}
