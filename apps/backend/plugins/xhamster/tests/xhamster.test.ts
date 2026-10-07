import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  videoIdFromUrl,
  isVideoUrl,
  providerFromUrl,
  minutesFromSeconds,
  durationSecondsFromLabel,
  decodeEntities,
  BASE_URL,
} from "../src/http.ts";
import { sceneList } from "../src/scenes.ts";
import { scrape } from "../src/scrape.ts";

// ---------------------------------------------------------------------------
// Pure URL/slug helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/newest")).toBe(`${BASE_URL}/newest`);
    expect(resolveUrl("newest")).toBe(`${BASE_URL}/newest`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("videoIdFromUrl extracts the trailing id token", () => {
    expect(videoIdFromUrl("/videos/some-clip-title-xhoyFVR")).toBe("xhoyFVR");
    expect(videoIdFromUrl(`${BASE_URL}/videos/two-cougars-on-the-prowl-xhhfMIn`)).toBe("xhhfMIn");
    expect(videoIdFromUrl("/videos/old-style-1234567")).toBe("1234567");
    expect(videoIdFromUrl("/categories/asian")).toBe("");
  });

  test("isVideoUrl detects video detail pages", () => {
    expect(isVideoUrl("/videos/foo-bar-xh123")).toBe(true);
    expect(isVideoUrl("/newest")).toBe(false);
    expect(isVideoUrl("/categories/asian")).toBe(false);
  });

  test("providerFromUrl returns the bare host label", () => {
    expect(providerFromUrl("https://xhamster.com/videos/abc")).toBe("xhamster");
    expect(providerFromUrl("not a url")).toBe("unknown");
  });

  test("minutesFromSeconds rounds to whole minutes", () => {
    expect(minutesFromSeconds(465)).toBe(8);
    expect(minutesFromSeconds(60)).toBe(1);
    expect(minutesFromSeconds(0)).toBe(0);
    expect(minutesFromSeconds(-5)).toBe(0);
  });

  test("durationSecondsFromLabel parses card runtime badges", () => {
    expect(durationSecondsFromLabel("01:36")).toBe(96);
    expect(durationSecondsFromLabel("8:23")).toBe(8 * 60 + 23);
    expect(durationSecondsFromLabel("1:02:03")).toBe(3600 + 2 * 60 + 3);
    expect(durationSecondsFromLabel(" 12:00 ")).toBe(12 * 60);
    expect(durationSecondsFromLabel("nonsense")).toBe(0);
    expect(durationSecondsFromLabel("")).toBe(0);
  });

  test("decodeEntities unescapes common HTML entities", () => {
    expect(decodeEntities("Kasumi&#039;s Bath &amp; Handjob")).toBe("Kasumi's Bath & Handjob");
    expect(decodeEntities("A &quot;quote&quot; &lt;here&gt;")).toBe('A "quote" <here>');
    expect(decodeEntities("Ayano&#8217;s Scene")).toBe("Ayano’s Scene");
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against xhamster.com. Chained so they survive
// catalog changes (scrape a title that scene:list actually returned).

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
        expect(item.source_url).toContain("/videos/");
        expect(item.external_id).toStartWith("xhamster-");
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
        expect(item.source_url).toContain("/videos/");
        expect(item.external_id).toStartWith("xhamster-");
        expect(item.media_type).toBe("scene");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape (live)", () => {
  test(
    "scrapes a real video surfaced by scene:list into a scene",
    async () => {
      const { items } = await sceneList({ query: QUERY, limit: 20 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const scene = results.flatMap((r) => (r.type === "scene" ? [r.scene] : []))[0];

      expect(scene).toBeDefined();
      expect(scene!.title.trim()).not.toBe("");
      expect(scene!.external_id).toStartWith("xhamster-");
      expect(scene!.source_url).toContain("/videos/");
      expect(scene!.duration_seconds ?? 0).toBeGreaterThan(0);

      const urls = (scene!.downloads ?? []).map((d) => d.url);
      for (const url of urls) expect(url).toStartWith("http");
      expect(new Set(urls).size).toBe(urls.length);
    },
    3 * NET_TIMEOUT
  );
});
