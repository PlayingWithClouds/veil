import { hostFromUrl, videoIdFromUrl, formatTagFromUrl } from "./http.ts";
import { fetchSources } from "./streams.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

// Resolve returns a freshly signed, playable source. It accepts both a video
// page URL and a previously stored /get_file/ stream handle (whose signature may
// have expired) — in the latter case it re-derives the same quality so the
// chosen download keeps working. The full quality list lives on the scene's
// downloads.
export async function resolve(url: string): Promise<ResolveResult> {
  const host = hostFromUrl(url);
  const id = videoIdFromUrl(url);
  if (!id) throw new Error(`txxx-network resolver: no video id in ${url}`);

  const sources = await fetchSources(host, id, url);
  if (sources.length === 0) {
    throw new Error(`txxx-network resolver: no playable format for ${url}`);
  }

  // Preserve the requested quality when resolving a stored stream handle.
  const requestedTag = formatTagFromUrl(url);
  const match = requestedTag && sources.find((source) => source.tag === requestedTag);
  const chosen = match || sources[0];

  return { url: chosen.url, mime_type: "video/mp4", quality: chosen.quality };
}
