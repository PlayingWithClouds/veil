import { test, expect, describe } from "bun:test";

import {
  videoIdFromUrl,
  isVideoUrl,
  hostFromUrl,
  videoApiPath,
  base164Decode,
  configuredHost,
  DEFAULT_HOST,
} from "../src/http.ts";
import { parseDurationSeconds } from "../src/listing.ts";
import { sceneList } from "../src/scenes.ts";
import { scrape } from "../src/scrape.ts";
import { resolve } from "../src/resolve.ts";

// ---------------------------------------------------------------------------
// Pure helpers (no network)

describe("http helpers", () => {
  test("videoIdFromUrl extracts the numeric id", () => {
    expect(videoIdFromUrl("https://txxx.com/videos/21707775/some-slug/")).toBe("21707775");
    expect(videoIdFromUrl("https://hclips.com/videos/10672515/x")).toBe("10672515");
    expect(videoIdFromUrl("https://txxx.com/video/977501/x")).toBe("977501");
    expect(videoIdFromUrl("https://txxx.com/latest-updates/")).toBe("");
  });

  test("videoIdFromUrl reads the id from a stored get_file handle", () => {
    const handle =
      "https://txxx.com/get_file/4/abc123def/977000/977501/977501_hq.mp4/?d=480&br=314&ti=1783872263";
    expect(videoIdFromUrl(handle)).toBe("977501");
  });

  test("isVideoUrl detects video pages", () => {
    expect(isVideoUrl("https://txxx.com/videos/123/x/")).toBe(true);
    expect(isVideoUrl("https://txxx.com/categories/asian/")).toBe(false);
  });

  test("hostFromUrl strips www and falls back to configured host", () => {
    expect(hostFromUrl("https://www.hdzog.com/videos/1/x")).toBe("hdzog.com");
    expect(hostFromUrl("not a url")).toBe(configuredHost());
  });

  test("configuredHost defaults when unset", () => {
    expect(configuredHost()).toBe(DEFAULT_HOST);
  });

  test("videoApiPath buckets the id by million and thousand", () => {
    expect(videoApiPath("21707775")).toBe(
      "/api/json/video/8640000/21000000/21707000/21707775.json"
    );
    expect(videoApiPath("977501")).toBe("/api/json/video/8640000/0/977000/977501.json");
  });

  test("parseDurationSeconds handles MM:SS and HH:MM:SS", () => {
    expect(parseDurationSeconds("10:05")).toBe(10 * 60 + 5);
    expect(parseDurationSeconds("1:02:03")).toBe(3600 + 2 * 60 + 3);
    expect(parseDurationSeconds("bogus")).toBe(0);
    expect(parseDurationSeconds(undefined)).toBe(0);
  });

  test("base164Decode reverses the site's obfuscation", () => {
    // Real videofile.php payload (homoglyph-obfuscated) → /get_file/ path.
    const encoded =
      "L2dldF9maWxlLzЕxL2QwNjА0ZDFmМjVhNDVmNGМyNTZhNzgzODdiOTВhY2Q3N2FlNmМwМGNhМА8yМTcwNzАwМА8yМTcwNzc3NS8yМTcwNzc3NV9ocS5tcDQvP2Q9NjQ5JmJyPTk3JnRpPTЕ3ODМ4Njk1NDU~";
    const decoded = base164Decode(encoded);
    expect(decoded).toStartWith("/get_file/");
    expect(decoded).toContain("21707775_hq.mp4");
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against the TxxxNetwork API. Chained so they
// survive catalog changes.

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
        expect(videoIdFromUrl(item.source_url)).not.toBe("");
        expect(item.external_id).toStartWith("txxx-");
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
        expect(item.external_id).toStartWith("txxx-");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape + resolve (live)", () => {
  test(
    "scrapes a real video and resolves a playable mp4",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 10 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.external_id).toStartWith("txxx-");
      // Every available quality is offered as a download.
      expect(scene!.downloads?.length).toBeGreaterThan(0);
      for (const download of scene!.downloads ?? []) {
        expect(download.url).toStartWith("http");
      }

      // Resolving a stored get_file handle re-derives a fresh signed source.
      const resolved = await resolve(scene!.downloads![0].url);
      expect(resolved.url).toContain("/get_file/");
      expect(resolved.mime_type).toBe("video/mp4");
    },
    3 * NET_TIMEOUT
  );
});
