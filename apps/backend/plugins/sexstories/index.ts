import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { storyList } from "./src/listing.ts";
import { findStory, readChapter } from "./src/story.ts";

runPlugin({
  meta: {
    name: "sexstories",
    display_name: "Sexstories",
    description: "Free erotic stories from sexstories.com (XNXX's story site): browse the latest, search, read",
    icon: "https://www.sexstories.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["story:list", "story:find", "story:chapter"],
    domains: ["sexstories.com"],
    content: ["stories"],
  },
  find: (url) => findStory(url),
  storyList,
  readChapter,
});
