import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeGallery } from "./src/galleries.ts";
import { galleryList } from "./src/listing.ts";

runPlugin({
  meta: {
    name: "erocurves",
    display_name: "Erocurves",
    description: "Adult photo galleries — browse, search and scrape photo sets from erocurves.com",
    icon: "https://www.erocurves.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["gallery:find", "gallery:list"],
    domains: ["erocurves.com"],
  },
  find: (url) => scrapeGallery(url),
  galleryList,
});
