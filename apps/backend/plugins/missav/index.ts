import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "missav",
    display_name: "MissAV",
    description: "Adult JAV streaming site — browse, search and scrape videos from missav.ws",
    icon: "https://missav.ws/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    domains: ["missav.ws"],
    requires_solver: true,
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
