import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";

runPlugin({
  meta: {
    name: "spankbang",
    display_name: "SpankBang",
    description: "Adult streaming site — browse, search and scrape videos from spankbang.com",
    icon: "https://spankbang.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    // The site sits behind Cloudflare; fetches route through FlareSolverr. The
    // in-page mp4/HLS sources are served straight from the video page URL.
    domains: ["spankbang.com"],
    requires_solver: true,
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
