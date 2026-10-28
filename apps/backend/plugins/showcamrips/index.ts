import { runPlugin } from "@playingwithclouds/veil-sdk";
import packageJson from "./package.json" with { type: "json" };
import { scrape } from "./src/scrape.ts";
import { sceneList } from "./src/scenes.ts";
import { resolve } from "./src/resolve.ts";
import { BASE_URL } from "./src/http.ts";

runPlugin({
  meta: {
    name: "showcamrips",
    display_name: "ShowCamRips",
    description:
      "Recordings of live cam shows (Chaturbate, Stripchat, Camsoda, …) from showcamrips.com — search by model name, browse by model, cam site or category",
    icon: `${BASE_URL}/favicon.ico`,
    version: packageJson.version,
    capabilities: ["scene:find", "scene:list", "scene:list:page", "stream:resolve"],
    // The play page and the player iframe are on the site itself.
    domains: ["showcamrips.com"],
    content: ["cam-shows"],
  },
  find: (url) => scrape(url),
  sceneList,
  resolve: (url) => resolve(url),
});
