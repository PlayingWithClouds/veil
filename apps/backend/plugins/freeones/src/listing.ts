import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { BASE_URL, attributeOf, externalIdFor, fetchPage, profileUrl, slugFromUrl } from "./http.ts";
import { normalizeText } from "./values.ts";

const DEFAULT_PAGE_SIZE = 24;
// The largest page the site serves in one request (its l= parameter).
const MAX_PAGE_SIZE = 100;

/**
 * Lists performers: a keyword search when args.query is set, otherwise the
 * site's performers A to Z. The offset maps onto the site's 1-based pages.
 */
export async function performerList(args: ListArgs): Promise<ItemsResult> {
  const pageSize = pageSizeFrom(args.limit);
  const offset = offsetFrom(args.offset);
  const pageNumber = Math.floor(offset / pageSize) + 1;
  return { items: performerItemsFromPage(parse(await fetchPage(listingUrl(args.query, pageNumber, pageSize)))) };
}

/** The number of performers per request: the limit, bounded to what the site serves. */
export function pageSizeFrom(limit: number | undefined): number {
  if (limit === undefined || limit < 1) return DEFAULT_PAGE_SIZE;
  return Math.min(limit, MAX_PAGE_SIZE);
}

/** The offset into the listing, zero when the request has none. */
function offsetFrom(offset: number | undefined): number {
  if (offset === undefined || offset < 0) return 0;
  return offset;
}

/**
 * The performers listing URL: a name search when query is set, otherwise the
 * performers A to Z. The name order is stable across pages; the site's rank
 * order reshuffles ties between requests, so offsets would repeat performers.
 */
export function listingUrl(query: string | undefined, pageNumber: number, pageSize: number): string {
  const parameters = new URLSearchParams();
  if (query === undefined) {
    parameters.set("s", "name");
    parameters.set("o", "asc");
  } else {
    parameters.set("q", query);
  }
  parameters.set("l", String(pageSize));
  parameters.set("p", String(pageNumber));
  return `${BASE_URL}/performers?${parameters.toString()}`;
}

/** The performer cards of a listing page, one DiscoveredItem per card. */
export function performerItemsFromPage(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  for (const card of root.querySelectorAll('[data-test="teaser-subject"]')) {
    const item = performerItemFromCard(card);
    if (item !== undefined) items.push(item);
  }
  return items;
}

/**
 * One performer card. Its name is the subject-name element (portrait-less
 * cards carry it too), its slug the profile link, its portrait the image when present.
 */
function performerItemFromCard(card: HTMLElement): DiscoveredItem | undefined {
  const link = card.querySelector("a.teaser__link");
  const nameElement = card.querySelector('[data-test="subject-name"]');
  if (link === null || nameElement === null) return undefined;
  const slug = slugFromUrl(`${BASE_URL}${attributeOf(link, "href")}`);
  const name = normalizeText(nameElement.text);
  if (slug === "" || name === "") return undefined;
  const item: DiscoveredItem = {
    title: name,
    media_type: "performer",
    source_url: profileUrl(slug),
    external_id: externalIdFor(slug),
  };
  const image = card.querySelector("img");
  if (image !== null) item.poster_path = attributeOf(image, "src");
  return item;
}
