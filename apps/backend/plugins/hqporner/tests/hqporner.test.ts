import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  videoIdFromUrl,
  isVideoUrl,
  actressSlugFromUrl,
  categorySlugFromUrl,
  parseDurationSeconds,
  relativeDateToISO,
  providerFromUrl,
  BASE_URL,
} from "../src/http.ts";
import { sceneList } from "../src/scenes.ts";
import { performerList, tagList } from "../src/catalog.ts";
import { scrape } from "../src/scrape.ts";
import { resolve } from "../src/resolve.ts";

// ---------------------------------------------------------------------------
// Pure URL/slug helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/hdporn/1")).toBe(`${BASE_URL}/hdporn/1`);
    expect(resolveUrl("hdporn/1")).toBe(`${BASE_URL}/hdporn/1`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("videoIdFromUrl extracts the numeric id", () => {
    expect(videoIdFromUrl("/hdporn/126872-the_little_hole_that_can.html")).toBe("126872");
    expect(videoIdFromUrl(`${BASE_URL}/hdporn/1-a.html`)).toBe("1");
    expect(videoIdFromUrl("/hdporn/2")).toBe("");
  });

  test("isVideoUrl detects video detail pages", () => {
    expect(isVideoUrl("/hdporn/126872-foo.html")).toBe(true);
    expect(isVideoUrl("/hdporn/2")).toBe(false);
    expect(isVideoUrl("/categories")).toBe(false);
  });

  test("actressSlugFromUrl extracts the slug", () => {
    expect(actressSlugFromUrl("/actress/little-caprice")).toBe("little-caprice");
    expect(actressSlugFromUrl(`${BASE_URL}/actress/little-caprice/3`)).toBe("little-caprice");
    expect(actressSlugFromUrl("/hdporn/1-a.html")).toBe("");
  });

  test("categorySlugFromUrl extracts the slug", () => {
    expect(categorySlugFromUrl("/category/anal-sex-hd")).toBe("anal-sex-hd");
    expect(categorySlugFromUrl("/girls")).toBe("");
  });

  test("parseDurationSeconds handles h/m/s formats", () => {
    expect(parseDurationSeconds("33m 15s")).toBe(33 * 60 + 15);
    expect(parseDurationSeconds("1h 33m 16s")).toBe(3600 + 33 * 60 + 16);
    expect(parseDurationSeconds("59m 1s")).toBe(59 * 60 + 1);
    expect(parseDurationSeconds("")).toBe(0);
  });

  test("relativeDateToISO approximates relative dates", () => {
    const now = new Date("2026-07-13T12:00:00Z");
    expect(relativeDateToISO("today", now)).toBe("2026-07-13");
    expect(relativeDateToISO("yesterday", now)).toBe("2026-07-12");
    expect(relativeDateToISO("3 days ago", now)).toBe("2026-07-10");
    expect(relativeDateToISO("2 weeks ago", now)).toBe("2026-06-29");
    expect(relativeDateToISO("1 month ago", now)).toBe("2026-06-13");
    expect(relativeDateToISO("nonsense", now)).toBe("");
  });

  test("providerFromUrl returns the bare host label", () => {
    expect(providerFromUrl("https://mydaddy.cc/video/abc/")).toBe("mydaddy");
    expect(providerFromUrl("not a url")).toBe("unknown");
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against hqporner.com. Chained so they survive
// catalog changes (scrape a title that scene:find actually returned).

const NET_TIMEOUT = 30_000;
const QUERY = "milf";

describe("scene:list (live)", () => {
  test(
    "returns a bounded set of catalog stubs with durations",
    async () => {
      const { items } = await sceneList({ limit: 6 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(6);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("/hdporn/");
        expect(item.external_id).toStartWith("hqporner-");
        expect(item.media_type).toBe("scene");
      }
      expect(items.some((item) => (item.duration_seconds ?? 0) > 0)).toBe(true);
    },
    2 * NET_TIMEOUT
  );
});

describe("scene:list with query (live)", () => {
  test(
    "returns well-formed items for a common query",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 10 });
      expect(items.length).toBeGreaterThan(0);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("/hdporn/");
        expect(item.external_id).toStartWith("hqporner-");
        expect(item.media_type).toBe("scene");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("performer catalog (live)", () => {
  test(
    "performer:list returns actress stubs",
    async () => {
      const { items } = await performerList({ limit: 10 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(10);
      for (const item of items) {
        expect(item.media_type).toBe("performer");
        expect(item.source_url).toContain("/actress/");
        expect(item.external_id).toStartWith("hqporner-actress-");
        expect(item.title.trim()).not.toBe("");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "performer:list with query filters by name",
    async () => {
      const { items } = await performerList({ query: "caprice" });
      expect(items.length).toBeGreaterThan(0);
      for (const item of items) {
        expect(item.title.toLowerCase()).toContain("caprice");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("tag:list (live)", () => {
  test(
    "returns category stubs",
    async () => {
      const { items } = await tagList({ limit: 15 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(15);
      for (const item of items) {
        expect(item.media_type).toBe("tag");
        expect(item.source_url).toContain("/category/");
        expect(item.external_id).toStartWith("hqporner-category-");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape (live)", () => {
  test(
    "scrapes a scene with full info plus standalone performers",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 20 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.external_id).toStartWith("hqporner-");
      expect(scene!.source_url).toContain("/hdporn/");
      expect(scene!.duration_seconds ?? 0).toBeGreaterThan(0);
      expect(scene!.date ?? "").toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect((scene!.tags ?? []).length).toBeGreaterThan(0);

      for (const credit of scene!.performers ?? []) {
        expect(credit.source_url).toContain("/actress/");
        expect(credit.external_id).toStartWith("hqporner-actress-");
      }

      const performers = results.flatMap((r) => (r.type === "performer" ? [r.performer] : []));
      expect(performers.length).toBe((scene!.performers ?? []).length);
      for (const performer of performers) {
        expect(performer.name.trim()).not.toBe("");
        expect(performer.source_url).toContain("/actress/");
      }

      const urls = (scene!.downloads ?? []).map((d) => d.url);
      for (const url of urls) expect(url).toStartWith("http");
      expect(new Set(urls).size).toBe(urls.length);
    },
    4 * NET_TIMEOUT
  );

  test(
    "scrapes an actress page into a performer record",
    async () => {
      const results = await scrape(`${BASE_URL}/actress/little-caprice`);
      expect(results.length).toBe(1);
      const performer = results[0].type === "performer" ? results[0].performer : undefined;
      expect(performer).toBeDefined();
      expect(performer!.name.toLowerCase()).toContain("caprice");
      expect(performer!.external_id).toBe("hqporner-actress-little-caprice");
    },
    2 * NET_TIMEOUT
  );
});

describe("resolve (live)", () => {
  test(
    "resolves a scraped embed to a playable mp4",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 20 });
      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];
      const embedUrl = scene!.downloads?.[0]?.url;
      expect(embedUrl).toBeDefined();

      const resolved = await resolve(embedUrl!);
      expect(resolved.url).toStartWith("http");
      expect(resolved.url).toContain(".mp4");
      expect(resolved.mime_type).toBe("video/mp4");
    },
    3 * NET_TIMEOUT
  );
});
