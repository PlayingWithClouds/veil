import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import type { SiteConfig } from "./sites.ts";
import { fetchPage, fetchPageAt } from "./http.ts";
import { parseListingCards } from "./listing.ts";
import { pathPageUrl } from "./urls.ts";

/**
 * KVS's ajax pagination: a listing block re-rendered by
 * `?mode=async&function=get_block&block_id=<id>&<parameters>`, the page
 * number going into the block's "from" parameter(s).
 */
export interface AsyncBlock {
  blockId: string;
  /** The block's parameters, page-independent ones verbatim. */
  parameters: [string, string][];
  /** Parameter names that take the page number ("from", "from_videos", …). */
  pageKeys: string[];
}

/** A pagination key: "from" or "from_<list>", never e.g. "duration_from". */
const PAGE_KEY = /^from(_\w+)?$/;

/**
 * The async pagination block of a listing page, read off its pager links'
 * `data-block-id`/`data-parameters`; undefined when the page paginates by
 * path only. A video list's block wins over side blocks (model lists, …).
 */
export function asyncBlock(root: HTMLElement): AsyncBlock | undefined {
  let fallback: AsyncBlock | undefined;
  for (const element of root.querySelectorAll("[data-block-id][data-parameters]")) {
    const block = parseBlock(element.getAttribute("data-block-id") ?? "", element.getAttribute("data-parameters") ?? "");
    if (!block) continue;
    if (block.blockId.includes("video")) return block;
    if (!fallback) fallback = block;
  }
  return fallback;
}

/** Parses "q:blonde;sort_by:;from_videos+from_albums:02" into a block; undefined without a page key. */
export function parseBlock(blockId: string, rawParameters: string): AsyncBlock | undefined {
  if (!blockId) return undefined;
  const parameters: [string, string][] = [];
  const pageKeys: string[] = [];

  for (const pair of rawParameters.split(";")) {
    const separator = pair.indexOf(":");
    if (separator <= 0) continue;
    const value = safeDecode(pair.slice(separator + 1));
    for (const name of pair.slice(0, separator).split("+")) {
      if (PAGE_KEY.test(name)) {
        pageKeys.push(name);
        continue;
      }
      parameters.push([name, value]);
    }
  }
  if (pageKeys.length === 0) return undefined;
  return { blockId, parameters, pageKeys };
}

/** The URL of page `page` of `firstPageUrl`'s listing via its async block. */
export function asyncPageUrl(firstPageUrl: string, block: AsyncBlock, page: number): string {
  const url = new URL(firstPageUrl);
  url.searchParams.set("mode", "async");
  url.searchParams.set("function", "get_block");
  url.searchParams.set("block_id", block.blockId);
  for (const [name, value] of block.parameters) {
    url.searchParams.set(name, value);
  }
  for (const name of block.pageKeys) {
    url.searchParams.set(name, String(page));
  }
  return url.toString();
}

/** Reads the entries (video cards, catalog entries) off one listing page. */
export type PageParser = (site: SiteConfig, root: HTMLElement) => DiscoveredItem[];

/**
 * Collects the entries of a listing from its first page on, until `target`
 * items are found (0 = the first page only), `maxPages` is reached, or a page
 * adds nothing new. Later pages follow the listing's async block when it has
 * one, else its path pagination.
 */
export async function crawlListing(
  site: SiteConfig,
  firstPageUrl: string,
  target: number,
  maxPages: number,
  parseItems: PageParser = parseListingCards
): Promise<DiscoveredItem[]> {
  const first = await fetchPageAt(site, firstPageUrl);
  const firstRoot = parse(first.html);
  const block = asyncBlock(firstRoot);
  const items = parseItems(site, firstRoot);
  const seen = new Set(items.map((item) => item.external_id));
  // A redirect (to a mirror host) drops the query, so later pages build on the served URL.
  const baseUrl = servedListingUrl(firstPageUrl, first.url);

  for (let page = 2; page <= maxPages && target > 0 && items.length < target; page++) {
    const added = await crawlPage(site, listingPageUrl(baseUrl, block, page), seen, parseItems);
    if (added.length === 0) break;
    items.push(...added);
  }
  return items;
}

/** The requested listing URL moved onto the host it was served from, keeping its query. */
function servedListingUrl(requestedUrl: string, servedUrl: string): string {
  const served = new URL(servedUrl);
  const requested = new URL(requestedUrl);
  if (served.host === requested.host) return requestedUrl;
  requested.protocol = served.protocol;
  requested.host = served.host;
  return requested.toString();
}

/** The URL of page `page` of a listing. */
function listingPageUrl(firstPageUrl: string, block: AsyncBlock | undefined, page: number): string {
  if (block) return asyncPageUrl(firstPageUrl, block, page);
  return pathPageUrl(firstPageUrl, page);
}

/** The not-yet-seen entries of one later listing page; none when it fails to load. */
async function crawlPage(
  site: SiteConfig,
  url: string,
  seen: Set<string>,
  parseItems: PageParser
): Promise<DiscoveredItem[]> {
  let html: string;
  try {
    html = await fetchPage(site, url);
  } catch {
    return [];
  }
  const added: DiscoveredItem[] = [];
  for (const item of parseItems(site, parse(html))) {
    if (seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    added.push(item);
  }
  return added;
}

/** decodeURIComponent that leaves malformed escapes as they are. */
function safeDecode(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}
