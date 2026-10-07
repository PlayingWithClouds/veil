import { BASE_URL } from "./http.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

// Preview montage clips live on thumb-*.xhcdn.com; they are not the full video.
const THUMB_HOST = /thumb-[^.]*\.xhcdn\.com/;

// xhamster serves the full video from the page's `sources`. Newer videos encrypt
// those URLs for in-browser decryption, but some still expose a plain HLS master
// or mp4. Fetch the page and return the best plain stream; throw if only the
// encrypted sources are present.
export async function resolve(url: string): Promise<ResolveResult> {
  const html = await fetchPage(url);

  const hls = pickStream(html, /https:\/\/[^\s"'\\]+?\.m3u8[^\s"'\\]*/gi);
  if (hls) {
    return { url: hls, mime_type: "application/x-mpegURL" };
  }

  const mp4 = pickStream(html, /https:\/\/[^\s"'\\]+?\.mp4[^\s"'\\]*/gi);
  if (mp4) {
    return { url: mp4, mime_type: "video/mp4" };
  }

  throw new Error(`xhamster resolver: no plain stream at ${url} (sources are encrypted)`);
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

// Return the first stream URL matching the pattern that is not a preview thumb.
function pickStream(html: string, pattern: RegExp): string | null {
  for (const match of html.matchAll(pattern)) {
    const candidate = match[0];
    if (THUMB_HOST.test(candidate)) continue;
    return candidate;
  }
  return null;
}
