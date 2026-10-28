import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import {
  HOME_URL,
  BASE_URL,
  fetchPage,
  isShowcamripsUrl,
  listingPageFromUrl,
  listingPageUrl,
  pageUrl,
  type ListingPage,
} from "./http.ts";
import { parseVideoCards } from "./listing.ts";

const MAX_PAGES = 50;

/**
 * Lists recordings. A `url` lists that model, cam site or category page; a
 * query looks the name up in the site's model search (which only knows whole
 * model names) and, failing that, as a cam site or category; without either
 * the newest recordings are listed. `offset` skips leading items so callers
 * can page.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const limit = args.limit || 0;
  const offset = args.offset || 0;
  let wanted = 0;
  if (limit > 0) wanted = offset + limit;

  const query = (args.query || "").trim();
  let items: DiscoveredItem[];
  if (args.url) {
    items = await listPage(args.url, wanted);
  } else if (query) {
    items = await search(query, wanted);
  } else {
    items = await crawl(HOME_URL, wanted);
  }
  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

/** Crawls a model, cam site or category page; throws for other URLs. */
async function listPage(url: string, wanted: number): Promise<DiscoveredItem[]> {
  const page = listingPageFromUrl(url);
  if (!page || !isShowcamripsUrl(url)) throw new Error(`showcamrips: cannot list scenes of ${url}`);
  const items = await crawl(listingPageUrl(page), wanted);
  if (items.length === 0) throw new Error(`showcamrips: no recordings on ${url}`);
  return items;
}

/**
 * Recordings matching a query: the model search first (spaces also tried as
 * underscores, how most cam usernames are written), then a cam site or
 * category of that name. Nothing matching yields an empty list.
 */
async function search(query: string, wanted: number): Promise<DiscoveredItem[]> {
  const spellings = [query];
  if (query.includes(" ")) spellings.push(query.replace(/\s+/g, "_"));
  for (const spelling of spellings) {
    const items = await crawlSearch(spelling, wanted);
    if (items.length > 0) return items;
  }
  const page = await findNamedPage(query);
  if (!page) return [];
  return crawl(listingPageUrl(page), wanted);
}

/**
 * The model search for `name`. Its first page is the search result; the
 * pager of that page leads on to the model's own listing, so later pages
 * come from there.
 */
async function crawlSearch(name: string, wanted: number): Promise<DiscoveredItem[]> {
  const searchUrl = `${BASE_URL}/search.php?Src=${encodeURIComponent(name)}&lg=en`;
  const root = parse(await fetchPage(searchUrl));
  const items = parseVideoCards(root);
  if (items.length === 0 || (wanted > 0 && items.length >= wanted)) return items;

  const continuation = pagerBase(root);
  if (!continuation) return items;
  let remaining = 0;
  if (wanted > 0) remaining = wanted - items.length;
  const rest = await crawl(continuation, remaining, 2);
  return mergeUnique(items, rest);
}

/**
 * The first page URL behind the pager, i.e. the "2" link without its
 * `2-pg/` tail; undefined when the page has no second page.
 */
function pagerBase(root: HTMLElement): string | undefined {
  for (const link of root.querySelectorAll(".pages a")) {
    const href = link.getAttribute("href") || "";
    if (link.text.trim() === "2" && href.endsWith("2-pg/")) return href.slice(0, -"2-pg/".length);
  }
  return undefined;
}

/** The cam site or category named exactly `query`, from the home page's menus. */
async function findNamedPage(query: string): Promise<ListingPage | undefined> {
  const root = parse(await fetchPage(HOME_URL));
  for (const link of root.querySelectorAll(".az a")) {
    const page = listingPageFromUrl(link.getAttribute("href") || "");
    if (!page || page.kind === "model") continue;
    if (page.name.toLowerCase() === query.toLowerCase()) return page;
  }
  return undefined;
}

/**
 * Crawls a listing from `startPage`, following `<first page>/<n>-pg/`, until
 * `wanted` items are in (0 = every page) or a page adds nothing.
 */
async function crawl(firstPageUrl: string, wanted: number, startPage = 1): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (let page = startPage; page <= MAX_PAGES; page++) {
    const added = await addPage(pageUrl(firstPageUrl, page), items, seen);
    if (added === 0) break;
    if (wanted > 0 && items.length >= wanted) break;
  }
  return items;
}

/** Appends the new cards of one listing page to `items`; the count added (0 also when the page fails). */
async function addPage(url: string, items: DiscoveredItem[], seen: Set<string>): Promise<number> {
  let html: string;
  try {
    html = await fetchPage(url);
  } catch {
    return 0;
  }
  let added = 0;
  for (const card of parseVideoCards(parse(html))) {
    if (seen.has(card.external_id)) continue;
    seen.add(card.external_id);
    items.push(card);
    added++;
  }
  return added;
}

/** `first` followed by the items of `rest` it does not already hold. */
function mergeUnique(first: DiscoveredItem[], rest: DiscoveredItem[]): DiscoveredItem[] {
  const known = new Set(first.map((item) => item.external_id));
  return [...first, ...rest.filter((item) => !known.has(item.external_id))];
}
