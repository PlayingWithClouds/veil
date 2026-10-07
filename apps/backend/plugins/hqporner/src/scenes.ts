import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import {
  BASE_URL,
  actressSlugFromUrl,
  categorySlugFromUrl,
  isActressUrl,
  isCategoryUrl,
  resolveUrl,
} from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 500;
const MAX_PAGE_PAGES = 5;

// A url lists that actress or category page's videos. A query searches
// /?q=<query> (paginated via &p=N); without one the HD catalog at
// /hdporn/<page> is crawled. offset skips leading items so callers can
// paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  let items: DiscoveredItem[];
  if (args.url) {
    items = await crawlFollowablePage(args.url, target);
  } else {
    const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
    items = await crawlCards(maxPages, target, (page) => pageUrl(query, page));
  }
  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

/**
 * The URL of page `page` (1-based) of an actress or category page, or ""
 * when `url` is neither. Both paginate as `/<kind>/<slug>/<page>`.
 */
export function followablePageUrl(url: string, page: number): string {
  if (!isHqpornerUrl(url)) return "";
  let base = "";
  if (isActressUrl(url)) base = `${BASE_URL}/actress/${actressSlugFromUrl(url)}`;
  if (isCategoryUrl(url)) base = `${BASE_URL}/category/${categorySlugFromUrl(url)}`;
  if (!base) return "";
  if (page === 1) return base;
  return `${base}/${page}`;
}

/** Whether `url` (absolute or site-relative) points at hqporner.com. */
function isHqpornerUrl(url: string): boolean {
  try {
    return new URL(resolveUrl(url)).hostname.replace(/^www\./, "") === "hqporner.com";
  } catch {
    return false;
  }
}

/** Crawls an actress or category page's video grid; throws for other URLs. */
async function crawlFollowablePage(url: string, target: number): Promise<DiscoveredItem[]> {
  if (!followablePageUrl(url, 1)) throw new Error(`hqporner: cannot list scenes of ${url}`);
  const items = await crawlCards(MAX_PAGE_PAGES, target, (page) => followablePageUrl(url, page));
  if (items.length === 0) throw new Error(`hqporner: no videos on ${url}`);
  return items;
}

function pageUrl(query: string, page: number): string {
  if (!query) return `${BASE_URL}/hdporn/${page}`;
  const encoded = encodeURIComponent(query);
  if (page === 1) return `${BASE_URL}/?q=${encoded}`;
  return `${BASE_URL}/?q=${encoded}&p=${page}`;
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

    const cards = parseVideoCards(parse(html), html);
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
