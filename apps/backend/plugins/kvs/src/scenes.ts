import type { ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { enabledSites, siteForUrl } from "./sites.ts";
import { crawlListing } from "./crawl.ts";
import { listAcrossSites, slice } from "./fanout.ts";
import { searchUrl } from "./urls.ts";

const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 200;
const MAX_LISTING_PAGES = 20;

/**
 * scene:list. A url lists that page of its site (channel, model, category,
 * tag or any other video listing, paginated); a query searches every enabled
 * site; neither browses their newest videos. Multi-site results interleave.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  if (args.url) return listPage(args);

  const query = (args.query ?? "").trim();
  if (query) {
    return listAcrossSites(enabledSites(), args, (site, target) =>
      crawlListing(site, searchUrl(site, query), target, MAX_SEARCH_PAGES)
    );
  }
  return listAcrossSites(enabledSites(), args, (site, target) =>
    crawlListing(site, site.origin + site.latestPath, target, MAX_BROWSE_PAGES)
  );
}

/** The videos listed on one site page (scene:list:page). */
async function listPage(args: ListArgs): Promise<ItemsResult> {
  const url = args.url as string;
  const site = siteForUrl(url);
  if (!site) throw new Error(`kvs: cannot list scenes of ${url}`);

  const offset = args.offset ?? 0;
  const limit = args.limit ?? 0;
  let target = 0;
  if (limit > 0) target = offset + limit;
  const pageUrl = new URL(url);
  pageUrl.hash = "";
  const items = await crawlListing(site, pageUrl.toString(), target, MAX_LISTING_PAGES);
  if (items.length === 0) throw new Error(`kvs: no videos on ${url}`);
  return { items: slice(items, offset, limit) };
}
