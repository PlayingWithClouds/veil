import { configuredHost, fetchJson } from "./http.ts";
import { videoToDiscoveredItem, type ApiVideo } from "./listing.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

const PAGE_SIZE = 60;
const MAX_SEARCH_PAGES = 20;
const MAX_BROWSE_PAGES = 500;

interface Videos2Response {
  videos?: ApiVideo[];
  pages?: number;
}

// A query searches the videos2.php relevance endpoint; without one the
// latest-updates catalog is paged via the videos2 API. offset skips leading
// items so callers can paginate for infinite scroll.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : 0;

  const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
  const items = await crawlVideos(maxPages, target, (page) => apiPath(query, page));
  return { items: limit > 0 ? items.slice(offset, offset + limit) : items.slice(offset) };
}

function apiPath(query: string, page: number): string {
  if (!query) {
    return `/api/json/videos2/14400/str/latest-updates/${PAGE_SIZE}/..${page}.all...jsond`;
  }
  const encoded = encodeURIComponent(query);
  return (
    `/api/videos2.php?params=14400/str/relevance/${PAGE_SIZE}/search..${page}.all..` +
    `&s=${encoded}`
  );
}

async function crawlVideos(
  maxPages: number,
  target: number,
  pathForPage: (page: number) => string
): Promise<DiscoveredItem[]> {
  const host = configuredHost();
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= maxPages; page++) {
    let response: Videos2Response;
    try {
      response = await fetchJson<Videos2Response>(host, pathForPage(page));
    } catch {
      break;
    }

    const added = collectVideos(host, response, seen, items);
    if (added === 0) break;
    if (target > 0 && items.length >= target) break;
    if (response.pages != null && page >= response.pages) break;
  }

  return items;
}

function collectVideos(
  host: string,
  response: Videos2Response,
  seen: Set<string>,
  items: DiscoveredItem[]
): number {
  let added = 0;
  for (const video of response.videos ?? []) {
    const item = videoToDiscoveredItem(host, video);
    if (!item || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
    added++;
  }
  return added;
}
