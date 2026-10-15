import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { identifyFace } from "./src/identify.ts";

runPlugin({
  meta: {
    name: "pornstarbyface",
    display_name: "PornstarByFace",
    description:
      "Face lookalike search over professional adult performers (pornstarbyface.com). Its database holds only performers, never private people.",
    icon: "https://pornstarbyface.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["face:identify"],
    domains: ["pornstarbyface.com"],
  },
  identifyFace,
});
