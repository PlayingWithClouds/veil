import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeScene } from "./src/scene.ts";
import { sceneList } from "./src/catalog.ts";

runPlugin({
  meta: {
    name: "porn-star",
    display_name: "Porn-Star",
    description: "Adult video catalog — browse and scrape the free pornstar video scenes of porn-star.com",
    icon: "https://www.porn-star.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list"],
    domains: ["porn-star.com"],
  },
  find: (url) => scrapeScene(url),
  sceneList,
});
