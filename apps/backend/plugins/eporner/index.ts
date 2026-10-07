import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { find } from "./src/find.ts";
import { sceneList } from "./src/scenes.ts";
import { performerList } from "./src/performers.ts";
import { studioList } from "./src/studios.ts";
import { galleryList } from "./src/galleries.ts";
import { resolve } from "./src/resolve.ts";
import { enrich } from "./src/enrich.ts";

runPlugin({
  meta: {
    name: "eporner",
    display_name: "EPorner",
    description: "Adult site — videos, pornstars, channels and photo galleries via eporner.com",
    icon: "https://www.eporner.com/favicon.ico",
    version: packageJson.version,
    capabilities: [
      "scene:find",
      "scene:list",
      "scene:list:page",
      "performer:find",
      "performer:list",
      "studio:find",
      "studio:list",
      "gallery:find",
      "gallery:list",
      "stream:resolve",
      "enrich",
    ],
    // Resolve handles eporner's own pages/embeds and its CDN host.
    domains: ["eporner.com"],
  },
  find: (url) => find(url),
  sceneList,
  performerList,
  studioList,
  galleryList,
  resolve: (url) => resolve(url),
  enrich: (args) => enrich(args),
});
