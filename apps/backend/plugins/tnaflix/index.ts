import { runPlugin } from "@playingwithclouds/veil-sdk";
import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";
import { suggest } from "./src/suggest.ts";
import { isProfileUrl, performerFind } from "./src/performer.ts";
import { identifyPerformer } from "./src/identify.ts";

/** Routes a find by URL shape: profile pages to the performer scraper, everything else to the video scraper. */
function find(url: string): Promise<ScrapeResult[]> {
  if (isProfileUrl(url)) return performerFind(url);
  return scrape(url);
}

runPlugin({
  meta: {
    name: "tnaflix",
    display_name: "TNAflix",
    description: "Adult streaming site — browse, search and scrape videos from tnaflix.com",
    icon: "https://www.tnaflix.com/assets/img/favicon/tnaflix/favicon-32x32.png",
    version: packageJson.version,
    capabilities: ["scene:find", "performer:find", "performer:identify", "scene:list", "scene:list:page", "stream:resolve", "search:suggest"],
    // Downloads carry the tnaflix video page URL; the resolver re-fetches it
    // for a freshly signed mp4.
    domains: ["tnaflix.com"],
  },
  find,
  identifyPerformer: (args) => identifyPerformer(args),
  sceneList,
  resolve: (url) => resolve(url),
  suggest,
});
