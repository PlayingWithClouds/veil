import { fetchHtmlSmart } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://missav.ws";

// missav sits behind Cloudflare, answering plain requests (listings, search and
// video pages alike) with 403. fetchHtmlSmart transparently retries through
// FlareSolverr when it is configured, so every fetch in this plugin routes
// through it.
export function fetchPage(url: string): Promise<string> {
  return fetchHtmlSmart(url);
}

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

/**
 * Canonical absolute form of a missav link with any /dm<n> mirror prefix
 * removed, so the same actress/maker page keeps one identity across mirrors.
 */
export function canonicalUrl(href: string): string {
  return resolveUrl(href).replace(/\/dm\d+\//, "/");
}

// Video pages look like /en/<code> (optionally behind a /dm<n> mirror prefix,
// e.g. /dm539/en/<code>). The trailing path segment is the DVD-style code such
// as "ebon-006" and serves as the stable per-video identifier.
export function videoCodeFromUrl(url: string): string {
  try {
    const segments = new URL(resolveUrl(url)).pathname.split("/").filter(Boolean);
    return segments[segments.length - 1] ?? "";
  } catch {
    return "";
  }
}

// A video detail URL is /<lang>/<code> where the code carries a hyphen+digit
// (e.g. ebon-006). Section pages like /en/genres or /en/actresses/<name> do not
// match, so this cleanly separates videos from navigation links.
export function isVideoUrl(url: string): boolean {
  const code = videoCodeFromUrl(url);
  return /-\d/.test(code);
}

export function providerFromUrl(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "").split(".")[0];
  } catch {
    return "unknown";
  }
}

// Listing cards overlay the runtime as "H:MM:SS" or "MM:SS". Returns 0 when the
// text is not a clock duration.
export function parseClockDuration(text: string): number {
  const match = text.trim().match(/^(?:(\d+):)?(\d{1,2}):(\d{2})$/);
  if (!match) return 0;
  const hours = match[1] ? parseInt(match[1], 10) : 0;
  const minutes = parseInt(match[2], 10);
  const seconds = parseInt(match[3], 10);
  return hours * 3600 + minutes * 60 + seconds;
}

// Release dates render as an ISO "YYYY-MM-DD". Extract the leading year.
export function yearFromDate(date: string): number | undefined {
  const match = date.match(/^(\d{4})/);
  return match ? parseInt(match[1], 10) : undefined;
}

// node-html-parser leaves entities untouched inside attribute values, so card
// titles arrive as raw HTML. Decode the handful that show up.
export function decodeEntities(text: string): string {
  return text
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;|&#x27;/gi, "'")
    .replace(/&#x3D;/gi, "=")
    .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(parseInt(code, 10)))
    .replace(/&#x([0-9a-f]+);/gi, (_, code) => String.fromCodePoint(parseInt(code, 16)));
}
