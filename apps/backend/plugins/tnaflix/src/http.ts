export const BASE_URL = "https://www.tnaflix.com";

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

// Video pages look like /<category>/<Slug>/video25810750. The trailing numeric
// token after "video" is the stable per-video identifier.
export function videoIdFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/video(\d+)(?:[/?#]|$)/);
  return match ? match[1] : "";
}

export function isVideoUrl(url: string): boolean {
  return videoIdFromUrl(url) !== "";
}

// Listing cards show runtimes as colon-separated chips: "MM:SS" or "HH:MM:SS".
export function parseDurationSeconds(text: string): number {
  const trimmed = text.trim();
  if (!/^\d+(:\d{1,2}){1,2}$/.test(trimmed)) return 0;

  let seconds = 0;
  for (const part of trimmed.split(":")) {
    seconds = seconds * 60 + parseInt(part, 10);
  }
  return seconds;
}

// node-html-parser leaves entities untouched inside attribute values (alt,
// src, data-trailer), so decode the handful that show up.
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
