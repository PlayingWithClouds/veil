import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { performerList, tagList } from "./src/catalog.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "hqporner",
    display_name: "HQporner",
    description: "Adult streaming site — browse, search and scrape videos from hqporner.com",
    icon: "https://hqporner.com/favicon.ico",
    version: packageJson.version,
    // find handles video, actress and category URLs — hence three find entities.
    capabilities: [
      "scene:find",
      "performer:find",
      "tag:find",
      "scene:list",
      "scene:list:page",
      "performer:list",
      "tag:list",
      "stream:resolve",
    ],
    // The scraped stream links point at the mydaddy.cc embed player.
    domains: ["mydaddy.cc"],
  },
  find: (url) => scrape(url),
  sceneList,
  performerList,
  tagList,
  resolve: (url) => resolve(url),
});
