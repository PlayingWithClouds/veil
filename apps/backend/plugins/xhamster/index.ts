import { runPlugin } from "@playingwithclouds/veil-sdk";
import type { ScrapeResult } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";
import { isPornstarUrl, performerFind } from "./src/performer.ts";
import { identifyPerformer } from "./src/identify.ts";

/** Routes a find by URL shape: pornstar profiles to the performer scraper, everything else to the video scraper. */
function find(url: string): Promise<ScrapeResult[]> {
  if (isPornstarUrl(url)) return performerFind(url);
  return scrape(url);
}

runPlugin({
  meta: {
    name: "xhamster",
    display_name: "xHamster",
    description: "Adult streaming site — browse, search and scrape videos from xhamster.com",
    icon: "https://xhamster.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "performer:find", "performer:identify", "scene:list", "stream:resolve"],
    domains: ["xhamster.com"],
  },
  find,
  identifyPerformer: (args) => identifyPerformer(args),
  sceneList,
  resolve: (url) => resolve(url),
});
