import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { performerList, tagList } from "./src/catalog.ts";
import { resolve } from "./src/resolve.ts";
import { SITES, supportedDomains } from "./src/sites.ts";

runPlugin({
  meta: {
    name: "kvs",
    display_name: "KVS Tubes",
    description: `Tube sites running Kernel Video Sharing (${SITES.map((site) => site.name).join(", ")}) — search, browse, scrape and resolve, one config entry per site`,
    icon: "https://www.kernel-video-sharing.com/favicon.ico",
    version: packageJson.version,
    capabilities: [
      "scene:find",
      "scene:list",
      "scene:list:page",
      "performer:list",
      "tag:list",
      "stream:resolve",
    ],
    // find/resolve derive the site from the URL, so every site routes here.
    domains: supportedDomains(),
    settings: [
      {
        key: "KVS_SITES",
        label: "Sites",
        description: `Comma-separated sites that search and browse span (keys or hosts, e.g. "okxxx, analdin.com"). Empty = all: ${SITES.map((site) => site.key).join(", ")}. Opening and playing videos works on every site regardless.`,
        type: "string",
        default: "",
      },
    ],
  },
  find: (url) => scrape(url),
  sceneList,
  performerList,
  tagList,
  resolve: (url) => resolve(url),
});
