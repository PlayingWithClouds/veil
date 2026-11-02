import { parse } from "node-html-parser";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 10;
const MAX_BROWSE_PAGES = 500;
const MAX_PROFILE_PAGES = 200;

// A url lists that actress page's videos (paginated via ?page=N). A query
// searches /en/search/<query> (paginated via ?page=N); without one the
// newest-feed catalog at /en/new?page=<n> is crawled. offset skips leading
// items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  let maxPages = MAX_BROWSE_PAGES;
  if (args.url) {
    maxPages = MAX_PROFILE_PAGES;
  } else if (query) {
    maxPages = MAX_SEARCH_PAGES;
  }
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= maxPages; page++) {
    let html: string;
    try {
      html = await fetchPage(pageUrl(args.url, query, page));
    } catch {
      break;
    }

    const added = collectCards(html, seen, items);
    if (added === 0) break;
    if (target > 0 && items.length >= target) break;
  }

  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

/** Page URL: a profile url (query string and fragment dropped), else the search or newest feed. Page 1 has no ?page; past the last page the site answers 200 with no cards. */
export function pageUrl(url: string | undefined, query: string, page: number): string {
  let base = `${BASE_URL}/en/new`;
  if (url) {
    base = url.replace(/[?#].*$/, "").replace(/\/+$/, "");
  } else if (query) {
    base = `${BASE_URL}/en/search/${encodeURIComponent(query)}`;
  }
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
