import { parse } from "node-html-parser";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseProfileCards, parseVideoCards } from "./listing.ts";
import { ItemSink, type DiscoveredItem, type ListArgs, type ItemsResult } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 10;
const MAX_BROWSE_PAGES = 200;
const MAX_PROFILE_PAGES = 200;

// A url lists that pornstar/channel page's videos (/<id>/pornstar/<slug>/<page>/).
// A query searches /s/<query>/ (paginated as /s/<query>/<page>/, spaces as
// '+'); without one the newest-uploads feed at /new_videos/<page>/ is crawled.
// offset skips leading items so callers can paginate for infinite scroll.
// Each page's cards are emitted as soon as it is parsed.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const sink = new ItemSink(args);
  if (args.url) {
    await crawlCards(MAX_PROFILE_PAGES, sink, profilePageUrl(args.url), parseProfileCards);
    return sink.result();
  }
  const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
  await crawlCards(maxPages, sink, (page) => pageUrl(query, page), parseVideoCards);
  return sink.result();
}

function pageUrl(query: string, page: number): string {
  if (!query) {
    if (page === 1) return `${BASE_URL}/new_videos/`;
    return `${BASE_URL}/new_videos/${page}/`;
  }
  const slug = encodeURIComponent(query).replace(/%20/g, "+");
  if (page === 1) return `${BASE_URL}/s/${slug}/`;
  return `${BASE_URL}/s/${slug}/${page}/`;
}

/** Page URL builder for a pornstar/channel page: page 1 is the URL itself, later pages append <page>/. Past the last page the site shows no results. */
export function profilePageUrl(url: string): (page: number) => string {
  const base = url.replace(/[?#].*$/, "").replace(/\/+$/, "");
  return (page) => {
    if (page === 1) return `${base}/`;
    return `${base}/${page}/`;
  };
}

/** Feeds listing pages into sink until one adds nothing new or the sink is full. */
async function crawlCards(
  maxPages: number,
  sink: ItemSink,
  urlForPage: (page: number) => string,
  parseCards: (root: ReturnType<typeof parse>) => DiscoveredItem[]
): Promise<void> {
  for (let page = 1; page <= maxPages; page++) {
    let html: string;
    try {
      html = await fetchPage(urlForPage(page));
    } catch {
      break;
    }

    const cards = parseCards(parse(html));
    if (cards.length === 0) break;
    if (sink.add(cards) === 0) break;
    if (sink.full()) break;
  }
}
