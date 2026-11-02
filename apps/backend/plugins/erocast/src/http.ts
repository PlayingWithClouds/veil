// The bare domain sits behind a Cloudflare challenge; www does not.
export const SITE_HOST = "www.erocast.me";
export const BASE_URL = `https://${SITE_HOST}`;

/** Tracks per listing page, on the site's search, channel, genre and tag pages. */
export const PAGE_SIZE = 20;

// Cloudflare answers 403 to the SDK's Chrome user agent when the TLS handshake
// is not Chrome's; a Firefox user agent passes from Bun and from Go.
const REQUEST_HEADERS = {
  "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0",
  "Accept-Language": "en",
  "X-Requested-With": "XMLHttpRequest",
};

const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Fetches a page or JSON reply from the site, failing on any non-2xx status. */
export async function fetchText(url: string): Promise<string> {
  const response = await fetch(url, { headers: REQUEST_HEADERS });
  if (!response.ok) throw new Error(`erocast: HTTP ${response.status} fetching ${url}`);
  return response.text();
}

/** Points a URL on the bare domain at www, which answers without a challenge. */
export function toSiteUrl(url: string): string {
  return url.replace(/^https?:\/\/erocast\.me(?=[/?#]|$)/i, BASE_URL);
}

/** The track id in a /track/<id>/<slug> URL, or an empty string when the URL is not a track page. */
export function trackIdFromUrl(url: string): string {
  const match = url.match(/\/track\/(\d+)/);
  if (!match) return "";
  return match[1];
}

/** The identity an audio carries for a track id. */
export function externalIdOf(trackId: number | string): string {
  return `erocast-${trackId}`;
}

/** Collapses whitespace runs and trims. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

/** Decodes named and numeric HTML entities in one pass, so "&amp;lt;" becomes "&lt;". */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => decodeEntity(name, match));
}

/** The character of one entity name (the text between "&" and ";"); the whole match when unknown. */
function decodeEntity(name: string, match: string): string {
  if (/^#x/i.test(name)) return String.fromCodePoint(parseInt(name.slice(2), 16));
  if (name.startsWith("#")) return String.fromCodePoint(parseInt(name.slice(1), 10));
  const decoded = NAMED_ENTITIES[name.toLowerCase()];
  if (decoded === undefined) return match;
  return decoded;
}
