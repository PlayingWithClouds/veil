import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://www.erocurves.com";
// The site is WordPress; its REST API serves posts (one gallery each) as JSON.
export const API_URL = `${BASE_URL}/wp-json/wp/v2`;

/** Fetches a WordPress REST URL and returns its parsed JSON body. */
export async function fetchJson<Body>(url: string): Promise<Body> {
  const response = await fetch(url, {
    headers: { ...BROWSER_HEADERS, Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`erocurves ${response.status} for ${url}`);
  }
  return (await response.json()) as Body;
}

// Named entities WordPress uses in titles and excerpts, plus numeric ones.
const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Decodes HTML entities in a WordPress text field. */
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

/** Removes HTML tags, keeping the text between them. */
export function stripTags(html: string): string {
  return html.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
}
