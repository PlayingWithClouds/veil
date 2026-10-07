import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "xhamster",
    display_name: "xHamster",
    description: "Adult streaming site — browse, search and scrape videos from xhamster.com",
    icon: "https://xhamster.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    domains: ["xhamster.com"],
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
