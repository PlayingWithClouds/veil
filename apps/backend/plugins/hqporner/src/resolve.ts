import { BASE_URL } from "./http.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

// hqporner embeds a mydaddy.cc player whose page lists mp4 sources (360/720/1080)
// as escaped <source> tags. Fetch it with an hqporner referer, then pick the
// highest-quality mp4.
export async function resolve(url: string): Promise<ResolveResult> {
  const html = await fetchEmbed(url);
  const source = pickBestSource(html);
  if (!source) throw new Error(`hqporner resolver: no mp4 source at ${url}`);
  return { url: source.url, mime_type: "video/mp4", quality: source.quality };
}

async function fetchEmbed(url: string): Promise<string> {
  const target = url.startsWith("//") ? "https:" + url : url;
  const res = await fetch(target, {
    headers: {
      "User-Agent": UA,
      Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
      Referer: `${BASE_URL}/`,
    },
    signal: AbortSignal.timeout(12_000),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status} fetching ${target}`);
  return res.text();
}

interface Source {
  url: string;
  height: number;
  quality: string;
}

// Source URLs render escaped (e.g. `\"//s63.bigcdn.cc/pubs/<id>/1080.mp4\"`), so
// match the bare protocol-relative mp4 URL and read the trailing resolution.
function pickBestSource(html: string): Source | null {
  const pattern = /\/\/[^\s"'\\]+?\/(\d{3,4})\.mp4/gi;
  const seen = new Set<string>();
  const sources: Source[] = [];

  for (const match of html.matchAll(pattern)) {
    const absolute = "https:" + match[0];
    if (seen.has(absolute)) continue;
    seen.add(absolute);
    const height = parseInt(match[1], 10);
    sources.push({ url: absolute, height, quality: `${height}p` });
  }

  if (sources.length === 0) return null;
  sources.sort((a, b) => b.height - a.height);
  return sources[0];
}
