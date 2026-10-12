import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";
import { BASE_URL } from "./http.ts";
import { cardsFromRoot } from "./cards.ts";

// The video catalog serves 24 scenes per page.
const PAGE_SIZE = 24;
// The site's search only finds performer pages, so a query filters the newest
// catalog pages instead; this many pages are scanned (192 scenes).
const SEARCH_SCAN_PAGES = 8;

/**
 * The site's video catalog, newest first, paged by offset. A query has no site
 * search to ask, so it filters the newest scanned catalog pages by performer
 * name and thumbnail description.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  let limit = PAGE_SIZE;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  let offset = 0;
  if (args.offset !== undefined && args.offset > 0) offset = args.offset;
  let query = "";
  if (args.query !== undefined) query = args.query.trim();

  if (query !== "") return { items: await searchScannedPages(query, offset, limit) };
  return { items: await browseCatalog(offset, limit) };
}

/** The catalog from an offset, page after page until the limit is filled or the catalog ends. */
async function browseCatalog(offset: number, limit: number): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  let pageNumber = Math.floor(offset / PAGE_SIZE) + 1;
  let skip = offset % PAGE_SIZE;
  while (items.length < limit) {
    const pageItems = cardsFromRoot(parse(await fetchCatalogHtml(pageNumber)));
    if (pageItems.length === 0) break;
    items.push(...pageItems.slice(skip).slice(0, limit - items.length));
    skip = 0;
    pageNumber += 1;
  }
  return items;
}

/** The scanned catalog pages' videos matching a query, from an offset. */
async function searchScannedPages(query: string, offset: number, limit: number): Promise<DiscoveredItem[]> {
  const matches: DiscoveredItem[] = [];
  for (let pageNumber = 1; pageNumber <= SEARCH_SCAN_PAGES; pageNumber++) {
    matches.push(...cardsFromRoot(parse(await fetchCatalogHtml(pageNumber)), query));
  }
  return matches.slice(offset, offset + limit);
}

/** The HTML of one catalog page; page 1 is /videos/, later pages /videos/<n>/. */
async function fetchCatalogHtml(pageNumber: number): Promise<string> {
  let path = "/videos/";
  if (pageNumber > 1) path = `/videos/${pageNumber}/`;
  return fetchHtml(BASE_URL + path);
}
