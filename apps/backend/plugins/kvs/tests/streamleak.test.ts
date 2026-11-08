import { test, expect, describe, afterEach } from "bun:test";
import { parse } from "node-html-parser";
import type { Scene } from "@playingwithclouds/veil-sdk";
import { parseListingCards } from "../src/listing.ts";
import { scrape } from "../src/scrape.ts";
import { siteByKey, siteForUrl, type SiteConfig } from "../src/sites.ts";
import { searchUrl, videoKey } from "../src/urls.ts";

// stream-leak.com and camgirlsleak.com: two catalogs of Chaturbate/MyFreeCams/
// Stripchat recordings on the stock KVS theme, saved from the live sites.

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** The configured site with this key; fails the test when missing. */
function site(key: string): SiteConfig {
  const found = siteByKey(key);
  if (!found) throw new Error(`no site ${key}`);
  return found;
}

/** A saved page. */
async function fixture(name: string): Promise<string> {
  return Bun.file(`${import.meta.dir}/fixtures/${name}.html`).text();
}

/** Scrapes `url` with every request answered by the saved page `name`. */
async function scrapeFixture(name: string, url: string): Promise<Scene> {
  const html = await fixture(name);
  globalThis.fetch = (async () => new Response(html, { status: 200 })) as unknown as typeof fetch;
  const results = await scrape(url);
  return (results[0] as { scene: Scene }).scene;
}

describe("stream-leak and camgirlsleak", () => {
  test("are two sites of cam recordings, each routed by its own domain", () => {
    expect(siteForUrl("https://stream-leak.com/videos/256673/x/")?.key).toBe("streamleak");
    expect(siteForUrl("https://camgirlsleak.com/videos/124602/x/")?.key).toBe("camgirlsleak");
    expect(site("streamleak").content).toBe("cam-shows");
    expect(site("camgirlsleak").content).toBe("cam-shows");
  });

  test("video and search URLs follow the /videos/<id>/ scheme", () => {
    const url = "https://stream-leak.com/videos/256673/watch-kenziesmithh-best-porn-leak-video/";
    expect(videoKey(site("streamleak"), url)).toBe("256673");
    expect(videoKey(site("streamleak"), "https://stream-leak.com/models/kenziesmithh/")).toBe("");
    expect(searchUrl(site("streamleak"), "kenziesmithh")).toBe("https://stream-leak.com/search/kenziesmithh/");
  });

  test("listing cards read the id, lazy poster and runtime", async () => {
    const items = parseListingCards(site("streamleak"), parse(await fixture("streamleak-listing")));
    expect(items).toHaveLength(24);
    expect(items[0]).toMatchObject({
      media_type: "scene",
      external_id: "kvs-streamleak-261628",
      source_url: "https://stream-leak.com/videos/261628/jully-lov-new-porn-video-chaturbate-daddy-fuckmachine-squirt-bigboobs/",
      poster_path: "https://stream-leak.com/contents/videos_screenshots/261000/261628/320x180/1.jpg",
      duration_seconds: 906,
    });
  });

  test("a stream-leak page yields the scene, linked model, cam site and related videos", async () => {
    const url = "https://stream-leak.com/videos/256673/watch-kenziesmithh-best-porn-leak-video-chaturbate-suck-oil-uncut-tattoo/";
    const scene = await scrapeFixture("streamleak-video", url);
    expect(scene).toMatchObject({
      external_id: "kvs-streamleak-256673",
      source_url: url,
      title: "Watch kenziesmithh Best Porn Leak Video [Chaturbate] - suck, oil, uncut, tattoo",
      date: "2026-01-04",
      duration_seconds: 907,
      poster_path: "https://stream-leak.com/contents/videos_screenshots/256000/256673/preview.jpg",
      studio: { name: "Chaturbate" },
    });
    expect(scene.performers).toEqual([
      {
        name: "kenziesmithh",
        order: 0,
        external_id: "kvs-streamleak-model-kenziesmithh",
        source_url: "https://stream-leak.com/models/kenziesmithh/",
      },
    ]);
    expect(scene.downloads).toEqual([{ label: "mp4", url, format: "mp4" }]);
    expect(scene.related!.length).toBeGreaterThan(0);
    expect(scene.related!.every((item) => item.external_id.startsWith("kvs-streamleak-"))).toBe(true);
    expect(scene.related!.some((item) => item.external_id === scene.external_id)).toBe(false);
  });

  test("a camgirlsleak page names its models although it leaves them unlinked", async () => {
    const url = "https://camgirlsleak.com/videos/124602/ellis-cherry-hd-porn-video-chaturbate-bigass-18-teen-cute-bigboobs/";
    const scene = await scrapeFixture("camgirlsleak-video", url);
    expect(scene.external_id).toBe("kvs-camgirlsleak-124602");
    expect(scene.performers).toEqual([{ name: "ellis_cherry", order: 0 }]);
    expect(scene.studio).toEqual({ name: "Chaturbate" });
    expect(scene.related!.every((item) => item.external_id.startsWith("kvs-camgirlsleak-"))).toBe(true);
  });
});
