import { parse } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, gallerySlugFromUrl, resolveUrl } from "./http.ts";

type Root = ReturnType<typeof parse>;
type Element = NonNullable<ReturnType<Root["querySelector"]>>;

/** The gallery stubs of a listing page, in page order, each gallery once. */
export function galleriesFromRoot(root: Root): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll('a[href*="/galls/"]')) {
    const slug = gallerySlugFromUrl(attributeOf(anchor, "href"));
    if (slug === "" || seen.has(slug)) continue;
    seen.add(slug);
    items.push(stubFromAnchor(anchor, slug));
  }
  return items;
}

/** The stub of one gallery card: the model's name as the title, the thumbnail as poster. */
function stubFromAnchor(anchor: Element, slug: string): DiscoveredItem {
  const image = anchor.querySelector("img");
  let title = decodeEntities(attributeOf(anchor, "title"));
  let poster = "";
  if (image !== null) {
    if (title === "") title = decodeEntities(attributeOf(image, "title"));
    if (title === "") title = decodeEntities(attributeOf(image, "alt"));
    poster = attributeOf(image, "src");
  }
  const item: DiscoveredItem = {
    title,
    media_type: "gallery",
    source_url: resolveUrl(`/galls/${slug}`),
    external_id: externalIdFromSlug(slug),
  };
  if (poster !== "") item.poster_path = resolveUrl(poster);
  return item;
}

/** The stable external id of a gallery, from its slug. */
export function externalIdFromSlug(slug: string): string {
  return `hqporn-${slug}`;
}
