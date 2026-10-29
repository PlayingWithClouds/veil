import { parse, type HTMLElement } from "node-html-parser";
import type { ForumPostHit, ForumSearchArgs, ForumSearchResult } from "@playingwithclouds/veil-sdk";
import { fetchPage } from "./http.ts";
import { SITE_ORIGIN } from "./parse.ts";

/** Hits returned when the caller names no limit; one page of the site holds about 30. */
const DEFAULT_LIMIT = 24;

/** Searches the site's threads by text: their titles, questions and answers. */
export async function searchPosts(args: ForumSearchArgs): Promise<ForumSearchResult> {
  const html = await fetchPage(`${SITE_ORIGIN}/search.html?${new URLSearchParams({ q: args.query })}`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { posts: parseSearchHits(html).slice(0, limit) };
}

/**
 * Reads the threads of a text search page, best first. The page also lists
 * tags and pornstars matching the query; only threads are kept.
 */
export function parseSearchHits(html: string): ForumPostHit[] {
  const hits: ForumPostHit[] = [];
  for (const result of parse(html).querySelectorAll("#nsw_r a.nsw_r_w")) {
    const hit = searchHit(result);
    if (hit) hits.push(hit);
  }
  return hits;
}

/** One search result as a hit; undefined for anything but a thread. */
function searchHit(result: HTMLElement): ForumPostHit | undefined {
  const path = result.getAttribute("href");
  if (!path || !path.startsWith("/post/")) return undefined;
  const hit: ForumPostHit = {
    title: result.querySelector(".nsw_r_tit")?.text.trim() || "Untitled post",
    url: SITE_ORIGIN + path,
    solved: result.querySelector(".nsw_r_slvd") !== null,
  };
  const thumbnail = squareThumbnail(result.querySelector("img.nsw_r_thmb")?.getAttribute("data-dyn"));
  if (thumbnail) hit.thumbnail_url = thumbnail;
  const date = result.querySelector(".nsw_r_det span")?.text.trim();
  if (date && /^\d{4}-\d{2}-\d{2}$/.test(date)) hit.date = `${date}T00:00:00.000Z`;
  const author = result.querySelector(".nsw_r_athr")?.text.replace(/^Asked by\s*/i, "").trim();
  if (author) hit.author = author;
  const snippet = result.querySelector(".nsw_r_desc")?.text.replace(/\s+/g, " ").trim();
  if (snippet) hit.snippet = snippet;
  return hit;
}

/**
 * The thumbnail at a size worth showing: the search lazy-loads 102 px "tinies"
 * (data-dyn, without extension); the same picture is kept as a larger square.
 */
function squareThumbnail(dynamicSource: string | undefined): string | undefined {
  if (!dynamicSource || !dynamicSource.startsWith(`${SITE_ORIGIN}/media/`)) return undefined;
  return `${dynamicSource.replace("/media/tinies/", "/media/squares/")}.jpg`;
}
