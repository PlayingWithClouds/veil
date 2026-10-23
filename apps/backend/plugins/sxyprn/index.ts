import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeScene } from "./src/scene.ts";
import { sceneList } from "./src/listing.ts";
import { resolve } from "./src/resolve.ts";
import { suggest } from "./src/suggest.ts";
import { SITE_HOST } from "./src/http.ts";

runPlugin({
  meta: {
    name: "sxyprn",
    display_name: "SxyPrn",
    description: "Free video tube — search, browse and play the videos posted on sxyprn.com",
    icon: `https://${SITE_HOST}/favicon.ico`,
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "stream:resolve", "search:suggest"],
    domains: [SITE_HOST],
    expiring_media: true,
  },
  find: (url) => scrapeScene(url),
  sceneList,
  resolve,
  suggest,
});
