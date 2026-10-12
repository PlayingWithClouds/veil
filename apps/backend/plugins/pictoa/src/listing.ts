import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";
import { BASE_URL } from "./http.ts";
import { albumsFromRoot } from "./cards.ts";

// Listing and search pages both serve 72 albums each.
const PAGE_SIZE = 72;

/**
 * Album stubs: the newest albums, or the site's search for a query, paged by
 * offset. Pages after the first are /page<N>.html and /s/<term>/page<N>.html.
 */
export async function galleryList(args: ListArgs): Promise<ItemsResult> {
  let limit = PAGE_SIZE;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  let offset = 0;
  if (args.offset !== undefined && args.offset > 0) offset = args.offset;
  const searchPath = searchPathFromQuery(args.query);

  const items: DiscoveredItem[] = [];
  let pageNumber = Math.floor(offset / PAGE_SIZE) + 1;
  let skip = offset % PAGE_SIZE;
  while (items.length < limit) {
    const pageItems = await fetchListingPage(searchPath, pageNumber);
    if (pageItems.length === 0) break;
    items.push(...pageItems.slice(skip).slice(0, limit - items.length));
    skip = 0;
    pageNumber += 1;
  }
  return { items };
}

/** The search path for a query ("wood sex" becomes /s/wood-sex/), or "" for the newest-albums listing. */
export function searchPathFromQuery(query: string | undefined): string {
  if (query === undefined) return "";
  const terms = query.trim().toLowerCase().split(/\s+/).filter((term) => term !== "");
  if (terms.length === 0) return "";
  return `/s/${terms.map((term) => encodeURIComponent(term)).join("-")}/`;
}

/** The album stubs on one page of the listing or search; page 1 has no page suffix. */
async function fetchListingPage(searchPath: string, pageNumber: number): Promise<DiscoveredItem[]> {
  let path = "/";
  if (searchPath !== "") path = searchPath;
  if (pageNumber > 1) path = `${path}page${pageNumber}.html`;
  const html = await fetchHtml(BASE_URL + path);
  return albumsFromRoot(parse(html));
}
