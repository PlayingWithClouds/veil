import { test, expect, describe } from "bun:test";
import { actionTagMarkers } from "../src/markers.ts";
import { buildScene } from "../src/scrape.ts";
import { ALL_SITES } from "../src/sites.ts";

describe("pornhub action tags", () => {
  test("a scene page's actionTags become point markers", async () => {
    const html = await Bun.file(`${import.meta.dir}/fixtures/ph-scene-acts.html`).text();
    const site = ALL_SITES.find((candidate) => candidate.key === "pornhub")!;
    const scene = buildScene(site, "ph66f8c1d6a9c51", html);
    expect(scene.markers).toEqual([
      { tag: "Blowjob", seconds: 379 },
      { tag: "Handjob", seconds: 462 },
      { tag: "Reverse Cowgirl", seconds: 921 },
    ]);
  });

  test("an empty actionTags adds no markers", () => {
    expect(actionTagMarkers("")).toEqual([]);
    expect(actionTagMarkers(undefined)).toEqual([]);
  });

  test("malformed entries are skipped and repeated acts kept", () => {
    expect(actionTagMarkers("Doggystyle:10,junk,:5,Anal:x,Doggystyle:90")).toEqual([
      { tag: "Doggystyle", seconds: 10 },
      { tag: "Doggystyle", seconds: 90 },
    ]);
  });
});
