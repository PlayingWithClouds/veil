export const BASE_URL = "https://xhamster.com";

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

// Video pages look like /videos/<slug>-<id>. The id is the trailing hyphen
// segment: a short alphanumeric token (e.g. "xhoyFVR") on new videos, or a bare
// number on older ones. Return it as the stable per-video identifier.
export function videoIdFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/videos\/([^/?#]+)/);
  if (!match) return "";
  const parts = match[1].split("-");
  return parts[parts.length - 1];
}

export function isVideoUrl(url: string): boolean {
  return /\/videos\/[^/?#]+/.test(resolveUrl(url));
}

export function providerFromUrl(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "").split(".")[0];
  } catch {
    return "unknown";
  }
}

// xplayerSettings reports the runtime in whole seconds. Convert to whole minutes
// to match the Movie.runtime shape.
export function minutesFromSeconds(seconds: number): number {
  if (!Number.isFinite(seconds) || seconds <= 0) return 0;
  return Math.round(seconds / 60);
}

// Listing cards show the runtime as a colon-separated badge: "MM:SS" or
// "H:MM:SS". Returns 0 when the text is not such a label.
export function durationSecondsFromLabel(text: string): number {
  const match = text.trim().match(/^(?:(\d+):)?(\d{1,2}):(\d{2})$/);
  if (!match) return 0;
  const hours = match[1] ? parseInt(match[1], 10) : 0;
  return hours * 3600 + parseInt(match[2], 10) * 60 + parseInt(match[3], 10);
}

// node-html-parser leaves entities untouched inside attribute values (aria-label,
// alt), so card titles arrive as raw HTML. Decode the handful that show up.
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
