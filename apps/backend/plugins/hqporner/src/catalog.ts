import { parse, HTMLElement } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import { BASE_URL, resolveUrl, actressSlugFromUrl, categorySlugFromUrl } from "./http.ts";
import type { ListArgs, ItemsResult, DiscoveredItem } from "@playingwithclouds/veil-sdk";

// The site has no performer/tag search endpoint, so a query filters the full
// catalog list locally.

// /girls lists the featured pornstar cards plus a sidebar of all actresses.
export async function performerList(args: ListArgs): Promise<ItemsResult> {
  return { items: selectEntries(await fetchPerformers(), args) };
}

// /categories lists every genre as a card grid plus a sidebar link list.
export async function tagList(args: ListArgs): Promise<ItemsResult> {
  return { items: selectEntries(await fetchCategories(), args) };
}

async function fetchPerformers(): Promise<DiscoveredItem[]> {
  const root = parse(await fetchHtml(`${BASE_URL}/girls`));
  return collectEntries(root, "/actress/", (href) => {
    const slug = actressSlugFromUrl(href);
    if (!slug) return undefined;
    return {
      media_type: "performer",
      external_id: `hqporner-actress-${slug}`,
      source_url: `${BASE_URL}/actress/${slug}`,
    };
  });
}

async function fetchCategories(): Promise<DiscoveredItem[]> {
  const root = parse(await fetchHtml(`${BASE_URL}/categories`));
  return collectEntries(root, "/category/", (href) => {
    const slug = categorySlugFromUrl(href);
    if (!slug) return undefined;
    return {
      media_type: "tag",
      external_id: `hqporner-category-${slug}`,
      source_url: `${BASE_URL}/category/${slug}`,
    };
  });
}

interface EntryIdentity {
  media_type: DiscoveredItem["media_type"];
  external_id: string;
  source_url: string;
}

// Both catalog pages share one layout: card sections with an image anchor plus
// plain sidebar links. The same target can appear in both; the card's poster is
// backfilled onto the entry regardless of which anchor was seen first.
function collectEntries(
  root: HTMLElement,
  hrefMarker: string,
  identify: (href: string) => EntryIdentity | undefined
): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const byId = new Map<string, DiscoveredItem>();

  for (const anchor of root.querySelectorAll(`a[href*="${hrefMarker}"]`)) {
    const href = anchor.getAttribute("href") ?? "";
    const identity = identify(href);
    if (!identity) continue;

    const poster = entryPoster(anchor);
    const existing = byId.get(identity.external_id);
    if (existing) {
      if (poster && !existing.poster_path) existing.poster_path = poster;
      continue;
    }

    const title = entryTitle(anchor);
    if (!title) continue;

    const item: DiscoveredItem = { title, ...identity };
    if (poster) item.poster_path = poster;
    byId.set(identity.external_id, item);
    items.push(item);
  }

  return items;
}

// Image anchors carry the name in the img alt; text anchors carry it inline.
function entryTitle(anchor: HTMLElement): string {
  const text = anchor.text.trim();
  if (text) return text;
  return anchor.querySelector("img")?.getAttribute("alt")?.trim() ?? "";
}

function entryPoster(anchor: HTMLElement): string | undefined {
  const card = anchor.closest("section");
  const src = card?.querySelector("a.image img")?.getAttribute("src");
  if (!src) return undefined;
  return resolveUrl(src);
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
