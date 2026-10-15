import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { camList } from "./src/rooms.ts";
import { camStream } from "./src/stream.ts";

runPlugin({
  meta: {
    name: "chaturbate",
    display_name: "Chaturbate",
    description: "Live cams — the public rooms online right now on chaturbate.com",
    icon: "https://chaturbate.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["cam:list", "cam:stream"],
  },
  camList,
  camStream,
});
