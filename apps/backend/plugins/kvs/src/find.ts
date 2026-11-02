import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import { profileFind, profileSlug } from "./profile.ts";
import { scrape } from "./scrape.ts";
import { siteForUrl } from "./sites.ts";

/** scene:find and performer:find: a model page is read as a profile, any other URL as a video. */
export function find(url: string): Promise<ScrapeResult[]> {
  const site = siteForUrl(url);
  if (site && profileSlug(site, url) !== "") return profileFind(url);
  return scrape(url);
}
