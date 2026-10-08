import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { fetchDocument, listingUrl, withPageParam } from "./http.ts";
import { parseCards } from "./listing.ts";
import { enabledSites, requireSite, type Site } from "./sites.ts";

const MAX_PAGES = 15;
const DEFAULT_TARGET = 30;

/**
 * Lists scenes. A `url` lists that channel/pornstar/category page on its own
 * site; a `query` searches every enabled site and a bare call browses their
 * newest feeds. Results from several sites are interleaved round-robin so each
 * contributes evenly, and offset/limit slice the merged list for pagination.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const limit = args.limit || 0;
  const offset = args.offset || 0;
  let target = DEFAULT_TARGET;
  if (limit > 0) target = offset + limit;

  let items: DiscoveredItem[];
  if (args.url) {
    items = await listPage(args.url, target);
  } else {
    items = await listAcrossSites(args.query ? args.query.trim() : "", target);
  }

  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

/** Lists the grid of one site page (channel, pornstar, category, tag). */
async function listPage(url: string, target: number): Promise<DiscoveredItem[]> {
  const site = requireSite(url);
  const firstPage = listingUrl(site, url);
  return crawlPages(site, (page) => withPageParam(firstPage, page), target);
}

/** Searches (or browses newest) on each enabled site in parallel and interleaves the results. */
async function listAcrossSites(query: string, target: number): Promise<DiscoveredItem[]> {
  const sites = enabledSites();
  const perSiteTarget = Math.ceil(target / sites.length);

  const outcomes = await Promise.allSettled(
    sites.map((site) => {
      let firstPage = site.newestUrl();
      if (query) firstPage = site.searchUrl(query);
      return crawlPages(site, (page) => withPageParam(firstPage, page), perSiteTarget);
    }),
  );

  const perSite: DiscoveredItem[][] = [];
  let firstFailure: unknown = undefined;
  for (const outcome of outcomes) {
    if (outcome.status === "fulfilled") {
      perSite.push(outcome.value);
    } else if (firstFailure === undefined) {
      firstFailure = outcome.reason;
    }
  }
  // One blocked site must not hide the others; only surface the error when nothing came back.
  if (perSite.length === 0 && firstFailure !== undefined) throw firstFailure;
  return interleave(perSite);
}

/** Walks consecutive listing pages until `target` items are collected or a page adds nothing new. */
async function crawlPages(
  site: Site,
  urlForPage: (page: number) => string,
  target: number,
): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= MAX_PAGES; page++) {
    let cards: DiscoveredItem[];
    try {
      cards = parseCards(site, await fetchDocument(urlForPage(page)));
    } catch (error) {
      // A failing first page is a real error; a failing later page just ends the list.
      if (page === 1) throw error;
      break;
    }

    let added = 0;
    for (const card of cards) {
      if (seen.has(card.external_id)) continue;
      seen.add(card.external_id);
      items.push(card);
      added++;
    }
    if (added === 0) break;
    if (items.length >= target) break;
  }
  return items;
}

/** Merges lists by taking one item from each in turn. */
export function interleave(lists: DiscoveredItem[][]): DiscoveredItem[] {
  const merged: DiscoveredItem[] = [];
  const longest = Math.max(0, ...lists.map((list) => list.length));
  for (let index = 0; index < longest; index++) {
    for (const list of lists) {
      if (index < list.length) merged.push(list[index]);
    }
  }
  return merged;
}
