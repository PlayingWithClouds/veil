import { fetchHtml, type DiscoveredItem, type ItemsResult, type ListArgs } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";
import { cardsFromRoot } from "./cards.ts";
import { BASE_URL, PAGE_SIZE } from "./site.ts";

/**
 * The newest videos, or the videos matching `query` when one is set, paged by
 * offset. Listing pages are numbered from 0 with ?p=<n>.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  let limit = PAGE_SIZE;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  let offset = 0;
  if (args.offset !== undefined && args.offset > 0) offset = args.offset;

  let listingUrl = `${BASE_URL}/recent`;
  if (args.query !== undefined && args.query.trim() !== "") {
    listingUrl = `${BASE_URL}/video/${searchTerm(args.query)}`;
  }
  return { items: await collectItems(listingUrl, offset, limit) };
}

/** The search term as the site's path segment: words joined with "+". */
function searchTerm(query: string): string {
  return query
    .trim()
    .split(/\s+/)
    .map((word) => encodeURIComponent(word))
    .join("+");
}

/** Items from `offset` on, page after page until `limit` is filled or the listing runs out. */
async function collectItems(listingUrl: string, offset: number, limit: number): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  let pageIndex = Math.floor(offset / PAGE_SIZE);
  let skip = offset % PAGE_SIZE;
  while (items.length < limit) {
    const cards = cardsFromRoot(parse(await fetchHtml(pageUrl(listingUrl, pageIndex))));
    if (cards.length === 0) break;
    items.push(...cards.slice(skip, skip + limit - items.length));
    skip = 0;
    pageIndex += 1;
  }
  return items;
}

/** Listing page `pageIndex` (0-based); page 0 is the listing itself. */
function pageUrl(listingUrl: string, pageIndex: number): string {
  if (pageIndex === 0) return listingUrl;
  return `${listingUrl}?p=${pageIndex}`;
}
