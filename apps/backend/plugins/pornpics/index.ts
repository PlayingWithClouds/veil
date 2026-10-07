import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { galleryList } from "./src/galleries.ts";
import { tagList } from "./src/categories.ts";

runPlugin({
  meta: {
    name: "pornpics",
    display_name: "PornPics",
    description: "Adult image galleries — browse, search and scrape photo sets from pornpics.com",
    icon: "https://www.pornpics.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["gallery:find", "gallery:list", "tag:list"],
    domains: ["pornpics.com"],
  },
  find: (url) => scrape(url),
  galleryList,
  tagList,
});
