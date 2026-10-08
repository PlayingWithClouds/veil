import { test, expect, describe } from "bun:test";
import { SITES } from "../src/sites.ts";
import { sceneList } from "../src/scenes.ts";
import { performerList, tagList } from "../src/catalog.ts";
import { scrape } from "../src/scrape.ts";
import { resolve } from "../src/resolve.ts";
import type { Scene } from "@playingwithclouds/veil-sdk";

// Live checks against every configured site: search, open the first hit and
// resolve its best download. They hit the real sites, like the other plugins'
// live tests.

const LIVE_TIMEOUT_MS = 90_000;

// Cloudflare answers Bun's TLS fingerprint with a challenge on these sites,
// while the backend's embedded (Go) runtime gets through; they are checked
// with the EmbeddedBundlesLive test instead.
const BUN_BLOCKED = new Set(["fapnado"]);

/** Runs `body` with KVS_SITES limited to one site. */
async function withSite<T>(key: string, body: () => Promise<T>): Promise<T> {
  const previous = process.env.KVS_SITES;
  process.env.KVS_SITES = key;
  try {
    return await body();
  } finally {
    if (previous === undefined) delete process.env.KVS_SITES;
    else process.env.KVS_SITES = previous;
  }
}

describe("every site (live)", () => {
  for (const site of SITES) {
    if (BUN_BLOCKED.has(site.key)) continue;
    test(
      `${site.key}: search, find, resolve`,
      async () => {
        const { items } = await withSite(site.key, () => sceneList({ query: "blonde", limit: 5 }));
        expect(items.length).toBeGreaterThan(0);
        for (const item of items) {
          expect(item.external_id.startsWith(`kvs-${site.key}-`)).toBe(true);
          expect(item.title.length).toBeGreaterThan(0);
        }

        const results = await scrape(items[0].source_url);
        const scene = (results[0] as { scene: Scene }).scene;
        expect(scene.external_id).toBe(items[0].external_id);
        expect(scene.title.length).toBeGreaterThan(0);
        expect(scene.poster_path).toBeTruthy();
        expect(scene.downloads?.length).toBeGreaterThan(0);

        const resolved = await resolve(scene.downloads![0].url);
        expect(resolved.url.startsWith("http")).toBe(true);
        expect(resolved.headers?.Referer).toBe(site.origin + "/");
      },
      LIVE_TIMEOUT_MS
    );
  }
});

describe("listings (live)", () => {
  test(
    "browse interleaves sites and pages stay distinct",
    async () => {
      const firstPage = await withSite("okxxx,analdin,w1mp", () => sceneList({ limit: 12 }));
      const secondPage = await withSite("okxxx,analdin,w1mp", () => sceneList({ limit: 12, offset: 12 }));
      const sites = new Set(firstPage.items.map((item) => item.external_id.split("-")[1]));
      expect(sites.size).toBe(3);
      const firstIds = new Set(firstPage.items.map((item) => item.external_id));
      expect(secondPage.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    LIVE_TIMEOUT_MS
  );

  test(
    "scene:list:page follows a page through async pagination",
    async () => {
      const url = "https://www.analdin.com/categories/big-dick/";
      const firstPage = await sceneList({ url, limit: 30 });
      const laterPage = await sceneList({ url, limit: 30, offset: 120 });
      expect(firstPage.items).toHaveLength(30);
      expect(laterPage.items).toHaveLength(30);
      const firstIds = new Set(firstPage.items.map((item) => item.external_id));
      expect(laterPage.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    LIVE_TIMEOUT_MS
  );

  test(
    "tag and performer catalogs",
    async () => {
      const tags = await withSite("okxxx,anysex", () => tagList({ limit: 20 }));
      expect(tags.items.length).toBe(20);
      expect(tags.items.every((item) => item.media_type === "tag")).toBe(true);
      const models = await withSite("porngo", () => performerList({ query: "a", limit: 10 }));
      expect(models.items.length).toBeGreaterThan(0);
      expect(models.items.every((item) => item.title.toLowerCase().includes("a"))).toBe(true);
    },
    LIVE_TIMEOUT_MS
  );
});
