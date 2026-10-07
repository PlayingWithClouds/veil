import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";
import { DEFAULT_HOST, NETWORK_DOMAINS } from "./src/http.ts";

runPlugin({
  meta: {
    name: "txxx-network",
    display_name: "TXXX Network",
    description:
      "TxxxNetwork adult sites (txxx, upornia, hclips, hdzog, vjav, …) — browse, search, scrape and resolve via their shared JSON API",
    // /favicon.ico is a near-transparent 16px placeholder; this is the real logo.
    icon: "https://txxx.com/static/images/favicons/apple-touch-icon.png",
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "stream:resolve"],
    // find/resolve derive the host from the video URL, so every network
    // domain routes to this plugin.
    domains: NETWORK_DOMAINS,
    settings: [
      {
        key: "TXXX_HOST",
        label: "Browse host",
        description:
          "Which TxxxNetwork site to list/search (e.g. txxx.com, hclips.com, hdzog.com). Finding and resolving work across all network sites regardless of this.",
        type: "string",
        default: DEFAULT_HOST,
      },
    ],
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
