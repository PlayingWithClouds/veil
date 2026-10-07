import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  videoCodeFromUrl,
  isVideoUrl,
  providerFromUrl,
  yearFromDate,
  parseClockDuration,
  decodeEntities,
  BASE_URL,
} from "../src/http.ts";
import { scrape } from "../src/scrape.ts";
import { sceneList } from "../src/scenes.ts";
import { resolve } from "../src/resolve.ts";

// ---------------------------------------------------------------------------
// Pure URL/slug helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/en/new")).toBe(`${BASE_URL}/en/new`);
    expect(resolveUrl("en/new")).toBe(`${BASE_URL}/en/new`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("videoCodeFromUrl extracts the trailing code segment", () => {
    expect(videoCodeFromUrl("/en/ebon-006")).toBe("ebon-006");
    expect(videoCodeFromUrl(`${BASE_URL}/dm539/en/mfcl-030`)).toBe("mfcl-030");
    expect(videoCodeFromUrl("https://missav.ws/en/fc2-ppv-1234567")).toBe("fc2-ppv-1234567");
  });

  test("isVideoUrl separates videos from section pages", () => {
    expect(isVideoUrl("/en/ebon-006")).toBe(true);
    expect(isVideoUrl("/dm33/en/mfcl-030")).toBe(true);
    expect(isVideoUrl("/en/genres")).toBe(false);
    expect(isVideoUrl("/en/actresses")).toBe(false);
  });

  test("providerFromUrl returns the bare host label", () => {
    expect(providerFromUrl("https://surrit.com/abc/playlist.m3u8")).toBe("surrit");
    expect(providerFromUrl("not a url")).toBe("unknown");
  });

  test("yearFromDate reads the leading year", () => {
    expect(yearFromDate("2026-07-02")).toBe(2026);
    expect(yearFromDate("1999-01-01")).toBe(1999);
    expect(yearFromDate("")).toBeUndefined();
  });

  test("parseClockDuration handles H:MM:SS and MM:SS", () => {
    expect(parseClockDuration("2:07:59")).toBe(2 * 3600 + 7 * 60 + 59);
    expect(parseClockDuration("59:01")).toBe(59 * 60 + 1);
    expect(parseClockDuration(" 3:05 ")).toBe(3 * 60 + 5);
    expect(parseClockDuration("HD")).toBe(0);
    expect(parseClockDuration("")).toBe(0);
  });

  test("decodeEntities unescapes common HTML entities", () => {
    expect(decodeEntities("friend&#039;s body &amp; more")).toBe("friend's body & more");
    expect(decodeEntities("&quot;Someday&quot;")).toBe('"Someday"');
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against missav.ws. Chained so they survive
// catalog changes (scrape a title that scene:list actually returned).

const NET_TIMEOUT = 30_000;

describe("scene:list (live)", () => {
  test(
    "returns a bounded set of catalog stubs with durations",
    async () => {
      const { items } = await sceneList({ limit: 6 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(6);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("/en/");
        expect(item.external_id).toStartWith("missav-");
        expect(item.media_type).toBe("scene");
      }
      expect(items.some((item) => (item.duration_seconds ?? 0) > 0)).toBe(true);
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape (live)", () => {
  test(
    "scrapes a real video surfaced by scene:list",
    async () => {
      const { items } = await sceneList({ limit: 6 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.external_id).toStartWith("missav-");
      expect(scene!.source_url).toContain("/en/");

      const urls = (scene!.downloads ?? []).map((d) => d.url);
      for (const url of urls) expect(url).toStartWith("http");
    },
    3 * NET_TIMEOUT
  );
});

// missav's search route is Cloudflare-walled, so this only runs when a solver
// is configured (FLARESOLVERR_URL). Skipped otherwise to keep offline runs green.
const hasSolver = Boolean(process.env.FLARESOLVERR_URL);

describe.if(hasSolver)("scene:list with query (live, via solver)", () => {
  test(
    "returns matching videos for a code query",
    async () => {
      const { items } = await sceneList({ query: "cawd", limit: 8 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(8);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.external_id).toStartWith("missav-");
        expect(item.source_url).toContain("/en/");
        expect(item.media_type).toBe("scene");
      }
    },
    3 * NET_TIMEOUT
  );
});

describe("resolve (live)", () => {
  test(
    "unpacks the player to a surrit HLS master",
    async () => {
      const { items } = await sceneList({ limit: 6 });
      const resolved = await resolve(items[0].source_url);
      expect(resolved.url).toContain("surrit.com");
      expect(resolved.url).toContain(".m3u8");
      expect(resolved.mime_type).toBe("application/x-mpegURL");
    },
    3 * NET_TIMEOUT
  );
});
