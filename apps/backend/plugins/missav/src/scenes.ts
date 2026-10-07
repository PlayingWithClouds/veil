import { parse } from "node-html-parser";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 10;
const MAX_BROWSE_PAGES = 500;

// A query searches /en/search/<query> (paginated via ?page=N); without one the
// newest-feed catalog at /en/new?page=<n> is crawled. offset skips leading
// items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= maxPages; page++) {
    let html: string;
    try {
      html = await fetchPage(pageUrl(query, page));
    } catch {
      break;
    }

    const added = collectCards(html, seen, items);
    if (added === 0) break;
    if (target > 0 && items.length >= target) break;
  }

  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

function pageUrl(query: string, page: number): string {
  const base = query ? `${BASE_URL}/en/search/${encodeURIComponent(query)}` : `${BASE_URL}/en/new`;
  if (page === 1) return base;
  return `${base}?page=${page}`;
}

function collectCards(html: string, seen: Set<string>, items: DiscoveredItem[]): number {
  let added = 0;
  for (const card of parseVideoCards(parse(html))) {
    if (seen.has(card.external_id)) continue;
    seen.add(card.external_id);
    items.push(card);
    added++;
  }
  return added;
}
