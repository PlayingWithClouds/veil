import { fetchHtmlSmart } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://spankbang.com";

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

// spankbang sits behind Cloudflare's managed challenge, answering plain requests
// with 403. fetchHtmlSmart transparently retries through FlareSolverr when it is
// configured, so every fetch in this plugin routes through it.
export function fetchPage(url: string): Promise<string> {
  return fetchHtmlSmart(url);
}

// Video pages look like /<code>/video/<slug>, where <code> is a short base36
// token (e.g. "7gu17"). It is the stable per-video identifier.
export function videoCodeFromUrl(url: string): string {
  const match = resolveUrl(url).match(/spankbang\.com\/([0-9a-z]+)\/video\//i);
  if (match) return match[1];
  const relative = url.match(/^\/?([0-9a-z]+)\/video\//i);
  return relative ? relative[1] : "";
}

export function isVideoUrl(url: string): boolean {
  return videoCodeFromUrl(url) !== "";
}

// Card durations render as "21m", "1h 5m" or "45s"; player length elements use
// the same shorthand. Return whole seconds.
export function parseDurationSeconds(text: string): number {
  const hours = text.match(/(\d+)\s*h/);
  const minutes = text.match(/(\d+)\s*m(?!s)/);
  const seconds = text.match(/(\d+)\s*s/);
  const total =
    (hours ? parseInt(hours[1], 10) : 0) * 3600 +
    (minutes ? parseInt(minutes[1], 10) : 0) * 60 +
    (seconds ? parseInt(seconds[1], 10) : 0);
  return total;
}

export function decodeEntities(text: string): string {
  return text
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;|&#x27;/gi, "'")
    .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(parseInt(code, 10)))
    .replace(/&#x([0-9a-f]+);/gi, (_, code) => String.fromCodePoint(parseInt(code, 16)));
}
