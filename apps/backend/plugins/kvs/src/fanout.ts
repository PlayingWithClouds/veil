import { emitItems, type DiscoveredItem, type ItemsResult, type ListArgs } from "@playingwithclouds/veil-sdk";
import type { SiteConfig } from "./sites.ts";

/** Lists one site's entries, collecting at least `target` when it can (0 = one page). */
export type SiteLister = (site: SiteConfig, target: number) => Promise<DiscoveredItem[]>;

/**
 * Runs a listing on every site in parallel and interleaves the results
 * round-robin, so no single site floods the first page. Each site collects
 * its share of offset+limit; a failing site is left out, and only when every
 * site fails does the listing fail. On the first page each site's share is
 * emitted as soon as that site answers; later pages depend on every site's
 * count, so they arrive whole.
 */
export async function listAcrossSites(sites: SiteConfig[], args: ListArgs, lister: SiteLister): Promise<ItemsResult> {
  if (sites.length === 0) throw new Error("kvs: no sites enabled");
  const offset = args.offset ?? 0;
  const limit = args.limit ?? 0;
  let perSiteTarget = 0;
  if (limit > 0) perSiteTarget = Math.ceil((offset + limit) / sites.length);

  const outcomes = await Promise.all(
    sites.map((site) => settle(lister(site, perSiteTarget).then((items) => emitFirstPageShare(items, offset, perSiteTarget))))
  );
  const lists: DiscoveredItem[][] = [];
  let firstError: unknown;
  for (const outcome of outcomes) {
    if (outcome.items) {
      lists.push(outcome.items);
      continue;
    }
    if (firstError === undefined) firstError = outcome.error;
  }
  if (lists.length === 0) throw firstError;

  return { items: slice(interleave(lists), offset, limit) };
}

/** Emits a site's share of the first page right away and passes its items on. */
function emitFirstPageShare(items: DiscoveredItem[], offset: number, perSiteTarget: number): DiscoveredItem[] {
  if (offset > 0) return items;
  if (perSiteTarget > 0) emitItems(items.slice(0, perSiteTarget));
  else emitItems(items);
  return items;
}

/** A listing's items or its error, never rejecting. */
async function settle(listing: Promise<DiscoveredItem[]>): Promise<{ items?: DiscoveredItem[]; error?: unknown }> {
  try {
    return { items: await listing };
  } catch (error) {
    return { error };
  }
}

/** The lists merged round-robin: first of each, then second of each, and so on. */
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

/** items[offset, offset+limit), or everything from offset when limit is 0. */
export function slice(items: DiscoveredItem[], offset: number, limit: number): DiscoveredItem[] {
  if (limit > 0) return items.slice(offset, offset + limit);
  return items.slice(offset);
}
