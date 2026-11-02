import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";
import { isActressUrl, performerFind, performerIdentify } from "./src/performers.ts";

runPlugin({
  meta: {
    name: "missav",
    display_name: "MissAV",
    description: "Adult JAV streaming site — browse, search and scrape videos from missav.ws",
    icon: "https://missav.ws/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "performer:find", "performer:identify", "scene:list", "scene:list:page", "stream:resolve"],
    domains: ["missav.ws"],
    content: ["jav"],
    requires_solver: true,
  },
  // Video pages and actress profiles share the domain; route by URL shape.
  find: (url) => (isActressUrl(url) ? performerFind(url) : scrape(url)),
  identifyPerformer: (args) => performerIdentify(args),
  sceneList,
  resolve: (url) => resolve(url),
});
