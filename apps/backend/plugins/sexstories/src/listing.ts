import { parse } from "node-html-parser";
import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs, StudioRef } from "@playingwithclouds/veil-sdk";
import { postedDate } from "./dates.ts";
import {
  SITE_URL,
  absoluteUrl,
  authorOfProfile,
  externalIdOf,
  fetchPage,
  normalizeText,
  profileIdOf,
  storyIdOf,
} from "./site.ts";

// Items per page the site serves, by listing kind.
const TOPLIST_PAGE_SIZE = 100;
const CATEGORY_PAGE_SIZE = 70;

/** Where a listing's pages live. A source without a page size is one page only. */
export interface ListingSource {
  pageUrl: (page: number) => string;
  pageSize?: number;
  /** An author's page lists a table, not cards. */
  authorPage?: boolean;
}

/** The "Read N times" counter of a card, or undefined. */
function readCountOf(text: string): number | undefined {
  const match = text.match(/Read\s+(\d+)\s+times/);
  if (!match) return undefined;
  return Number(match[1]);
}

/** The rating of a card on the 0..10 scale (the site shows a percentage), or undefined. */
function ratingOf(text: string): number | undefined {
  const match = text.match(/Rated\s+([\d.]+)\s*%/);
  if (!match) return undefined;
  return Math.round(Number(match[1]) * 10) / 100;
}

/** Builds the discovered item of one story link, or undefined when the link is no story. */
function itemOf(link: { text: string; getAttribute: (name: string) => string | undefined }): DiscoveredItem | undefined {
  const href = link.getAttribute("href") ?? "";
  const storyId = storyIdOf(absoluteUrl(href));
  if (storyId === undefined) return undefined;
  return {
    title: normalizeText(link.text),
    media_type: "story",
    source_url: absoluteUrl(href),
    external_id: externalIdOf(storyId),
  };
}

/** The author in a card's byline link ("by <a>"), or undefined when the card has none or it is anonymous. */
function authorOfCard(link: { text: string; getAttribute: (name: string) => string | undefined } | undefined): StudioRef | undefined {
  if (link === undefined) return undefined;
  return authorOfProfile(link.text, link.getAttribute("href") ?? "");
}

/**
 * The owner of an author's page, read from the page's own "texts commented by"
 * link, which holds the profile path; undefined when the page has none.
 */
function authorOfProfilePage(root: HTMLElement): StudioRef | undefined {
  const name = root.querySelector("h3.notice .left")?.text ?? "";
  for (const link of root.querySelectorAll("a.info")) {
    const href = link.getAttribute("href") ?? "";
    if (profileIdOf(href) === undefined) continue;
    return authorOfProfile(name, href.replace(/\/comments$/, ""));
  }
  return undefined;
}

/** Lists the story cards of a toplist, update, category or search page. */
export function parseStoryCards(html: string): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  for (const card of parse(html).querySelectorAll(".stories_list li")) {
    const links = card.querySelectorAll("h4 a");
    if (links.length === 0) continue;
    const item = itemOf(links[0]);
    if (item === undefined) continue;
    const author = authorOfCard(links[1]);
    if (author !== undefined) item.studio = author;

    const text = normalizeText(card.text);
    const date = postedDate(text);
    if (date !== undefined) item.date = date;
    const rating = ratingOf(text);
    if (rating !== undefined) item.rating = rating;
    const viewCount = readCountOf(text);
    if (viewCount !== undefined) item.view_count = viewCount;
    items.push(item);
  }
  return items;
}

/** Lists the stories in the first table of an author's page; its rows carry no date. */
export function parseAuthorPage(html: string): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const root = parse(html);
  const table = root.querySelector("table");
  if (!table) return items;
  const author = authorOfProfilePage(root);
  for (const row of table.querySelectorAll("tbody tr")) {
    const link = row.querySelector("td a");
    if (!link) continue;
    const item = itemOf(link);
    if (item === undefined) continue;
    if (author !== undefined) item.studio = author;

    const cells = row.querySelectorAll("td");
    const viewCount = readCountOf(`Read ${normalizeText(cells[2]?.text ?? "")}`);
    if (viewCount !== undefined) item.view_count = viewCount;
    const rating = ratingOf(`Rated ${normalizeText(cells[3]?.text ?? "")}`);
    if (rating !== undefined) item.rating = rating;
    items.push(item);
  }
  return items;
}

