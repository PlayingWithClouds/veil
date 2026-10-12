import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Gallery, Image, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { albumIdFromUrl, externalIdFromAlbumId } from "./cards.ts";
import { attributeOf, BASE_URL, decodeEntities, resolveUrl } from "./http.ts";

type Root = ReturnType<typeof parse>;

// An album's image pages are /albums/<slug>-<albumId>/<imageId>.html.
const IMAGE_PAGE_PATTERN = /\/albums\/[^/]+\/\d+\.html$/;
// Album pages past the first are <slug>-<albumId>-p<N>.html.
const PAGE_SUFFIX_PATTERN = /-p(\d+)\.html$/;

/** Scrapes an album (every page of it) into a Gallery with its full-size images. */
export async function scrapeGallery(url: string): Promise<ScrapeResult[]> {
  const albumId = albumIdFromUrl(url);
  if (albumId === undefined) return [];
  const albumUrl = canonicalAlbumUrl(url);

  const firstPage = parse(await fetchHtml(albumUrl));
  const images = await collectImages(albumUrl, firstPage);
  if (images.length === 0) return [];

  const gallery: Gallery = {
    type: "gallery",
    external_id: externalIdFromAlbumId(albumId),
    source_url: albumUrl,
    title: titleOf(firstPage, albumId),
    cover_path: images[0].file_path,
    images,
  };
  return [{ type: "gallery", gallery }];
}

/** The album's heading, or a placeholder naming the album id when the page has none. */
export function titleOf(page: Root, albumId: string): string {
  const heading = page.querySelector("h1");
  if (heading === null) return `Album ${albumId}`;
  const title = decodeEntities(heading.text.trim());
  if (title === "") return `Album ${albumId}`;
  return title;
}

/** The album's own page URL, without any page suffix. */
export function canonicalAlbumUrl(url: string): string {
  const pathname = new URL(resolveUrl(url)).pathname.replace(PAGE_SUFFIX_PATTERN, ".html");
  return `${BASE_URL}${pathname}`;
}

/** The full-size images of the album on its first page and on every later page it links to. */
async function collectImages(albumUrl: string, firstPage: Root): Promise<Image[]> {
  const images: Image[] = [];
  const seen = new Set<string>();
  const pages = [firstPage];
  for (let pageNumber = 2; pageNumber <= pageCount(firstPage); pageNumber++) {
    pages.push(parse(await fetchHtml(albumPageUrl(albumUrl, pageNumber))));
  }
  for (const page of pages) {
    for (const url of imageUrlsOnPage(page)) {
      if (seen.has(url)) continue;
      seen.add(url);
      images.push({ type: "gallery", file_path: url, position: images.length });
    }
  }
  return images;
}

/** The number of album pages, read from the pagination links (1 when there are none). */
export function pageCount(page: Root): number {
  let highest = 1;
  for (const link of page.querySelectorAll("ul.pagination a")) {
    const match = attributeOf(link, "href").match(PAGE_SUFFIX_PATTERN);
    if (match === null) continue;
    highest = Math.max(highest, Number(match[1]));
  }
  return highest;
}

/** The URL of one page of an album: <slug>-<albumId>-p<N>.html. */
export function albumPageUrl(albumUrl: string, pageNumber: number): string {
  return albumUrl.replace(/\.html$/, `-p${pageNumber}.html`);
}

/**
 * The full-size image URLs of the image cards on one album page. The thumbnail
 * lives on t1.pictoa.com and the same file at full size on s2.pictoa.com.
 */
export function imageUrlsOnPage(page: Root): string[] {
  const urls: string[] = [];
  for (const link of page.querySelectorAll("a.gallery-link")) {
    if (!IMAGE_PAGE_PATTERN.test(new URL(resolveUrl(attributeOf(link, "href"))).pathname)) continue;
    const image = link.querySelector("img");
    if (image === null) continue;
    const thumbnail = attributeOf(image, "data-lazy-src");
    if (thumbnail === "") continue;
    urls.push(thumbnail.replace("//t1.pictoa.com/", "//s2.pictoa.com/"));
  }
  return urls;
}
