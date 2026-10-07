import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Gallery, Image, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { resolveUrl, galleryIdFromUrl, fullSizeImage, decodeEntities } from "./http.ts";

// Scrapes a pornpics gallery page into a Gallery with its full-size images.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const gid = galleryIdFromUrl(target);
  if (!gid) return [];

  const html = await fetchHtml(target);
  const root = parse(html);

  const title = decodeEntities(root.querySelector("h1")?.text?.trim() ?? "");
  const images = extractImages(root, gid);

  const gallery: Gallery = {
    type: "gallery",
    external_id: `pornpics-${gid}`,
    source_url: target,
    title: title || `Gallery ${gid}`,
    cover_path: images[0]?.file_path,
    images,
  };
  return [{ type: "gallery", gallery }];
}

// Gallery images live under a CDN path containing the gallery id; related-gallery
// thumbnails on the same page belong to other ids and are filtered out. Thumbnails
// are rewritten to their full-size (/1280/) variant.
function extractImages(root: ReturnType<typeof parse>, gid: string): Image[] {
  const images: Image[] = [];
  const seen = new Set<string>();

  for (const img of root.querySelectorAll("img[data-src]")) {
    const src = img.getAttribute("data-src") ?? "";
    if (!src.includes("cdni.pornpics.com")) continue;
    if (!src.includes(`/${gid}/`)) continue;

    const full = fullSizeImage(src);
    if (seen.has(full)) continue;
    seen.add(full);

    images.push({
      type: "gallery",
      file_path: full,
      position: images.length,
    });
  }
  return images;
}
