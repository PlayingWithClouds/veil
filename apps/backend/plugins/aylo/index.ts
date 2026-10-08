import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { tagList } from "./src/tags.ts";
import { resolve } from "./src/resolve.ts";
import { SITE_DOMAINS } from "./src/sites.ts";

runPlugin({
  meta: {
    name: "aylo",
    display_name: "Aylo",
    description: "Aylo tube sites (Pornhub, RedTube, YouPorn, Tube8) — search, browse, scrape and resolve",
    icon: "https://www.pornhub.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "tag:list", "stream:resolve"],
    // find/resolve/list-by-page derive the site from the URL, so every site domain routes here.
    domains: SITE_DOMAINS,
    settings: [
      {
        key: "AYLO_SITES",
        label: "Sites",
        description:
          "Which sites search and browse cover, comma-separated (pornhub, redtube, youporn, tube8). Finding, resolving and page listings work on all four regardless.",
        type: "string",
        default: "pornhub,redtube,youporn,tube8",
      },
    ],
  },
  find: (url) => scrape(url),
  sceneList,
  tagList,
  resolve: (url) => resolve(url),
});
