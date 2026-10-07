import { parse } from "node-html-parser";
import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { BASE_URL, resolveUrl } from "./http.ts";

// The site has no tag search endpoint, so a query filters the category list
// locally.

export async function tagList(args: ListArgs): Promise<ItemsResult> {
  return { items: selectEntries(await fetchCategories(), args) };
}

// Scrapes the PornPics landing page, whose tiles are the category index: each
// carries a name, its category URL, and a preview thumbnail.
async function fetchCategories(): Promise<DiscoveredItem[]> {
  const response = await fetch(BASE_URL + "/", { headers: BROWSER_HEADERS });
  if (!response.ok) {
    throw new Error(`pornpics categories ${response.status}`);
  }
  const root = parse(await response.text());

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll("li.thumbwook a.rel-link")) {
    const href = anchor.getAttribute("href") ?? "";
    // Category links are a single path segment (/milf/); gallery links contain
    // /galleries/ and are skipped.
    if (!/^\/[a-z0-9-]+\/$/.test(href)) continue;

    const name = anchor.querySelector("span.h2")?.text?.trim();
    if (!name) continue;

    const slug = href.replace(/\//g, "");
    if (seen.has(slug)) continue;
    seen.add(slug);

    const item: DiscoveredItem = {
      title: name,
      media_type: "tag",
      source_url: resolveUrl(href),
      external_id: `pornpics-category-${slug}`,
    };
    const poster = anchor.querySelector("img")?.getAttribute("data-src");
    if (poster) item.poster_path = poster;
    items.push(item);
  }
  return items;
}

function selectEntries(items: DiscoveredItem[], args: ListArgs): DiscoveredItem[] {
  const query = (args.query ?? "").trim().toLowerCase();
  let matches = items;
  if (query) {
    matches = items.filter((item) => item.title.toLowerCase().includes(query));
  }

  const offset = args.offset ?? 0;
  const limit = args.limit ?? 0;
  if (limit > 0) return matches.slice(offset, offset + limit);
  if (offset > 0) return matches.slice(offset);
  return matches;
}
