import { test, expect, describe } from "bun:test";
import { sceneList } from "../src/catalog.ts";
import { scrapeScene } from "../src/scene.ts";
import { resolve } from "../src/resolve.ts";
import type { Scene } from "@playingwithclouds/veil-sdk";

// Live checks against mat6tube.com: search, open the first hit, resolve its
// best download, and page the newest listing.

const LIVE_TIMEOUT_MS = 90_000;

describe("mat6tube (live)", () => {
  test(
    "search, find, resolve",
    async () => {
      const { items } = await sceneList({ query: "blonde", limit: 5 });
      expect(items.length).toBeGreaterThan(0);
      for (const item of items) {
        expect(item.external_id.startsWith("mat6tube-")).toBe(true);
        expect(item.title.length).toBeGreaterThan(0);
      }

      const results = await scrapeScene(items[0]!.source_url);
      const scene = (results[0] as { scene: Scene }).scene;
      expect(scene.external_id).toBe(items[0]!.external_id);
      expect(scene.title.length).toBeGreaterThan(0);
      expect(scene.poster_path).toBeTruthy();
      expect(scene.downloads?.length).toBeGreaterThan(0);

      const resolved = await resolve(scene.downloads![0]!.url);
      expect(resolved.url.startsWith("https://")).toBe(true);
      expect(resolved.mime_type).toBe("video/mp4");
    },
    LIVE_TIMEOUT_MS
  );

  test(
    "the newest listing pages stay distinct",
    async () => {
      const firstPage = await sceneList({ limit: 30 });
      const laterPage = await sceneList({ limit: 30, offset: 60 });
      expect(firstPage.items).toHaveLength(30);
      expect(laterPage.items).toHaveLength(30);
      const firstIds = new Set(firstPage.items.map((item) => item.external_id));
      expect(laterPage.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    LIVE_TIMEOUT_MS
  );
});
