import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { findAudio } from "./src/audio.ts";
import { audioList } from "./src/listing.ts";

runPlugin({
  meta: {
    name: "soundgasm",
    display_name: "Soundgasm",
    description: "Free erotic audio host — read one audio or list a user's audios from soundgasm.net",
    icon: "https://soundgasm.net/favicon.ico",
    version: packageJson.version,
    capabilities: ["audio:find", "audio:list"],
    domains: ["soundgasm.net"],
    content: ["audio"],
  },
  find: (url) => findAudio(url),
  audioList,
});
