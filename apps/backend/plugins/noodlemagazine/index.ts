import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeScene } from "./src/scene.ts";
import { sceneList } from "./src/listing.ts";
import { resolve } from "./src/resolve.ts";
import { suggest } from "./src/suggest.ts";

// One plugin for the site and its copies (ukdevilz.com, mat6tube.com): the
// same videos under the same ids, fetched from the first copy that answers
// (src/mirrors.ts). It needs no solver itself: without FlareSolverr only the
// copies behind Cloudflare are skipped.
runPlugin({
  meta: {
    name: "noodlemagazine",
    display_name: "Noodlemagazine",
    description:
      "Video search engine — search, browse and play videos from noodlemagazine.com, falling back to its copies ukdevilz.com and mat6tube.com",
    icon: "https://mat6tube.com/static/extend/dark/favicon-96x96.png",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "stream:resolve", "search:suggest"],
    domains: ["noodlemagazine.com", "ukdevilz.com", "mat6tube.com"],
  },
  find: (url) => scrapeScene(url),
  sceneList,
  resolve: (url) => resolve(url),
  suggest,
});
