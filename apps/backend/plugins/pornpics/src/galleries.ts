import type { ListArgs, ItemsResult } from "@playingwithclouds/veil-sdk";
import { searchGalleries } from "./api.ts";

const DEFAULT_PAGE_SIZE = 20;

// Evergreen terms used to fill the browse feed — the site has no unfiltered
// listing endpoint. Each page rotates to the next term and advances that term's
// own offset, so paging stays diverse and never repeats the same first page.
const BROWSE_TERMS = [
  "milf",
  "teen",
  "amateur",
  "asian",
  "latina",
  "ebony",
  "blonde",
  "brunette",
  "big tits",
  "anal",
  "mature",
  "redhead",
];

// A query hits the site's JSON search endpoint directly; without one the
// browse feed rotates through BROWSE_TERMS.
export async function galleryList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const pageSize = limit > 0 ? limit : DEFAULT_PAGE_SIZE;

  if (query) {
    const items = await searchGalleries(query, pageSize, offset);
    return { items };
  }

  const page = Math.floor(offset / pageSize);
  const term = BROWSE_TERMS[page % BROWSE_TERMS.length];
  const termOffset = Math.floor(page / BROWSE_TERMS.length) * pageSize;
  const items = await searchGalleries(term, pageSize, termOffset);
  return { items };
}
