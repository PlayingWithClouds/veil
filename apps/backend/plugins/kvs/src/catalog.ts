import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs, MediaType } from "@playingwithclouds/veil-sdk";
import { enabledSites, type SiteConfig } from "./sites.ts";
import { crawlListing } from "./crawl.ts";
import { listAcrossSites } from "./fanout.ts";
import { imageSource } from "./listing.ts";
import { cleanText, entityName } from "./text.ts";
import { entityExternalId, entitySlug, entityUrl } from "./urls.ts";

const MAX_INDEX_PAGES = 50;
// KVS has no model/tag search endpoint, so a query filters this many index pages.
const MAX_FILTERED_PAGES = 5;

/**
 * Page chrome whose links are sort orders ("/models/most-viewed/") or
 * shortcuts, not catalog entries: menus, list headlines with their sort
 * switches, pagers.
 */
const NAVIGATION_SELECTOR = [
  "nav",
  "header",
  "footer",
  '[class*="menu"]',
  '[class*="headline"]',
  '[class*="head_line"]',
  '[class~="sort"]',
  '[class*="sorting"]',
  '[class*="sort_by"]',
  '[id$="_sort_list"]',
  '[class*="pagination"]',
].join(", ");

/** One kind of catalog: which index to read and which links on it are entries. */
interface Catalog {
  mediaType: MediaType;
  /** external id segment: kvs-<site>-<kind>-<slug> */
  kind: string;
  indexPath: (site: SiteConfig) => string | undefined;
  prefixes: (site: SiteConfig) => string[];
}

export const TAGS: Catalog = {
  mediaType: "tag",
  kind: "tag",
  indexPath: (site) => site.tagIndexPath,
  prefixes: (site) => site.tagPrefixes,
};

export const MODELS: Catalog = {
  mediaType: "performer",
  kind: "model",
  indexPath: (site) => site.modelIndexPath,
  prefixes: (site) => site.modelPrefixes,
};

/** tag:list — the categories/tags of every enabled site that has an index. */
export function tagList(args: ListArgs): Promise<ItemsResult> {
  return listCatalog(TAGS, args);
}

/** performer:list — the models/pornstars of every enabled site that has an index. */
export function performerList(args: ListArgs): Promise<ItemsResult> {
  return listCatalog(MODELS, args);
}

/** Lists a catalog across sites; a query filters entries by name. */
async function listCatalog(catalog: Catalog, args: ListArgs): Promise<ItemsResult> {
  const sites = enabledSites().filter((site) => catalog.indexPath(site));
  if (sites.length === 0) return { items: [] };
  const query = (args.query ?? "").trim().toLowerCase();
  const parseEntries = (site: SiteConfig, root: HTMLElement) => parseCatalogEntries(catalog, site, root);

  return listAcrossSites(sites, args, async (site, target) => {
    const indexUrl = site.origin + (catalog.indexPath(site) as string);
    if (!query) return crawlListing(site, indexUrl, target, MAX_INDEX_PAGES, parseEntries);
    const entries = await crawlListing(site, indexUrl, Number.MAX_SAFE_INTEGER, MAX_FILTERED_PAGES, parseEntries);
    return entries.filter((entry) => entry.title.toLowerCase().includes(query));
  });
}

/** The catalog entries linked from an index page, one per slug, in page order. */
export function parseCatalogEntries(catalog: Catalog, site: SiteConfig, root: HTMLElement): DiscoveredItem[] {
  const entries: DiscoveredItem[] = [];
  const bySlug = new Map<string, DiscoveredItem>();

  for (const prefix of catalog.prefixes(site)) {
    for (const anchor of root.querySelectorAll("a[href]")) {
      const slug = entitySlug(site, anchor.getAttribute("href") ?? "", [prefix]);
      if (!slug || anchor.closest(NAVIGATION_SELECTOR)) continue;
      const poster = entryPoster(site, anchor);
      const known = bySlug.get(slug);
      if (known) {
        if (poster && !known.poster_path) known.poster_path = poster;
        continue;
      }
      const [title, inlineCount] = splitCount(entryTitle(anchor));
      if (!title) continue;

      const entry: DiscoveredItem = {
        title,
        media_type: catalog.mediaType,
        source_url: entityUrl(site, prefix, slug),
        external_id: entityExternalId(site, catalog.kind, slug),
      };
      if (poster) entry.poster_path = poster;
      const count = entryCount(anchor) || inlineCount;
      if (count > 0) entry.count = count;
      bySlug.set(slug, entry);
      entries.push(entry);
    }
  }
  return entries;
}

/** Splits "Blonde (1,234)" into its name and count; count 0 when the name carries none. */
export function splitCount(text: string): [string, number] {
  const match = text.match(/^(.*\S)\s*\((\d[\d,.]*)\)$/);
  if (!match) return [text, 0];
  return [match[1], parseInt(match[2].replace(/[,.]/g, ""), 10)];
}

/** An index entry's name: a title/name element, the image's alt, else the link's own name. */
function entryTitle(anchor: HTMLElement): string {
  const named = anchor.querySelector('[class*="title"], [class*="name"]');
  if (named) {
    const text = cleanText(named.text);
    if (text) return text;
  }
  const alt = cleanText(anchor.querySelector("img[alt]")?.getAttribute("alt") ?? "");
  if (alt) return alt;
  return entityName(anchor);
}

/** The entry's thumbnail inside its link, if any. */
function entryPoster(site: SiteConfig, anchor: HTMLElement): string {
  const image = anchor.querySelector("img");
  if (!image) return "";
  return imageSource(site, image);
}

/** The video count an index entry shows ("1,348 videos"), 0 when none. */
function entryCount(anchor: HTMLElement): number {
  const counter = anchor.querySelector('[class*="count"], [class*="videos"], [class*="total"]');
  if (!counter) return 0;
  // Abbreviated counts ("1.2K") would read as the wrong number.
  if (/\d\s*[kKmM]\b/.test(counter.text)) return 0;
  const digits = counter.text.replace(/[^\d]/g, "");
  if (!digits) return 0;
  return parseInt(digits, 10);
}
