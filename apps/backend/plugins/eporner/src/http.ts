export const SITE = "https://www.eporner.com";
export const API_BASE = "https://www.eporner.com/api/v2";

export const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

// eporner links come in a few shapes — /video-<id>/<slug>/, /hd-porn/<id>/<slug>/
// and /embed/<id>/ — all carrying the same alphanumeric id used across the API
// and the sources endpoint.
export function idFromUrl(url: string): string {
  const match = url.match(/\/(?:video-|hd-porn\/|embed\/)(\w+)/);
  return match ? match[1] : "";
}

export function embedUrl(id: string): string {
  return `${SITE}/embed/${id}/`;
}

// eporner's API dates render as "2026-07-03 10:16:57" (UTC). Normalize to ISO.
export function toIsoDate(added: string | undefined): string | undefined {
  if (!added) return undefined;
  const trimmed = added.trim();
  if (!trimmed) return undefined;
  return trimmed.replace(" ", "T") + "Z";
}

// API rate is a 0–5 string ("5.00"); the scene scale is 0–10.
export function toRating(rate: string | undefined): number | undefined {
  if (!rate) return undefined;
  const value = parseFloat(rate);
  // "0.00" means nobody voted yet, not a zero rating.
  if (!Number.isFinite(value) || value <= 0) return undefined;
  return Math.min(10, value * 2);
}

// The keywords field is a comma-separated tag list that also echoes the title and
// occasional empty fragments. Keep concise, sensible tags only.
export function parseKeywords(keywords: string | undefined, title: string): string[] {
  if (!keywords) return [];
  const titleLower = title.trim().toLowerCase();
  const seen = new Set<string>();
  const tags: string[] = [];
  for (const raw of keywords.split(",")) {
    const tag = raw.trim();
    const lower = tag.toLowerCase();
    if (tag.length < 2 || tag.length > 40) continue;
    if (lower === titleLower) continue;
    if (seen.has(lower)) continue;
    seen.add(lower);
    tags.push(tag);
  }
  return tags;
}
