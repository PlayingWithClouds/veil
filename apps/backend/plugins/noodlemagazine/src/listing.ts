import { parse } from "node-html-parser";
import type { ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseVideoCards } from "./cards.ts";

// The site serves 24 videos per page; its own page parameter p is 0-based.
const PAGE_SIZE = 24;
const DEFAULT_BROWSE_PATH = "/home?range=month";

/**
 * Lists videos: a keyword search when args.query is set, a site page when
 * args.url is set, otherwise the site's monthly top list. The offset is read
 * against the 24-video pages, so offsets between pages still line up.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const offset = offsetFrom(args.offset);
  const pageNumber = Math.floor(offset / PAGE_SIZE);
  const items = parseVideoCards(parse(await fetchPage(listingUrl(args, pageNumber)))).slice(offset % PAGE_SIZE);
  if (args.limit !== undefined && args.limit > 0) return { items: items.slice(0, args.limit) };
  return { items };
}

/** The offset into the listing, zero when the request has none. */
function offsetFrom(offset: number | undefined): number {
  if (offset === undefined || offset < 0) return 0;
  return offset;
}

/**
 * The listing page for a request. Search uses the site's /video/<words> pages;
 * a site page given as args.url keeps its own path and parameters.
 */
export function listingUrl(args: ListArgs, pageNumber: number): string {
  if (args.url !== undefined) {
    if (!args.url.startsWith(BASE_URL)) throw new Error(`noodlemagazine: args.url is not a page of the site`);
    const target = new URL(args.url);
    target.searchParams.set("p", String(pageNumber));
    return target.toString();
  }
  if (args.query !== undefined && args.query.trim() !== "") {
    return `${BASE_URL}${searchPathFromQuery(args.query)}?p=${pageNumber}`;
  }
  return `${BASE_URL}${DEFAULT_BROWSE_PATH}&p=${pageNumber}`;
}

/** The search page path of a query: its words joined with "+", as the site writes them. */
export function searchPathFromQuery(query: string): string {
  const words = query.trim().split(/\s+/).map((word) => encodeURIComponent(word));
  return `/video/${words.join("+")}`;
}
