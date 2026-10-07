import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import { BASE_URL } from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 500;

// A query searches /search?what=<query> (paginated via &page=N); without one
// the newest-uploads feed at /new/<page> is crawled. offset skips leading
// items so callers can paginate for infinite scroll.
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
  if (!query) return `${BASE_URL}/new/${page}`;
  const encoded = encodeURIComponent(query);
  return `${BASE_URL}/search?what=${encoded}&page=${page}`;
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
