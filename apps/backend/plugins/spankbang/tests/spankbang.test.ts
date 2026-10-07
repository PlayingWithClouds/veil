import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  videoCodeFromUrl,
  isVideoUrl,
  parseDurationSeconds,
  decodeEntities,
  BASE_URL,
} from "../src/http.ts";
import { extractSources } from "../src/streams.ts";
import { sceneList } from "../src/scenes.ts";
import { scrape } from "../src/scrape.ts";
import { resolve } from "../src/resolve.ts";

// ---------------------------------------------------------------------------
// Pure helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/new_videos/")).toBe(`${BASE_URL}/new_videos/`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("videoCodeFromUrl extracts the short code", () => {
    expect(videoCodeFromUrl("https://spankbang.com/7gu17/video/hostile+takeover")).toBe("7gu17");
    expect(videoCodeFromUrl("/a4v8m/video/emo+emily")).toBe("a4v8m");
    expect(videoCodeFromUrl("https://spankbang.com/new_videos/")).toBe("");
  });

  test("isVideoUrl detects video pages", () => {
    expect(isVideoUrl("https://spankbang.com/7gu17/video/x")).toBe(true);
    expect(isVideoUrl("https://spankbang.com/s/milf/")).toBe(false);
  });

  test("parseDurationSeconds reads the shorthand", () => {
    expect(parseDurationSeconds("21m")).toBe(21 * 60);
    expect(parseDurationSeconds("1h 5m")).toBe(3900);
    expect(parseDurationSeconds("45s")).toBe(45);
  });

  test("decodeEntities unescapes common entities", () => {
    expect(decodeEntities("A &amp; B")).toBe("A & B");
    expect(decodeEntities("Kasumi&#039;s")).toBe("Kasumi's");
  });
});

describe("stream extraction (no network)", () => {
  test("extractSources returns every quality plus the HLS master", () => {
    const html = `
      var stream_data = {
        '240p': ['https://cdn.sb/1/2/9-240p.mp4?secure=a,1'],
        '320p': [],
        '480p': ['https://cdn.sb/1/2/9-480p.mp4?secure=a,1'],
        '720p': [],
        '1080p': ['https://cdn.sb/1/2/9-1080p.mp4?secure=a,1'],
      };
      "https://hls-uranus.sb-cd.com/hls/1/2/9-,240p,480p,1080p,.mp4.urlset/master.m3u8?secure=a,1&_tid=9"
    `;
    const sources = extractSources(html);
    const mp4 = sources.filter((source) => source.format === "mp4");
    expect(mp4.map((source) => source.quality)).toEqual(["1080p", "480p", "240p"]);
    const hls = sources.find((source) => source.format === "hls");
    expect(hls?.url).toContain("master.m3u8");
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP through FlareSolverr. Skipped unless
// FLARESOLVERR_URL is set, since spankbang is behind Cloudflare.

const hasSolver = Boolean(process.env.FLARESOLVERR_URL);
const NET_TIMEOUT = 90_000;
const QUERY = "riley reid";

describe.if(hasSolver)("live (needs FlareSolverr)", () => {
  test(
    "scene:list returns bounded catalog stubs with durations",
    async () => {
      const { items } = await sceneList({ limit: 6 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(6);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.external_id).toStartWith("spankbang-");
        expect(item.media_type).toBe("scene");
      }
      expect(items.some((item) => (item.duration_seconds ?? 0) > 0)).toBe(true);
    },
    2 * NET_TIMEOUT
  );

  test(
    "scene:list with query returns relevant results, not the promoted block",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 12 });
      expect(items.length).toBeGreaterThan(0);
      const relevant = items.filter((item) => /riley/i.test(item.title)).length;
      expect(relevant).toBeGreaterThan(0);
    },
    2 * NET_TIMEOUT
  );

  test(
    "scene:find lists every quality and resolve returns a playable source",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 8 });
      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.downloads?.length).toBeGreaterThan(0);
      for (const download of scene!.downloads ?? []) {
        expect(download.url).toStartWith("http");
      }

      const resolved = await resolve(items[0].source_url);
      expect(resolved.url).toStartWith("http");
    },
    3 * NET_TIMEOUT
  );
});
