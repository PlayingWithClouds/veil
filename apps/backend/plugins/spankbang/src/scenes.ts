import { parse } from "node-html-parser";
import { BASE_URL, fetchPage } from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import { ItemSink, type ListArgs, type ItemsResult } from "@playingwithclouds/veil-sdk";

const MAX_SEARCH_PAGES = 10;
const MAX_BROWSE_PAGES = 200;

// A query searches /s/<query>/ (paginated as /s/<query>/<page>/, spaces as
// '+'); without one the newest-uploads feed at /new_videos/<page>/ is crawled.
// offset skips leading items so callers can paginate for infinite scroll.
// Each page's cards are emitted as soon as it is parsed.
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const maxPages = query ? MAX_SEARCH_PAGES : MAX_BROWSE_PAGES;
  const sink = new ItemSink(args);
  await crawlCards(maxPages, sink, (page) => pageUrl(query, page));
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

/** Feeds listing pages into sink until one adds nothing new or the sink is full. */
async function crawlCards(
  maxPages: number,
  sink: ItemSink,
  urlForPage: (page: number) => string
): Promise<void> {
  for (let page = 1; page <= maxPages; page++) {
    let html: string;
    try {
      html = await fetchPage(urlForPage(page));
    } catch {
      break;
    }

    const cards = parseVideoCards(parse(html));
    if (cards.length === 0) break;
    if (sink.add(cards) === 0) break;
    if (sink.full()) break;
  }
}
