import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { identifyFace } from "./src/identify.ts";
import { findPost } from "./src/thread.ts";
import { searchPosts } from "./src/search.ts";

runPlugin({
  meta: {
    name: "namethatporn",
    display_name: "NameThatPorn",
    description:
      "Reverse image search over NameThatPorn.com's identification posts: the posts whose picture looks like yours, the pornstars their accepted answers name, each post's thread with its links, and a text search over the threads.",
    icon: "https://namethatporn.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["face:identify", "post:find", "post:search"],
    domains: ["namethatporn.com"],
  },
  identifyFace,
  findPost,
  searchPosts,
});
