import { parse } from "node-html-parser";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 10;
const MAX_BROWSE_PAGES = 200;

// A query searches /s/<query>/ (paginated as /s/<query>/<page>/, spaces as
// '+'); without one the newest-uploads feed at /new_videos/<page>/ is crawled.
// offset skips leading items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
  const items = await crawlCards(maxPages, target, (page) => pageUrl(query, page));
  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

function pageUrl(query: string, page: number): string {
  if (!query) {
    if (page === 1) return `${BASE_URL}/new_videos/`;
    return `${BASE_URL}/new_videos/${page}/`;
  }
  const slug = encodeURIComponent(query).replace(/%20/g, "+");
  if (page === 1) return `${BASE_URL}/s/${slug}/`;
  return `${BASE_URL}/s/${slug}/${page}/`;
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
      html = await fetchPage(urlForPage(page));
    } catch {
      break;
    }

    const cards = parseVideoCards(parse(html));
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
