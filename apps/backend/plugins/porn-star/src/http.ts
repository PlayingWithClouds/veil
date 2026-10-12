import { resolveUrl as resolveAgainst } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://www.porn-star.com";

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

/** Seconds for a "SS", "MM:SS" or "HH:MM:SS" duration, or undefined when the text is not one. */
export function durationSeconds(text: string): number | undefined {
  const match = text.match(/^(\d+)(?::(\d+))?(?::(\d+))?$/);
  if (!match) return undefined;
  if (match[3] !== undefined) return Number(match[1]) * 3600 + Number(match[2]) * 60 + Number(match[3]);
  if (match[2] !== undefined) return Number(match[1]) * 60 + Number(match[2]);
  return Number(match[1]);
}

/** An element's attribute value, or an empty string when the attribute is missing. */
export function attributeOf(element: { getAttribute(name: string): string | undefined }, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}
