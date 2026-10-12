import type { ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { API_URL, fetchJson } from "./http.ts";
import { stubFromPost, type WordPressPost } from "./posts.ts";

const DEFAULT_PAGE_SIZE = 20;
// WordPress caps per_page at 100.
const MAX_PAGE_SIZE = 100;

/** Gallery stubs from the site's post list, newest first; a query runs the site's own search. */
export async function galleryList(args: ListArgs): Promise<ItemsResult> {
  const pageSize = pageSizeFrom(args.limit);
  let offset = 0;
  if (args.offset !== undefined) offset = args.offset;
  const parameters = new URLSearchParams({
    per_page: String(pageSize),
    offset: String(offset),
    _fields: "id,slug,title,content",
  });
  if (args.query !== undefined && args.query.trim() !== "") {
    parameters.set("search", args.query.trim());
  }
  const posts = await fetchJson<WordPressPost[]>(`${API_URL}/posts?${parameters}`);
  return { items: posts.map(stubFromPost) };
}

/** The page size asked for, within what the site serves. */
export function pageSizeFrom(limit: number | undefined): number {
  if (limit === undefined || limit <= 0) return DEFAULT_PAGE_SIZE;
  return Math.min(limit, MAX_PAGE_SIZE);
}
