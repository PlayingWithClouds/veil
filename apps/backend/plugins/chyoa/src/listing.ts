import { parse } from "node-html-parser";
import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { SITE_URL, absoluteUrl, authorOfLink, externalIdOf, fetchPublicPage, normalizeText, storyIdOf } from "./site.ts";

// Every listing (new stories, search, categories) shows 20 stories per page.
const PAGE_SIZE = 20;

/** Where a listing's pages live. */
export interface ListingSource {
  pageUrl: (page: number) => string;
}

/** The view counter of a card ("1,204 Views"), or undefined. */
function viewCountOf(text: string): number | undefined {
  const match = text.match(/([\d,]+)\s+Views/);
  if (!match) return undefined;
  return Number(match[1].replace(/,/g, ""));
}

/** Builds the discovered item of one story card, or undefined when the card links to no story. */
function itemOf(card: HTMLElement): DiscoveredItem | undefined {
  const link = card.querySelector("h3 a");
  if (!link) return undefined;
  const sourceUrl = absoluteUrl(link.getAttribute("href") ?? "");
  const storyId = storyIdOf(sourceUrl);
  if (storyId === undefined) return undefined;

  const item: DiscoveredItem = {
    title: normalizeText(link.text),
    media_type: "story",
    source_url: sourceUrl,
    external_id: externalIdOf(storyId),
  };
  const byline = card.querySelector('.redesign-search-story__byline a[href*="/user/"]');
  if (byline !== null) {
    const author = authorOfLink(byline);
    if (author !== undefined) item.studio = author;
  }
  const cover = card.querySelector("img")?.getAttribute("src");
  if (cover !== undefined && cover.includes("/data/covers/")) item.poster_path = cover;
  const updated = card.querySelector("[data-livestamp]")?.getAttribute("data-livestamp");
  if (updated !== undefined) item.date = updated.slice(0, 10);
  const viewCount = viewCountOf(normalizeText(card.text));
  if (viewCount !== undefined) item.view_count = viewCount;
  return item;
}

/** Lists the story cards of a listing, category or search page. */
export function parseStoryCards(html: string): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  for (const card of parse(html).querySelectorAll("article.redesign-search-story")) {
    const item = itemOf(card);
    if (item !== undefined) items.push(item);
  }
  return items;
}

/** The newest stories. */
function latestSource(): ListingSource {
  return { pageUrl: (page) => `${SITE_URL}/new-sex-stories/page/${page}` };
}

/** Keyword search over stories. */
function searchSource(query: string): ListingSource {
  const encoded = encodeURIComponent(normalizeText(query));
  return { pageUrl: (page) => `${SITE_URL}/search?q=${encoded}&type=story&page=${page}` };
}

/** A category page: /category/<slug>?page=<n>. */
function categorySource(slug: string): ListingSource {
  return { pageUrl: (page) => `${SITE_URL}/category/${slug}?page=${page}` };
}

/** A ranking page such as /most-liked-sex-stories: /<name>/page/<n>. */
function rankingSource(name: string): ListingSource {
  return { pageUrl: (page) => `${SITE_URL}/${name}/page/${page}` };
}

/** The listing a site address names (category, ranking, search), or undefined for other addresses. */
export function sourceOfUrl(url: string): ListingSource | undefined {
  const parsed = new URL(url, SITE_URL);
  const category = parsed.pathname.match(/^\/category\/([^/]+)/);
  if (category) return categorySource(category[1]);
  const ranking = parsed.pathname.match(/^\/([a-z-]+-sex-stories|adult-text-games)(?:\/|$)/);
  if (ranking) return rankingSource(ranking[1]);
  const query = parsed.searchParams.get("q");
  if (parsed.pathname === "/search" && query !== null) return searchSource(query);
  return undefined;
}

/** Picks the listing a request names: its search query, else its url, else the newest stories. */
export function sourceOf(args: ListArgs): ListingSource {
  const query = args.query ?? "";
  if (query.trim() !== "") return searchSource(query);
  const fromUrl = sourceOfUrl(args.url ?? "");
  if (fromUrl !== undefined) return fromUrl;
  return latestSource();
}

/** Reads the pages that cover `offset` to `offset + wanted`, stopping early at the last page. */
async function readPages(source: ListingSource, offset: number, wanted: number): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  const skip = offset % PAGE_SIZE;
  let page = Math.floor(offset / PAGE_SIZE) + 1;
  while (items.length < skip + wanted) {
    const pageItems = parseStoryCards(await fetchPublicPage(source.pageUrl(page)));
    items.push(...pageItems);
    if (pageItems.length < PAGE_SIZE) break;
    page++;
  }
  return items.slice(skip, skip + wanted);
}

/**
 * Lists stories: the newest without a query, a search with one, or the
 * category or ranking page named by args.url. Without a limit one page is
 * returned.
 */
export async function storyList(args: ListArgs): Promise<ItemsResult> {
  let wanted = PAGE_SIZE;
  if (args.limit !== undefined) wanted = args.limit;
  let offset = 0;
  if (args.offset !== undefined) offset = args.offset;
  return { items: await readPages(sourceOf(args), offset, wanted) };
}
