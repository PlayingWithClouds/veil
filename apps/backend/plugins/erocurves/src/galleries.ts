import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import { API_URL, BASE_URL, fetchJson } from "./http.ts";
import { galleryFromPost, type WordPressPost } from "./posts.ts";

/** Scrapes one gallery post by its page URL, with its full-size images and terms. */
export async function scrapeGallery(url: string): Promise<ScrapeResult[]> {
  const slug = slugFromUrl(url);
  if (slug === "") return [];
  const parameters = new URLSearchParams({ slug, _embed: "1" });
  const posts = await fetchJson<WordPressPost[]>(`${API_URL}/posts?${parameters}`);
  if (posts.length === 0) return [];
  return [{ type: "gallery", gallery: galleryFromPost(posts[0]) }];
}

/** The post slug at the end of a site URL, e.g. "vina-rae-in-exotic-taste". */
export function slugFromUrl(url: string): string {
  const segments = new URL(url, BASE_URL).pathname.split("/").filter((segment) => segment !== "");
  if (segments.length === 0) return "";
  return segments[segments.length - 1];
}
