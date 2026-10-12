import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeGallery } from "./src/album.ts";
import { galleryList } from "./src/listing.ts";

runPlugin({
  meta: {
    name: "pictoa",
    display_name: "Pictoa",
    description: "Adult photo albums — browse, search and scrape photo sets from pictoa.com",
    icon: "https://www.pictoa.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["gallery:find", "gallery:list"],
    domains: ["pictoa.com"],
  },
  find: (url) => scrapeGallery(url),
  galleryList,
});
