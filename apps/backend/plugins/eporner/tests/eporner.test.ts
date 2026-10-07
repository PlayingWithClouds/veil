import { test, expect, describe } from "bun:test";

import { idFromUrl, embedUrl, toIsoDate, toRating, parseKeywords } from "../src/http.ts";
import { toScene, toDiscoveredItem, type EpVideo } from "../src/api.ts";
import { sceneList } from "../src/scenes.ts";

// ---------------------------------------------------------------------------
// Pure helpers (no network)

describe("http helpers", () => {
  test("idFromUrl extracts the id from page and embed urls", () => {
    expect(idFromUrl("https://www.eporner.com/video-tuvMM4RP0eA/some-slug/")).toBe("tuvMM4RP0eA");
    expect(idFromUrl("https://www.eporner.com/hd-porn/tuvMM4RP0eA/some-slug/")).toBe("tuvMM4RP0eA");
    expect(idFromUrl("https://www.eporner.com/embed/tuvMM4RP0eA/")).toBe("tuvMM4RP0eA");
    expect(idFromUrl("https://www.eporner.com/pornstar/foo/")).toBe("");
  });

  test("embedUrl builds the embed page url", () => {
    expect(embedUrl("abc123")).toBe("https://www.eporner.com/embed/abc123/");
  });

  test("toIsoDate normalizes the API date", () => {
    expect(toIsoDate("2026-07-03 10:16:57")).toBe("2026-07-03T10:16:57Z");
    expect(toIsoDate("")).toBeUndefined();
    expect(toIsoDate(undefined)).toBeUndefined();
  });

  test("toRating maps the 0-5 scale onto 0-10", () => {
    expect(toRating("5.00")).toBe(10);
    expect(toRating("2.5")).toBe(5);
    expect(toRating(undefined)).toBeUndefined();
    expect(toRating("0.00")).toBeUndefined();
  });

  test("parseKeywords cleans, dedupes and drops the title echo and junk", () => {
    const tags = parseKeywords("blowjob, Blowjob, , a, my title, big ass", "My Title");
    expect(tags).toEqual(["blowjob", "big ass"]);
  });
});

// ---------------------------------------------------------------------------
// API mappers

const fixture: EpVideo = {
  id: "tuvMM4RP0eA",
  title: "Sample",
  keywords: "blowjob, teens",
  rate: "4.00",
  url: "https://www.eporner.com/video-tuvMM4RP0eA/sample/",
  added: "2026-07-03 10:16:57",
  length_sec: 2157,
  embed: "https://www.eporner.com/embed/tuvMM4RP0eA/",
  default_thumb: { src: "https://cdn/x_360.jpg" },
  thumbs: [{ src: "https://cdn/1.jpg" }, { src: "https://cdn/2.jpg" }],
};

describe("mappers", () => {
  test("toDiscoveredItem carries the essentials", () => {
    const item = toDiscoveredItem(fixture);
    expect(item.external_id).toBe("eporner-tuvMM4RP0eA");
    expect(item.media_type).toBe("scene");
    expect(item.source_url).toBe(fixture.url);
    expect(item.poster_path).toBe("https://cdn/x_360.jpg");
    expect(item.preview_images).toEqual(["https://cdn/1.jpg", "https://cdn/2.jpg"]);
    expect(item.duration_seconds).toBe(2157);
  });

  test("toScene includes tags, duration, rating and an embed download", () => {
    const scene = toScene(fixture);
    expect(scene.external_id).toBe("eporner-tuvMM4RP0eA");
    expect(scene.duration_seconds).toBe(2157);
    expect(scene.rating).toBe(8);
    expect(scene.tags).toEqual(["blowjob", "teens"]);
    expect(scene.downloads).toEqual([
      { label: "eporner", url: "https://www.eporner.com/embed/tuvMM4RP0eA/" },
    ]);
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against the eporner API.

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
        expect(item.source_url).toContain("eporner.com");
        expect(item.external_id).toStartWith("eporner-");
        expect(item.media_type).toBe("scene");
      }
      expect(items.some((item) => (item.duration_seconds ?? 0) > 0)).toBe(true);
    },
    2 * NET_TIMEOUT
  );

  test(
    "offset still fills the requested window",
    async () => {
      const { items } = await sceneList({ limit: 2, offset: 2 });
      expect(items.length).toBe(2);
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
      expect(items.length).toBeLessThanOrEqual(10);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("eporner.com");
        expect(item.external_id).toStartWith("eporner-");
        expect(item.media_type).toBe("scene");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "blank query falls back to the catalog listing",
    async () => {
      const { items } = await sceneList({ query: "   ", limit: 2 });
      expect(items.length).toBe(2);
    },
    2 * NET_TIMEOUT
  );
});
