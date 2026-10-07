import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import { scrape } from "./scrape.ts";
import { performerFind } from "./performers.ts";
import { studioFind } from "./studios.ts";
import { galleryFind } from "./galleries.ts";

// One find handler serves every "<entity>:find" capability; dispatch by the
// eporner URL shape to the right scraper.
export function find(url: string): Promise<ScrapeResult[]> {
  if (url.includes("/pornstar/")) return performerFind(url);
  if (url.includes("/channel/")) return studioFind(url);
  if (url.includes("/gallery/")) return galleryFind(url);
  return scrape(url);
}
