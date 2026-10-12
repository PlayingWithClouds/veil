import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Gallery, Image, ScenePerformer, ScrapeResult, StudioRef } from "@playingwithclouds/veil-sdk";
import { attributeOf, BASE_URL, decodeEntities, gallerySlugFromUrl, resolveUrl } from "./http.ts";
import { externalIdFromSlug } from "./cards.ts";

type Root = ReturnType<typeof parse>;

// The page heading reads "Gallery of Mia Malkova from Hard X".
const HEADING_PATTERN = /^Gallery of (.+) from (.+)$/;
// Fancybox anchors also include placeholders that open nothing; only image files count.
const IMAGE_FILE_PATTERN = /\.(?:jpe?g|png|webp|gif)$/i;

/** Scrapes one gallery page into a Gallery with its full-size images, model and site. */
export async function scrapeGallery(url: string): Promise<ScrapeResult[]> {
  const slug = gallerySlugFromUrl(url);
  if (slug === "") return [];
  const pageUrl = `${BASE_URL}/galls/${slug}`;
  return galleryFromPage(pageUrl, await fetchHtml(pageUrl));
}

/** Builds the scrape result of a gallery page's HTML. */
export function galleryFromPage(sourceUrl: string, html: string): ScrapeResult[] {
  const slug = gallerySlugFromUrl(sourceUrl);
  if (slug === "") return [];
  const root = parse(html);

  const heading = findGalleryHeading(root);
  let modelLink: Root | null = null;
  let siteLink: Root | null = null;
  if (heading !== undefined) {
    modelLink = heading.querySelector('a[href^="/model/"]');
    siteLink = heading.querySelector('a[href^="/site/"]');
  }

  const images = imagesFromRoot(root);
  const gallery: Gallery = {
    type: "gallery",
    external_id: externalIdFromSlug(slug),
    source_url: resolveUrl(`/galls/${slug}`),
    title: titleFrom(root, modelLink),
    images,
  };
  if (images.length > 0) gallery.cover_path = images[0].file_path;
  if (modelLink !== null) gallery.performers = [performerFromLink(modelLink)];
  if (siteLink !== null) gallery.studio = studioFromLink(siteLink);
  return [{ type: "gallery", gallery }];
}

/** The page's "Gallery of <model> from <site>" heading, if it has one. */
function findGalleryHeading(root: Root): Root | undefined {
  for (const heading of root.querySelectorAll("h2")) {
    if (HEADING_PATTERN.test(decodeEntities(heading.text.trim()))) return heading;
  }
  return undefined;
}

/** The full-size images of a gallery, from the fancybox links, in page order. */
function imagesFromRoot(root: Root): Image[] {
  const images: Image[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll("a[data-fancybox]")) {
    const url = resolveUrl(attributeOf(anchor, "href"));
    if (!IMAGE_FILE_PATTERN.test(url)) continue;
    if (seen.has(url)) continue;
    seen.add(url);
    images.push({ type: "gallery", file_path: url, position: images.length });
  }
  return images;
}

/** The title: the model's name when the page links one, else the page's <title> without the site name. */
function titleFrom(root: Root, modelLink: Root | null): string {
  if (modelLink !== null) return decodeEntities(modelLink.text.trim());
  const pageTitle = root.querySelector("title");
  if (pageTitle === null) return "";
  return decodeEntities(pageTitle.text.trim()).split(" | ")[0];
}

/** The performer credited by a /model/<name> link. */
function performerFromLink(link: NonNullable<ReturnType<Root["querySelector"]>>): ScenePerformer {
  return {
    name: decodeEntities(link.text.trim()),
    source_url: resolveUrl(attributeOf(link, "href")),
  };
}

/** The studio (the site the set came from) of a /site/<name> link. */
function studioFromLink(link: NonNullable<ReturnType<Root["querySelector"]>>): StudioRef {
  const sourceUrl = resolveUrl(attributeOf(link, "href"));
  const segments = new URL(sourceUrl).pathname.split("/").filter((segment) => segment !== "");
  return {
    name: decodeEntities(link.text.trim()),
    external_id: `hqporn-site-${segments[1]}`,
    source_url: sourceUrl,
  };
}
