import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { SITE } from "./http.ts";
import { fetchDoc } from "./html.ts";
import { parseVideoCards } from "./listing.ts";

const MAX_PAGE_PAGES = 5;

// Followable eporner pages ("scene:list:page"): channels, pornstars and
// categories render the site's video grid; an uploader profile lists its
// uploads under /videos/.
const FOLLOWABLE_PAGE = /^https?:\/\/(?:www\.)?eporner\.com\/(channel|pornstar|cat|profile)\/([^/?#]+)/;

/**
 * The URL of page `page` (1-based) of a followable eporner page's video grid,
 * or "" when `url` is not one or has no such page. Channels, pornstars and
 * categories paginate as `<page url><n>/`; profiles have a single page.
 */
export function pageListingUrl(url: string, page: number): string {
  const match = url.match(FOLLOWABLE_PAGE);
  if (!match) return "";
  const [, section, slug] = match;
  if (section === "profile") {
    if (page > 1) return "";
    return `${SITE}/profile/${slug}/videos/`;
  }
  const base = `${SITE}/${section}/${slug}/`;
  if (page === 1) return base;
  return `${base}${page}/`;
}

/**
 * Lists the videos on a channel, pornstar, category or profile page, newest
 * site order, following its pagination until `target` items are collected.
 */
export async function crawlPage(url: string, target: number): Promise<DiscoveredItem[]> {
  if (!pageListingUrl(url, 1)) throw new Error(`eporner: cannot list scenes of ${url}`);

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (let page = 1; page <= MAX_PAGE_PAGES; page++) {
    const listingUrl = pageListingUrl(url, page);
    if (!listingUrl) break;
    const doc = await fetchDoc(listingUrl);
    if (!doc) {
      if (page === 1) throw new Error(`eporner: could not load ${listingUrl}`);
      break;
    }

    let added = 0;
    for (const card of parseVideoCards(doc)) {
      if (seen.has(card.external_id)) continue;
      seen.add(card.external_id);
      items.push(card);
      added++;
    }
    if (added === 0) break;
    if (items.length >= target) break;
  }
  return items;
}
