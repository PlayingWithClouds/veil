import type { ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { BASE_URL, PAGE_SIZE, fetchText, toSiteUrl } from "./http.ts";
import { cardOf, tracksFromPage, type RawTrack } from "./track.ts";

const NEWEST_PATH = "/channel/new-audios";

// First path segments of listing pages that are used as they are.
const LISTING_ROOTS = ["tracks", "genre", "channel", "playlist"];

/**
 * Lists audio: a search for a query, a creator, tag, genre, channel or playlist
 * page for a url, otherwise the newest uploads. The site pages by 20, so the
 * pages covering offset..offset+limit are read and trimmed.
 */
export async function audioList(args: ListArgs): Promise<ItemsResult> {
  let limit = PAGE_SIZE;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  let offset = 0;
  if (args.offset !== undefined && args.offset > 0) offset = args.offset;
  const readPage = pageReader(args);

  const skipped = offset % PAGE_SIZE;
  const tracks: RawTrack[] = [];
  let page = Math.floor(offset / PAGE_SIZE) + 1;
  while (tracks.length < skipped + limit) {
    const found = await readPage(page);
    tracks.push(...found);
    if (found.length < PAGE_SIZE) break;
    page++;
  }
  return { items: tracks.slice(skipped, skipped + limit).map(cardOf) };
}

/** The reader of one listing page (1-based) for the arguments. */
function pageReader(args: ListArgs): (page: number) => Promise<RawTrack[]> {
  if (args.query !== undefined && args.query !== "") {
    const query = args.query;
    return (page) => searchPage(query, page);
  }
  const path = listingPath(args.url);
  return (page) => htmlPage(path, page);
}

/** One page of the song search, which answers JSON. */
async function searchPage(query: string, page: number): Promise<RawTrack[]> {
  const params = new URLSearchParams({ q: query, page: String(page) });
  const reply = await fetchText(`${BASE_URL}/api/search/song?${params.toString()}`);
  return tracksFromSearch(reply);
}

/** The tracks of a search reply, {"data": [track, …], "total": …}. */
export function tracksFromSearch(reply: string): RawTrack[] {
  const parsed = JSON.parse(reply) as { data?: RawTrack[] };
  if (!Array.isArray(parsed.data)) return [];
  return parsed.data;
}

/** One page of an HTML listing; a page past the end has no tracks. */
async function htmlPage(path: string, page: number): Promise<RawTrack[]> {
  return tracksFromPage(await fetchText(`${BASE_URL}${path}?page=${page}`));
}

/**
 * The listing path for a page URL. A creator's profile lists its tracks under
 * /<username>/tracks; the other listing pages already are lists.
 */
export function listingPath(url: string | undefined): string {
  if (url === undefined || url === "") return NEWEST_PATH;
  const segments = new URL(toSiteUrl(url)).pathname.split("/").filter((segment) => segment !== "");
  if (segments.length === 0) return NEWEST_PATH;
  if (LISTING_ROOTS.includes(segments[0])) return `/${segments.join("/")}`;
  if (segments.length === 1) return `/${segments[0]}/tracks`;
  if (segments[1] === "tracks") return `/${segments[0]}/tracks`;
  throw new Error(`erocast: cannot list audio from ${url}`);
}
