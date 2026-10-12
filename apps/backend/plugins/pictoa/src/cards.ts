import { parse } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, resolveUrl } from "./http.ts";

type Root = ReturnType<typeof parse>;
type Element = NonNullable<ReturnType<Root["querySelector"]>>;

// An album page path ends in "<slug>-<albumId>.html"; the id is the trailing number.
const ALBUM_PATH_PATTERN = /\/albums\/([^/]+?)-(\d+)\.html$/;

/** The album id at the end of an album URL, or undefined when the URL is not an album. */
export function albumIdFromUrl(url: string): string | undefined {
  const match = new URL(resolveUrl(url)).pathname.match(ALBUM_PATH_PATTERN);
  if (match === null) return undefined;
  return match[2];
}

/** The stable external id of an album. */
export function externalIdFromAlbumId(albumId: string): string {
  return `pictoa-${albumId}`;
}

/** The album stubs of a listing page (home, a numbered page or a search), in page order. */
export function albumsFromRoot(root: Root): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const titleLink of root.querySelectorAll("a.gallery-title")) {
    const item = stubFromTitleLink(titleLink);
    if (item === undefined || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/** The stub of one album card, with the cover from the card's thumbnail. */
function stubFromTitleLink(titleLink: Element): DiscoveredItem | undefined {
  const sourceUrl = resolveUrl(attributeOf(titleLink, "href"));
  const albumId = albumIdFromUrl(sourceUrl);
  if (albumId === undefined) return undefined;

  const item: DiscoveredItem = {
    title: decodeEntities(titleLink.text.trim()),
    media_type: "gallery",
    source_url: sourceUrl,
    external_id: externalIdFromAlbumId(albumId),
  };
  const cover = coverOfCard(titleLink);
  if (cover !== "") item.poster_path = cover;
  return item;
}

/** The thumbnail URL of the album card that holds a title link, or "" when the card has none. */
function coverOfCard(titleLink: Element): string {
  const card = titleLink.parentNode;
  if (card === null) return "";
  const image = card.querySelector("img");
  if (image === null) return "";
  return resolveUrl(attributeOf(image, "data-lazy-src"));
}
