import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { identifyFace } from "./src/identify.ts";

runPlugin({
  meta: {
    name: "namethatporn",
    display_name: "NameThatPorn",
    description:
      "Reverse image search over NameThatPorn.com's identification posts: the posts whose picture looks like yours, and the pornstars their accepted answers name.",
    icon: "https://namethatporn.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["face:identify"],
    domains: ["namethatporn.com"],
  },
  identifyFace,
});
