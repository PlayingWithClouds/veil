import { apiSearch, toDiscoveredItem, type EpOrder, type EpVideo } from "./api.ts";
import { crawlPage } from "./pages.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 100;
const DEFAULT_SEARCH_TARGET = 30;

// A url lists that channel/pornstar/category/profile page's videos. A query
// searches the API relevance-ordered ("most-popular"); without one the
// newest-videos feed (empty API query, order=latest) is crawled. offset skips
// leading items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;
  // Page listings and searches stop early; the newest feed may crawl it all.
  let boundedTarget = target;
  if (boundedTarget <= 0) boundedTarget = DEFAULT_SEARCH_TARGET;

  let items: DiscoveredItem[];
  if (args.url) {
    items = await crawlPage(args.url, boundedTarget);
  } else if (query) {
    items = await crawlVideos(query, "most-popular", MAX_SEARCH_PAGES, boundedTarget);
  } else {
    items = await crawlVideos("", "latest", MAX_BROWSE_PAGES, target);
  }
  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

async function crawlVideos(
  query: string,
  order: EpOrder,
  maxPages: number,
  target: number
): Promise<DiscoveredItem[]> {
  const perPage = target > 0 ? Math.min(1000, target) : 100;
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= maxPages; page++) {
    let videos: EpVideo[];
    try {
      videos = await apiSearch(query, perPage, page, order);
    } catch {
      break;
    }
    if (videos.length === 0) break;

    let added = 0;
    for (const video of videos) {
      if (seen.has(video.id)) continue;
      seen.add(video.id);
      items.push(toDiscoveredItem(video));
      added++;
    }
    if (added === 0) break;
    if (target > 0 && items.length >= target) break;
  }

  return items;
}
