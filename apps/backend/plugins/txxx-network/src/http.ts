import { fetchHtml } from "@playingwithclouds/veil-sdk";

// All TxxxNetwork sites share one backend and JSON API. discover/search use the
// host configured via the TXXX_HOST setting; scrape/resolve derive the host from
// the video URL so they work across every network domain.
export const DEFAULT_HOST = "txxx.com";

// Known TxxxNetwork domains. Used for resolve routing (meta.domains).
export const NETWORK_DOMAINS = [
  "txxx.com",
  "upornia.com",
  "hclips.com",
  "hdzog.com",
  "vjav.com",
  "voyeurhit.com",
  "thegay.com",
  "shemalez.com",
  "fetishshrine.com",
  "hotmovs.com",
];

export function configuredHost(): string {
  const host = process.env.TXXX_HOST;
  if (host && host.trim()) return host.trim();
  return DEFAULT_HOST;
}

export function hostFromUrl(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return configuredHost();
  }
}

export function baseUrl(host: string): string {
  return `https://${host}`;
}

// Video pages look like /videos/<id>/<slug>/; resolved stream handles look like
// /get_file/<n>/<hash>/<bucket>/<id>/<id>_<tag>.mp4/. Both carry the numeric id.
export function videoIdFromUrl(url: string): string {
  const page = url.match(/\/videos?\/(\d+)/);
  if (page) return page[1];
  const getFile = url.match(/\/get_file\/[^?]*?\/(\d+)\/\d+_[a-z]+\.mp4/i);
  return getFile ? getFile[1] : "";
}

// A resolved stream handle encodes the format tag (e.g. "_hq") in its filename.
export function formatTagFromUrl(url: string): string {
  const match = url.match(/\/\d+(_[a-z]+)\.mp4/i);
  return match ? match[1] : "";
}

export function isVideoUrl(url: string): boolean {
  return videoIdFromUrl(url) !== "";
}

// The site scopes video records into id buckets: /<1e6 floor>/<1e3 floor>/<id>.
export function videoApiPath(id: string, lifetime = 8640000): string {
  const numeric = parseInt(id, 10);
  const millionBucket = Math.floor(numeric / 1000000) * 1000000;
  const thousandBucket = Math.floor(numeric / 1000) * 1000;
  return `/api/json/video/${lifetime}/${millionBucket}/${thousandBucket}/${id}.json`;
}

// The scene page's "related videos" block, same id bucketing as videoApiPath.
// Returns a videos2-shaped payload ({ videos: [...] }).
export function relatedApiPath(id: string, count: number, lifetime = 432000): string {
  const numeric = parseInt(id, 10);
  const millionBucket = Math.floor(numeric / 1000000) * 1000000;
  const thousandBucket = Math.floor(numeric / 1000) * 1000;
  return `/api/json/videos_related2/${lifetime}/${count}/${millionBucket}/${thousandBucket}/${id}.all.1.json`;
}

// The API only answers XHR-style requests carrying a same-origin referer.
export async function fetchJson<T>(host: string, path: string): Promise<T> {
  const url = baseUrl(host) + path;
  const text = await fetchHtml(url, {
    "X-Requested-With": "XMLHttpRequest",
    Referer: baseUrl(host) + "/",
  });
  return JSON.parse(text) as T;
}

// The playable stream URL returned by videofile.php is obfuscated with the
// site's custom "base164" codec: a base64 variant whose alphabet swaps five
// Latin letters (A B C E M) for their Cyrillic homoglyphs and uses "~" as the
// padding character.
const BASE164_ALPHABET =
  "АВСDЕFGHIJKLМNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789.,~";

export function base164Decode(encoded: string): string {
  const cleaned = encoded.replace(/[^АВСЕМA-Za-z0-9.,~]/g, "");
  const codes: number[] = [];
  let index = 0;

  while (index < cleaned.length) {
    const first = BASE164_ALPHABET.indexOf(cleaned.charAt(index++));
    const second = charIndexOrPad(cleaned, index++);
    const third = charIndexOrPad(cleaned, index++);
    const fourth = charIndexOrPad(cleaned, index++);

    codes.push((first << 2) | (second >> 4));
    if (third !== 64) codes.push(((15 & second) << 4) | (third >> 2));
    if (fourth !== 64) codes.push(((3 & third) << 6) | fourth);
  }

  return String.fromCharCode(...codes);
}

function charIndexOrPad(text: string, index: number): number {
  if (index >= text.length) return 64;
  return BASE164_ALPHABET.indexOf(text.charAt(index));
}
