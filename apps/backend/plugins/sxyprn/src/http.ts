import { fetchHtml, resolveUrl as resolveAgainst } from "@playingwithclouds/veil-sdk";
import type { HTMLElement } from "node-html-parser";

export const SITE_HOST = "sxyprn.com";
export const BASE_URL = `https://${SITE_HOST}`;

/** Fetches a page with a plain request; the site answers without a Cloudflare challenge. */
export function fetchPage(url: string): Promise<string> {
  return fetchHtml(url);
}

/** Resolves a path, absolute URL or protocol-relative URL against the site. */
export function resolveUrl(path: string): string {
  return resolveAgainst(BASE_URL, path);
}

/** The post page of a post id such as 6ac9cd21eb1d5. */
export function postUrl(postId: string): string {
  return `${BASE_URL}/post/${postId}.html`;
}

/** The post id in a post URL, or an empty string when the URL has none. */
export function postIdFromUrl(url: string): string {
  const match = url.match(/\/post\/([0-9a-f]+)\.html/i);
  if (!match) return "";
  return match[1];
}

/** The identity a scene carries for a post id. */
export function externalIdOf(postId: string): string {
  return `sxyprn-${postId}`;
}

/** The blog (uploader) id in a /blog/<id>/... URL, or an empty string. */
export function blogIdFromUrl(url: string): string {
  const match = url.match(/\/blog\/([0-9a-f]+)\//i);
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
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

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
