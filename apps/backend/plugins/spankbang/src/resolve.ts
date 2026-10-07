import { fetchPage } from "./http.ts";
import { extractSources } from "./streams.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

// Resolve yields a single playable source for immediate streaming. Progressive
// mp4 is returned highest-quality first; the adaptive HLS master is used only
// when no direct mp4 is exposed. The full quality list is surfaced separately as
// the scene's downloads.
export async function resolve(url: string): Promise<ResolveResult> {
  const html = await fetchPage(url);
  const sources = extractSources(html);

  const mp4 = sources.find((source) => source.format === "mp4");
  if (mp4) {
    return { url: mp4.url, mime_type: "video/mp4", quality: mp4.quality };
  }

  const hls = sources.find((source) => source.format === "hls");
  if (hls) {
    return { url: hls.url, mime_type: "application/x-mpegURL", quality: hls.quality };
  }

  throw new Error(`spankbang resolver: no stream found at ${url}`);
}
