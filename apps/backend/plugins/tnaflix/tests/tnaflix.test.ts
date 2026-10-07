import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  videoIdFromUrl,
  isVideoUrl,
  decodeEntities,
  parseDurationSeconds,
  BASE_URL,
} from "../src/http.ts";
import { sceneList } from "../src/scenes.ts";
import { scrape } from "../src/scrape.ts";
import { resolve } from "../src/resolve.ts";

// ---------------------------------------------------------------------------
// Pure URL/slug helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/new/1")).toBe(`${BASE_URL}/new/1`);
    expect(resolveUrl("new/1")).toBe(`${BASE_URL}/new/1`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("videoIdFromUrl extracts the trailing numeric id", () => {
    expect(videoIdFromUrl("/amateur-porn/Some-Clip/video25810750")).toBe("25810750");
    expect(videoIdFromUrl(`${BASE_URL}/anal-porn/Other-Clip/video11723565`)).toBe("11723565");
    expect(videoIdFromUrl("/amateur-porn/Some-Clip/video25810750?hd=1")).toBe("25810750");
    expect(videoIdFromUrl("/new/1")).toBe("");
    expect(videoIdFromUrl("/profile/somebody")).toBe("");
  });

  test("isVideoUrl detects video detail pages", () => {
    expect(isVideoUrl("/amateur-porn/Some-Clip/video25810750")).toBe(true);
    expect(isVideoUrl("/new/1")).toBe(false);
    expect(isVideoUrl("/search?what=milf")).toBe(false);
  });

  test("decodeEntities unescapes common HTML entities", () => {
    expect(decodeEntities("Kasumi&#039;s Bath &amp; Handjob")).toBe("Kasumi's Bath & Handjob");
    expect(decodeEntities("A &quot;quote&quot; &lt;here&gt;")).toBe('A "quote" <here>');
    expect(decodeEntities("trailer.mp4?a=1&amp;b=2")).toBe("trailer.mp4?a=1&b=2");
  });

  test("parseDurationSeconds handles MM:SS and HH:MM:SS chips", () => {
    expect(parseDurationSeconds("03:00")).toBe(180);
    expect(parseDurationSeconds("07:33")).toBe(7 * 60 + 33);
    expect(parseDurationSeconds("01:07:09")).toBe(3600 + 7 * 60 + 9);
    expect(parseDurationSeconds(" 05:42 ")).toBe(5 * 60 + 42);
    expect(parseDurationSeconds("")).toBe(0);
    expect(parseDurationSeconds("soon")).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against tnaflix.com. Chained so they survive
// catalog changes (scrape a video that scene:list actually returned).

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
        expect(item.external_id).toStartWith("tnaflix-");
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
        expect(videoIdFromUrl(item.source_url)).not.toBe("");
        expect(item.external_id).toStartWith("tnaflix-");
        expect(item.media_type).toBe("scene");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape + resolve (live)", () => {
  test(
    "scrapes a real video surfaced by scene:list and resolves a playable mp4",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 20 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.external_id).toStartWith("tnaflix-");
      expect(scene!.duration_seconds).toBeGreaterThan(0);
      // Every inlined quality is offered as a download.
      expect(scene!.downloads?.length).toBeGreaterThan(0);
      for (const download of scene!.downloads ?? []) {
        expect(download.url).toStartWith("http");
        expect(download.url).toContain(".mp4");
      }

      const resolved = await resolve(scene!.source_url);
      expect(resolved.url).toStartWith("http");
      expect(resolved.url).toContain(".mp4");
      expect(resolved.mime_type).toBe("video/mp4");
    },
    3 * NET_TIMEOUT
  );
});