/** The latest stories, 100 to a page. */
function latestSource(): ListingSource {
  return { pageUrl: (page) => `${SITE_URL}/toplists/last/${page}`, pageSize: TOPLIST_PAGE_SIZE };
}

/**
 * Keyword search. The site's own paging breaks after the first page (page 2
 * onwards answers with every remaining match at once), so only page 1 is used.
 */
function searchSource(query: string): ListingSource {
  const words = normalizeText(query).split(" ").map(encodeURIComponent).join("_");
  return { pageUrl: () => `${SITE_URL}/search/1/relevance/${words}////` };
}

/** The listing a site address names (category, toplist, updates, author), or undefined for other addresses. */
export function sourceOfUrl(url: string): ListingSource | undefined {
  const path = url.replace(/^https?:\/\/(?:www\.)?sexstories\.com/i, "").split(/[?#]/)[0];

  const category = path.match(/^\/(themes|genres)\/(\d+)\/([^/]+)/);
  if (category) {
    const base = `${SITE_URL}/${category[1]}/${category[2]}/${category[3]}`;
    return { pageUrl: (page) => `${base}/s-date/p-${page}`, pageSize: CATEGORY_PAGE_SIZE };
  }
  const toplist = path.match(/^\/toplists\/([a-z0-9]+)/);
  if (toplist) {
    return { pageUrl: (page) => `${SITE_URL}/toplists/${toplist[1]}/${page}`, pageSize: TOPLIST_PAGE_SIZE };
  }
  if (path.startsWith("/updates")) {
    return { pageUrl: (page) => `${SITE_URL}/updates/${page}`, pageSize: TOPLIST_PAGE_SIZE };
  }
  if (/^\/profile\d+/.test(path)) {
    return { pageUrl: () => `${SITE_URL}${path}`, authorPage: true };
  }
  return undefined;
}

/** Picks the listing a request names: its url, else its search query, else the latest stories. */
export function sourceOf(args: ListArgs): ListingSource {
  const query = args.query ?? "";
  if (query.trim() !== "") return searchSource(query);
  const fromUrl = sourceOfUrl(args.url ?? "");
  if (fromUrl !== undefined) return fromUrl;
  return latestSource();
}

/** Parses a fetched page according to the kind of listing it came from. */
function parsePage(source: ListingSource, html: string): DiscoveredItem[] {
  if (source.authorPage === true) return parseAuthorPage(html);
  return parseStoryCards(html);
}

/** Reads one page of a single-page source and slices it. */
async function readSinglePage(source: ListingSource, offset: number, limit: number | undefined): Promise<DiscoveredItem[]> {
  const items = parsePage(source, await fetchPage(source.pageUrl(1)));
  if (limit === undefined) return items.slice(offset);
  return items.slice(offset, offset + limit);
}

/** Reads the pages that cover `offset` to `offset + wanted`, stopping early at the last page. */
async function readPages(source: ListingSource, pageSize: number, offset: number, wanted: number): Promise<DiscoveredItem[]> {
  const items: DiscoveredItem[] = [];
  const skip = offset % pageSize;
  let page = Math.floor(offset / pageSize) + 1;
  while (items.length < skip + wanted) {
    const pageItems = parsePage(source, await fetchPage(source.pageUrl(page)));
    items.push(...pageItems);
    if (pageItems.length < pageSize) break;
    page++;
  }
  return items.slice(skip, skip + wanted);
}

/**
 * Lists stories: the latest without a query, a search with one, or the
 * category / toplist / author page named by args.url. limit and offset slice
 * the result; without a limit a paged listing returns one page.
 */
export async function storyList(args: ListArgs): Promise<ItemsResult> {
  const source = sourceOf(args);
  const offset = args.offset ?? 0;
  if (source.pageSize === undefined) {
    return { items: await readSinglePage(source, offset, args.limit) };
  }
  let wanted = source.pageSize;
  if (args.limit !== undefined) wanted = args.limit;
  return { items: await readPages(source, source.pageSize, offset, wanted) };
}
