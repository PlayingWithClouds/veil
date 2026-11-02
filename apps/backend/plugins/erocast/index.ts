import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeAudio } from "./src/audio.ts";
import { audioList } from "./src/listing.ts";
import { suggest } from "./src/suggest.ts";

runPlugin({
  meta: {
    name: "erocast",
    display_name: "Erocast",
    description: "Free erotic audio and ASMR — browse, search and play the tracks hosted on erocast.me",
    icon: "https://www.erocast.me/skins/default/images/favicon.png",
    version: packageJson.version,
    capabilities: ["audio:find", "audio:list", "search:suggest"],
    domains: ["erocast.me", "www.erocast.me"],
    content: ["audio"],
  },
  find: (url) => scrapeAudio(url),
  audioList,
  suggest,
});
