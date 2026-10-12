import { fetchHtmlSmart, resolveUrl as resolveAgainst } from "@playingwithclouds/veil-sdk";
import type { HTMLElement } from "node-html-parser";

export const BASE_URL = "https://noodlemagazine.com";

// The site answers plain requests with a Cloudflare challenge (403), so every
// page is fetched through fetchHtmlSmart, which retries via FlareSolverr.
export function fetchPage(url: string): Promise<string> {
  return fetchHtmlSmart(url);
}

/** Resolves a page or asset path against the site. */
export function resolveUrl(path: string): string {
  return resolveAgainst(BASE_URL, path);
}

/** The watch page of a video id such as -152603790_456239053. */
export function watchUrl(videoId: string): string {
  return `${BASE_URL}/watch/${videoId}`;
}

/** The video id in a watch URL, or an empty string when the URL has none. */
export function videoIdFromUrl(url: string): string {
  const match = url.match(/\/watch\/(-?\d+_\d+)/);
  if (!match) return "";
  return match[1];
}

/** An element's attribute value, or an empty string when the attribute is missing. */
export function attributeOf(element: HTMLElement, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}

/** Collapses whitespace runs and trims. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

const NAMED_ENTITIES: Record<string, string> = {
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Decodes the HTML entities a page writes into attribute values and text. */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => {
    if (name.startsWith("#x") || name.startsWith("#X")) return String.fromCodePoint(parseInt(name.slice(2), 16));
    if (name.startsWith("#")) return String.fromCodePoint(parseInt(name.slice(1), 10));
    const decoded = NAMED_ENTITIES[name.toLowerCase()];
    if (decoded === undefined) return match;
    return decoded;
  }).replace(/&amp;/g, "&");
}
