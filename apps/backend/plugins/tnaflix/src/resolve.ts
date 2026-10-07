import { parse } from "node-html-parser";
import { BASE_URL } from "./http.ts";
import { extractSources } from "./streams.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

// Resolve yields a single playable source for immediate streaming — the highest
// resolution. Because the inlined mp4s are signed with an expiring token, the
// page is re-fetched for a fresh source. The full quality list is surfaced
// separately as the scene's downloads.
export async function resolve(url: string): Promise<ResolveResult> {
  const html = await fetchPage(url);
  const sources = extractSources(parse(html));
  if (sources.length === 0) throw new Error(`tnaflix resolver: no mp4 source at ${url}`);
  const best = sources[0];
  return { url: best.url, mime_type: "video/mp4", quality: best.quality };
}

async function fetchPage(url: string): Promise<string> {
  const target = url.startsWith("//") ? "https:" + url : url;
  const res = await fetch(target, {
    headers: {
      "User-Agent": UA,
      Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
      Referer: `${BASE_URL}/`,
    },
    signal: AbortSignal.timeout(15_000),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status} fetching ${target}`);
  return res.text();
}
