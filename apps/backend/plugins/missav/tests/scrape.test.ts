import { test, expect, describe } from "bun:test";
import { canonicalUrl } from "../src/http.ts";
import { parseScenePage } from "../src/scrape.ts";
import type { Scene, ScrapeResult } from "@playingwithclouds/veil-sdk";

// Saved video pages. Both carry the site header, whose genre nav links
// (/en/genres/VR among them) must not leak into the scene's tags.
const mideHtml = await Bun.file(`${import.meta.dir}/fixtures/mide-900.html`).text();
const ewdxHtml = await Bun.file(`${import.meta.dir}/fixtures/ewdx-599.html`).text();

/** The scene of a parsed page (always the first result). */
function sceneOf(results: ScrapeResult[]): Scene {
  const first = results[0];
  if (first.type !== "scene") throw new Error("expected the scene first");
  return first.scene;
}

describe("canonicalUrl", () => {
  test("drops the /dm<n> mirror prefix", () => {
    expect(canonicalUrl("https://missav.ws/dm31/en/actresses/Kazumi%20Yanagida")).toBe(
      "https://missav.ws/en/actresses/Kazumi%20Yanagida"
    );
    expect(canonicalUrl("/en/makers/FALENO")).toBe("https://missav.ws/en/makers/FALENO");
  });
});

describe("parseScenePage", () => {
  test("reads tags and performers from the info panel only", () => {
    const scene = sceneOf(parseScenePage(mideHtml, "https://missav.ws/en/mide-900"));

    expect(scene.tags).toEqual([
      "Beautiful Breasts",
      "Pretty Girl",
      "Documentary",
      "Individual",
      "Selfie",
      "Slim Pixelated",
      "Hd",
      "Exclusive",
      "MIDE-900",
    ]);
    expect(scene.performers).toEqual([
      {
        name: "Nozomi Ishihara",
        order: 0,
        source_url: "https://missav.ws/en/actresses/Nozomi%20Ishihara",
      },
    ]);
    expect(scene.date).toBe("2021-03-13");
  });

  test("credits the maker as studio and emits a differing label under it", () => {
    const results = parseScenePage(mideHtml, "https://missav.ws/en/mide-900");

    expect(sceneOf(results).studio).toEqual({
      name: "Moody's",
      source_url: "https://missav.ws/en/makers/Moody%27s",
    });
    expect(results[1]).toEqual({
      type: "studio",
      studio: {
        type: "studio",
        external_id: "missav-label-MOODYZ DIVA",
        source_url: "https://missav.ws/en/labels/MOODYZ%20DIVA",
        name: "MOODYZ DIVA",
        parent: "Moody's",
      },
    });
  });

  test("skips the label when it matches the maker and tolerates a missing cast", () => {
    const results = parseScenePage(ewdxHtml, "https://missav.ws/en/ewdx-599");
    const scene = sceneOf(results);

    expect(results).toHaveLength(1);
    expect(scene.tags).toEqual(["Wife", "Big Breasts", "Creampie", "Hd", "4K", "EWDX-599"]);
    expect(scene.tags).not.toContain("VR");
    expect(scene.performers).toEqual([]);
    expect(scene.studio?.name).toBe("E★人妻DX");
  });
});
