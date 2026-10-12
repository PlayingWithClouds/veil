import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeScene } from "./src/scene.ts";
import { sceneList } from "./src/catalog.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "mat6tube",
    display_name: "Mat6Tube",
    description: "Adult video catalog — search, browse, scrape and resolve the free videos of mat6tube.com",
    icon: "https://mat6tube.com/static/extend/dark/favicon-96x96.png",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    domains: ["mat6tube.com"],
  },
  find: (url) => scrapeScene(url),
  sceneList,
  resolve,
});
