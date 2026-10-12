import { parse } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { BASE_URL, SITE_HOST, fetchPage } from "./http.ts";
import { parseVideoCards } from "./cards.ts";

// The site serves 30 posts per page; its page parameter is the offset of the first post.
const PAGE_SIZE = 30;
const DEFAULT_LIMIT = PAGE_SIZE;
// Upper bound on the pages one call may fetch, so a large limit cannot flood the site.
const MAX_PAGES_PER_CALL = 10;
const DEFAULT_BROWSE_PATH = "/latest-videos.html";
// Characters the site's own search box turns into spaces before it builds a keyword page.
const SEARCH_PUNCTUATION = /[`~!@#$%^&*()_|+\-=?;:'",.<>{}[\]\\/]/g;

/**
 * Lists posts: a keyword search when args.query is set, a listing page when
 * args.url is set, otherwise the site's newest posts. Offsets and limits count
 * posts, so pages are fetched as far as the limit needs.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const offset = offsetFrom(args.offset);
  const limit = limitFrom(args.limit);
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  let nextOffset = offset;
  for (let pagesFetched = 0; items.length < limit && pagesFetched < MAX_PAGES_PER_CALL; pagesFetched++) {
    const pageItems = await itemsFromOffset(args, nextOffset);
    if (pageItems.length === 0) break;
    for (const item of pageItems) {
      if (seen.has(item.external_id)) continue;
      seen.add(item.external_id);
      items.push(item);
    }
    nextOffset = pageStartOf(nextOffset) + PAGE_SIZE;
  }
  return { items: items.slice(0, limit) };
}

/** The posts of the page holding an offset, from that offset to the page's end. */
async function itemsFromOffset(args: ListArgs, offset: number): Promise<DiscoveredItem[]> {
  const pageStart = pageStartOf(offset);
  const html = await fetchPage(listingUrl(args, pageStart));
  return parseVideoCards(parse(html)).slice(offset - pageStart);
}

/** The offset of the first post on the page that holds an offset. */
function pageStartOf(offset: number): number {
  return offset - (offset % PAGE_SIZE);
}

/** The offset into the listing, zero when the request has none. */
function offsetFrom(offset: number | undefined): number {
  if (offset === undefined || offset < 0) return 0;
  return offset;
}

/** The number of posts to return, one page when the request has no limit. */
function limitFrom(limit: number | undefined): number {
  if (limit === undefined || limit <= 0) return DEFAULT_LIMIT;
  return limit;
}

/**
 * The listing page URL for a page's first post offset. The first page carries no
 * page parameter, so it is the site's own URL for that listing.
 */
export function listingUrl(args: ListArgs, pageStart: number): string {
  const target = new URL(listingBaseUrl(args));
  if (pageStart > 0) target.searchParams.set("page", String(pageStart));
  return target.toString();
}

/** The listing a request names before its page parameter is added. */
function listingBaseUrl(args: ListArgs): string {
  if (args.url !== undefined) return listingPageUrl(args.url);
  if (args.query !== undefined && args.query.trim() !== "") {
    return `${BASE_URL}/${encodeURIComponent(keywordFromQuery(args.query))}.html`;
  }
  return `${BASE_URL}${DEFAULT_BROWSE_PATH}`;
}

/**
 * A listing page given as a URL: a single-segment .html page of the site (a
 * keyword, performer or the site's own listing). Channel pages page differently and are refused.
 */
export function listingPageUrl(url: string): string {
  const target = new URL(url);
  const isListingPath = target.hostname === SITE_HOST && /^\/[^/]+\.html$/i.test(target.pathname);
  if (!isListingPath) throw new Error(`sxyprn: args.url is not a listing page of the site: ${url}`);
  return target.toString();
}

/**
 * The keyword a query searches for, as the site's search box builds it: its words
 * joined with "-", punctuation removed.
 */
export function keywordFromQuery(query: string): string {
  return query
    .replace(SEARCH_PUNCTUATION, " ")
    .trim()
    .split(/\s+/)
    .filter((word) => word !== "")
    .join("-");
}
