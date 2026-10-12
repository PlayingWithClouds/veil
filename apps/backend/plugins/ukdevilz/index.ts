import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeScene } from "./src/scene.ts";
import { sceneList } from "./src/listing.ts";
import { resolve } from "./src/resolve.ts";
import { suggest } from "./src/suggest.ts";

runPlugin({
  meta: {
    name: "ukdevilz",
    display_name: "Ukdevilz",
    description: "Free video site — browse, search and play videos from ukdevilz.com",
    icon: "https://ukdevilz.com/static/extend/dark/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "stream:resolve", "search:suggest"],
    domains: ["ukdevilz.com"],
    requires_solver: true,
  },
  find: (url) => scrapeScene(url),
  sceneList,
  resolve: (url) => resolve(url),
  suggest,
});
