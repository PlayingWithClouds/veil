import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapePerformer } from "./src/bio.ts";
import { performerList } from "./src/listing.ts";

runPlugin({
  meta: {
    name: "freeones",
    display_name: "FreeOnes",
    description: "Performer database — browse, search and read the profiles of adult performers on freeones.com",
    icon: "https://www.freeones.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["performer:find", "performer:list"],
    domains: ["freeones.com"],
  },
  find: (url) => scrapePerformer(url),
  performerList,
});
