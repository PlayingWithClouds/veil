import type { ResolveResult } from "@playingwithclouds/veil-sdk";
import { canonicalUrl, fetchText, siteOf } from "./http.ts";
import { playerSettings } from "./scrape.ts";

/**
 * Turns a video page of either site into a playable stream. The page carries
 * signed stream addresses that expire after some hours, so it is fetched anew
 * each time: the HLS master (all qualities up to 1080p) when present, else the
 * better progressive mp4.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  if (!siteOf(url)) throw new Error(`xvideos: not an xvideos or xnxx page: ${url}`);
  const player = playerSettings(await fetchText(canonicalUrl(url)));
  const hls = player.get("setVideoHLS");
  if (hls) return { url: hls, mime_type: "application/x-mpegURL" };
  const mp4 = player.get("setVideoUrlHigh") || player.get("setVideoUrlLow");
  if (mp4) return { url: mp4, mime_type: "video/mp4" };
  throw new Error(`xvideos: no stream on ${url}`);
}
