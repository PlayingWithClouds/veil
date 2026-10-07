import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "tnaflix",
    display_name: "TNAflix",
    description: "Adult streaming site — browse, search and scrape videos from tnaflix.com",
    icon: "https://www.tnaflix.com/assets/img/favicon/tnaflix/favicon-32x32.png",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    // Downloads carry the tnaflix video page URL; the resolver re-fetches it
    // for a freshly signed mp4.
    domains: ["tnaflix.com"],
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
