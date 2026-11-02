import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { storyList } from "./src/listing.ts";
import { findStory, readChapter } from "./src/story.ts";

runPlugin({
  meta: {
    name: "literotica",
    display_name: "Literotica",
    description: "Erotic fiction — search, browse and read the stories and series on literotica.com",
    icon: "https://speedy.literotica.com/so/assets/favicon-C0wgbqSh.ico",
    version: packageJson.version,
    capabilities: ["story:find", "story:list", "story:chapter"],
    content: ["stories"],
    domains: ["literotica.com", "www.literotica.com"],
  },
  find: (url) => findStory(url),
  storyList,
  readChapter,
});
