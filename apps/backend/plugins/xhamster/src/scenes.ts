import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import { BASE_URL } from "./http.ts";
import { parseVideoCards, thumbsFromInitials } from "./listing.ts";
import { parseInitials } from "./initials.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 500;
const MAX_PROFILE_PAGES = 200;

// A url lists that pornstar/channel page's videos (page N lives at <url>/N). A
// query searches /search/<query> (paginated via ?page=N, spaces map to '+');
// without one the newest feed at /newest/<page> is crawled. offset skips
// leading items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  let maxPages = MAX_BROWSE_PAGES;
  let urlForPage = (page: number) => pageUrl(query, page);
  if (args.url) {
    maxPages = MAX_PROFILE_PAGES;
    urlForPage = profilePageUrl(args.url);
  } else if (query) {
    maxPages = MAX_SEARCH_PAGES;
  }
  const items = await crawlCards(maxPages, target, urlForPage);
  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

/** Page URL builder for a profile or channel page: page 1 is the URL itself, later pages append /N. Past the last page the site answers 404. */
export function profilePageUrl(url: string): (page: number) => string {
  const base = url.replace(/[?#].*$/, "").replace(/\/+$/, "");
  return (page) => {
    if (page === 1) return base;
    return `${base}/${page}`;
  };
}

function pageUrl(query: string, page: number): string {
  if (!query) {
    if (page === 1) return `${BASE_URL}/newest`;
    return `${BASE_URL}/newest/${page}`;
  }
  const slug = encodeURIComponent(query).replace(/%20/g, "+");
  if (page === 1) return `${BASE_URL}/search/${slug}`;
  return `${BASE_URL}/search/${slug}?page=${page}`;
}

async function crawlCards(
  maxPages: number,
  target: number,
  urlForPage: (page: number) => string
): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= maxPages; page++) {
    let html: string;
    try {
      html = await fetchHtml(urlForPage(page));
    } catch {
      break;
    }

    const cards = parseVideoCards(parse(html), thumbsFromInitials(parseInitials(html)));
    if (cards.length === 0) break;

    let added = 0;
    for (const card of cards) {
      if (seen.has(card.external_id)) continue;
      seen.add(card.external_id);
      items.push(card);
      added++;
    }

    if (added === 0) break;
    if (target > 0 && items.length >= target) break;
  }

  return items;
}
