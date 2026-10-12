import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { HTMLElement } from "node-html-parser";

export const BASE_URL = "https://www.freeones.com";

// The site localises by Accept-Language and by a path prefix (/de/...). English
// labels are what the bio parser reads, so every request asks for English.
const ENGLISH_HEADERS = { "Accept-Language": "en-US,en;q=0.9" };

/** Fetches a freeones page in English. */
export function fetchPage(url: string): Promise<string> {
  return fetchHtml(url, ENGLISH_HEADERS);
}

// Locale prefixes the site puts in front of a path (/de/sara-jay/bio).
const LOCALE_PREFIXES = new Set(["de", "en", "fr", "es", "it", "nl", "pt"]);

/** The performer slug in a freeones URL, skipping a locale prefix and any /bio or /feed suffix. */
export function slugFromUrl(url: string): string {
  const segments = new URL(url).pathname.split("/").filter((segment) => segment !== "");
  if (segments.length > 1 && LOCALE_PREFIXES.has(segments[0])) segments.shift();
  const [slug] = segments;
  if (slug === undefined) return "";
  return slug;
}

/** Canonical profile URL of a performer slug; the same URL is the performer's source_url. */
export function profileUrl(slug: string): string {
  return `${BASE_URL}/${slug}`;
}

/** Stable identity of a performer, shared by the list and find capabilities. */
export function externalIdFor(slug: string): string {
  return `freeones-${slug}`;
}

/** An element's attribute value, or an empty string when the attribute is missing. */
export function attributeOf(element: HTMLElement, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}
