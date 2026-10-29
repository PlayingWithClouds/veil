import { ItemSink, type DiscoveredItem, type ItemsResult, type ListArgs } from "@playingwithclouds/veil-sdk";
import { jsonItem, jsonVideoList, parseCardsHtml } from "./cards.ts";
import { canonicalUrl, fetchJson, fetchText, ORIGINS, siteOf } from "./http.ts";

const MAX_SEARCH_PAGES = 15;
const MAX_BROWSE_PAGES = 200;
const MAX_PROFILE_PAGES = 100;

/** Profile kinds whose videos the site lists through its JSON endpoint. */
const PROFILE_KINDS = ["pornstars", "models", "channels", "profiles", "amateurs"];

/**
 * Lists xvideos videos. A `url` lists that page: a tag (/tags/blonde), a
 * pornstar, channel or uploader profile; a query searches the site; without
 * either the newest uploads are listed. Each page's cards are emitted as soon
 * as they are read; `offset`/`limit` page through the result.
 */
export async function sceneList(args: ListArgs): Promise<ItemsResult> {
  const sink = new ItemSink(args);
  const query = (args.query || "").trim();
  if (args.url) {
    await listPage(args.url, sink);
  } else if (query) {
    await crawlHtml(MAX_SEARCH_PAGES, sink, (page) => searchPageUrl(query, page));
  } else {
    await crawlHtml(MAX_BROWSE_PAGES, sink, newestPageUrl);
  }
  return sink.result();
}

/** Search results page `page` (0-based) for a query. */
function searchPageUrl(query: string, page: number): string {
  const params = new URLSearchParams({ k: query });
  if (page > 0) params.set("p", String(page));
  return `${ORIGINS.xvideos}/?${params}`;
}

/** The newest uploads, page `page` (0-based): the front page, then /new/N. */
function newestPageUrl(page: number): string {
  if (page === 0) return `${ORIGINS.xvideos}/`;
  return `${ORIGINS.xvideos}/new/${page}`;
}

/** Lists one of the site's pages; throws for pages this plugin cannot list. */
async function listPage(url: string, sink: ItemSink): Promise<void> {
  if (siteOf(url) !== "xvideos") throw new Error(`xvideos: cannot list ${url}`);
  const path = new URL(canonicalUrl(url)).pathname.replace(/\/+$/, "");
  const tag = path.match(/^\/tags\/([^/]+)/);
  if (tag) {
    await crawlHtml(MAX_BROWSE_PAGES, sink, (page) => tagPageUrl(tag[1], page));
    return;
  }
  const profile = profilePath(path);
  if (!profile) throw new Error(`xvideos: cannot list ${url}`);
  await crawlProfile(profile, sink, url);
}

/** A tag's page `page` (0-based). */
function tagPageUrl(tag: string, page: number): string {
  if (page === 0) return `${ORIGINS.xvideos}/tags/${tag}`;
  return `${ORIGINS.xvideos}/tags/${tag}/${page}`;
}

/**
 * The profile a page path stands for ("/pornstars/nicole_aniston"), or
 * undefined. A bare one-segment path ("/bubblebooties") is a channel.
 */
export function profilePath(path: string): string | undefined {
  const segments = path.split("/").filter((segment) => segment !== "");
  if (segments.length >= 2 && PROFILE_KINDS.includes(segments[0])) return `/${segments[0]}/${segments[1]}`;
  if (segments.length === 1 && /^[\w-]+$/.test(segments[0])) return `/channels/${segments[0]}`;
  return undefined;
}

/** Feeds HTML listing pages into the sink until one adds nothing new or the sink is full. */
async function crawlHtml(maxPages: number, sink: ItemSink, urlForPage: (page: number) => string): Promise<void> {
  for (let page = 0; page < maxPages; page++) {
    let html: string;
    try {
      html = await fetchText(urlForPage(page));
    } catch {
      return;
    }
    const cards = parseCardsHtml(html, "xvideos");
    if (cards.length === 0 || sink.add(cards) === 0 || sink.full()) return;
  }
}

/** Feeds a profile's uploads, newest first, from the site's JSON endpoint into the sink. */
async function crawlProfile(profile: string, sink: ItemSink, referer: string): Promise<void> {
  for (let page = 0; page < MAX_PROFILE_PAGES; page++) {
    let answer: any;
    try {
      answer = await fetchJson(`${ORIGINS.xvideos}${profile}/videos/new/${page}`, referer);
    } catch {
      return;
    }
    const items = jsonVideoList(answer?.videos)
      .map((video) => jsonItem(video, "xvideos"))
      .filter((item): item is DiscoveredItem => item !== undefined);
    if (items.length === 0 || sink.add(items) === 0 || sink.full()) return;
  }
}
