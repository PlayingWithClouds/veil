import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { identifyFace } from "./src/identify.ts";

runPlugin({
  meta: {
    name: "facedigger",
    display_name: "FaceDigger",
    description:
      "Face lookalike search over live-cam models (facedigger.com, cam category only; its video, OnlyFans and general results are never used). Needs the owner to accept FaceDigger's user agreement in the plugin settings.",
    icon: "https://facedigger.com/favicon.ico",
    version: packageJson.version,
    capabilities: ["face:identify"],
    domains: ["facedigger.com"],
    settings: [
      {
        key: "FACEDIGGER_ACCEPT_AGREEMENT",
        label: "I accept FaceDigger's user agreement",
        description:
          "Required before any picture is sent. FaceDigger's agreement (shown on facedigger.com before a search) includes: you are 18+, you will not use results to approach, harass, stalk, extort or screen anyone, you will never upload minors, you are NOT a resident or citizen of the European Union, and your jurisdiction permits facial recognition search. Turn this on only if all of it is true for you.",
        type: "boolean",
        default: "false",
      },
    ],
  },
  identifyFace,
});
