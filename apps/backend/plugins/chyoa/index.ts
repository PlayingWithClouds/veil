import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { storyList } from "./src/listing.ts";
import { checkAccount, login } from "./src/login.ts";
import { findStory, readChapter } from "./src/story.ts";

runPlugin({
  meta: {
    name: "chyoa",
    display_name: "CHYOA",
    description: "Choose-your-own-adventure erotic stories from chyoa.com: browse, search, and read chapter by chapter along the branches (reading needs a free account)",
    icon: "https://chyoa.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["story:list", "story:find", "story:chapter", "account:login", "account:check"],
    domains: ["chyoa.com"],
    content: ["stories"],
    login: {
      fields: [
        { key: "username", label: "Username", type: "string", required: true },
        { key: "password", label: "Password", type: "password", required: true },
      ],
      cookies: true,
    },
  },
  find: (url) => findStory(url),
  storyList,
  readChapter,
  login,
  checkAccount,
});
