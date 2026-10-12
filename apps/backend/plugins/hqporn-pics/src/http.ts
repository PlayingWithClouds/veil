import { resolveUrl as resolveAgainst } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://hqporn.pics";

/** Resolves a page or asset path (absolute, protocol-relative or root-relative) against the site. */
export function resolveUrl(path: string): string {
  return resolveAgainst(BASE_URL, path);
}

// Entities the site writes into attributes and headings.
const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Decodes HTML entities in attribute values and headings. */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => {
    if (name.startsWith("#x") || name.startsWith("#X")) {
      return String.fromCodePoint(parseInt(name.slice(2), 16));
    }
    if (name.startsWith("#")) return String.fromCodePoint(parseInt(name.slice(1), 10));
    const decoded = NAMED_ENTITIES[name.toLowerCase()];
    if (decoded === undefined) return match;
    return decoded;
  });
}

/** An element's attribute value, or an empty string when the attribute is missing. */
export function attributeOf(element: { getAttribute(name: string): string | undefined }, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}

/** The last path segment of a gallery URL (/galls/<slug>), or "" when the URL is not a gallery. */
export function gallerySlugFromUrl(url: string): string {
  const match = new URL(resolveUrl(url)).pathname.match(/^\/galls\/([^/]+)\/?$/);
  if (match === null) return "";
  return match[1];
}
