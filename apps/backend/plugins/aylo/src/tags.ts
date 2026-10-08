import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { absoluteUrl, fetchDocument, slugFromHref } from "./http.ts";
import { enabledSites, type Site } from "./sites.ts";

/**
 * Lists the sites' categories (the browse-by-category index). A query filters
 * by name. Categories that several sites share by name are listed once, from
 * the first enabled site that has it.
 */
export async function tagList(args: ListArgs): Promise<ItemsResult> {
  const query = args.query ? args.query.trim().toLowerCase() : "";
  const limit = args.limit || 0;
  const offset = args.offset || 0;

  const outcomes = await Promise.allSettled(enabledSites().map((site) => siteCategories(site)));
  const items: DiscoveredItem[] = [];
  const seenNames = new Set<string>();
  let firstFailure: unknown = undefined;
  let succeeded = 0;
  for (const outcome of outcomes) {
    if (outcome.status === "rejected") {
      if (firstFailure === undefined) firstFailure = outcome.reason;
      continue;
    }
    succeeded++;
    for (const item of outcome.value) {
      const key = item.title.toLowerCase();
      if (seenNames.has(key)) continue;
      if (query && !key.includes(query)) continue;
      seenNames.add(key);
      items.push(item);
    }
  }
  if (succeeded === 0 && firstFailure !== undefined) throw firstFailure;

  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

/** Categories of one site, read from its category index page. */
export async function siteCategories(site: Site): Promise<DiscoveredItem[]> {
  const doc = await fetchDocument(site.categoriesUrl);
  const items: DiscoveredItem[] = [];
  const seenHrefs = new Set<string>();

  for (const anchor of doc.querySelectorAll(site.categoryLinkSelector)) {
    const href = anchor.getAttribute("href") || "";
    const name = categoryName(anchor.text);
    if (!href || !name || seenHrefs.has(href)) continue;
    seenHrefs.add(href);
    items.push({
      title: name,
      media_type: "tag",
      source_url: absoluteUrl(site, href),
      external_id: `${site.key}-cat-${categorySlug(href)}`,
    });
  }
  return items;
}

/** The label text of a category link, without any trailing video count. */
function categoryName(text: string): string {
  const lines = text
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
  if (lines.length === 0) return "";
  return lines[0];
}

/** Stable slug of a category link: the `c=` id on Pornhub, the last path segment elsewhere. */
function categorySlug(href: string): string {
  const categoryId = href.match(/[?&]c=(\d+)/);
  if (categoryId) return categoryId[1];
  return slugFromHref(href);
}
