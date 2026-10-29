import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "xvideos",
    display_name: "XVideos",
    description:
      "Adult tube site — search, browse and play videos from xvideos.com; also opens and plays xnxx.com video pages, which share its player.",
    icon: "https://www.xvideos.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "stream:resolve"],
    // xnxx is a sister site on the same CDN: its pages scrape and resolve
    // here, but search and listings stay on xvideos, which carries the same catalog.
    domains: ["xvideos.com", "xnxx.com"],
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
