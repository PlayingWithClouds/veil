import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrapeGallery } from "./src/gallery.ts";
import { galleryList } from "./src/listing.ts";

runPlugin({
  meta: {
    name: "hqporn-pics",
    display_name: "HQporn.pics",
    description: "Adult photo galleries — browse and scrape photo sets from hqporn.pics",
    icon: "https://hqporn.pics/favicon-32x32.png",
    version: packageJson.version,
    capabilities: ["gallery:find", "gallery:list"],
    domains: ["hqporn.pics"],
  },
  find: (url) => scrapeGallery(url),
  galleryList,
});
