import type { ListArgs, ItemsResult, DiscoveredItem, ScrapeResult, Gallery, Image } from "@playingwithclouds/veil-sdk";
import { SITE } from "./http.ts";
import { fetchDoc, absoluteUrl, imageSrc } from "./html.ts";

const MAX_LIST_PAGES = 10;
const DEFAULT_TARGET = 30;

// eporner photo galleries live at /gallery/<id>/<slug>/.
function idFromUrl(url: string): string {
  const match = url.match(/\/gallery\/([^/]+)/);
  return match ? match[1] : "";
}

function externalId(url: string): string {
  return `eporner-gal-${idFromUrl(url)}`;
}

// Lists photo galleries from the /pics/ index, paginating until the target is
// met. A query filters crawled titles (eporner has no photo search API).
export async function galleryList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim().toLowerCase();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : DEFAULT_TARGET;

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= MAX_LIST_PAGES; page++) {
    const path = page === 1 ? `${SITE}/pics/` : `${SITE}/pics/${page}/`;
    const doc = await fetchDoc(path);
    if (!doc) break;

    const blocks = doc.querySelectorAll(".mbphoto2");
    if (blocks.length === 0) break;

    for (const block of blocks) {
      const anchor = block.querySelector("a[href*='/gallery/']");
      const href = anchor?.getAttribute("href");
      if (!href) continue;
      const title = block.querySelector(".mbtitphoto2")?.text?.trim() ?? "";
      if (query && !title.toLowerCase().includes(query)) continue;

      const sourceUrl = absoluteUrl(href);
      if (seen.has(sourceUrl)) continue;
      seen.add(sourceUrl);

      items.push({
        title: title || "Gallery",
        media_type: "gallery",
        source_url: sourceUrl,
        external_id: externalId(sourceUrl),
        poster_path: imageSrc(anchor?.querySelector("img")),
        count: parseCount(block.querySelector(".mbphoto2_galleryico")?.text),
      });
    }
    if (items.length >= target) break;
  }

  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

// Fetches one gallery page as a Gallery with its photos.
export async function galleryFind(url: string): Promise<ScrapeResult[]> {
  const id = idFromUrl(url);
  if (!id) throw new Error(`eporner gallery: no id in ${url}`);

  const doc = await fetchDoc(url);
  if (!doc) throw new Error(`eporner gallery: no page for ${url}`);

  const title = doc.querySelector("h1")?.text?.trim() || "Gallery";
  const images: Image[] = [];
  let position = 0;
  for (const block of doc.querySelectorAll(".mbphoto2")) {
    const src = imageSrc(block.querySelector("img"));
    // Only the gallery's own photos carry its id in the CDN path.
    if (!src || !src.includes(id)) continue;
    images.push({ type: "gallery", file_path: src, position });
    position++;
  }

  const gallery: Gallery = {
    type: "gallery",
    external_id: externalId(url),
    source_url: url,
    title,
    images,
  };
  if (images.length > 0) gallery.cover_path = images[0].file_path;
  return [{ type: "gallery", gallery }];
}

function parseCount(text: string | undefined): number | undefined {
  if (!text) return undefined;
  const match = text.match(/\d+/);
  if (!match) return undefined;
  return parseInt(match[0], 10);
}
