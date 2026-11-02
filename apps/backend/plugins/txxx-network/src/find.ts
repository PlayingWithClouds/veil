import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import { modelDir, profileFind } from "./profile.ts";
import { scrape } from "./scrape.ts";

/** scene:find and performer:find: a /models/<dir>/ page is read as a profile, any other URL as a video. */
export function find(url: string): Promise<ScrapeResult[]> {
  if (modelDir(url) !== "") return profileFind(url);
  return scrape(url);
}
