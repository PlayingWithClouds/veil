import { test, expect, describe, afterEach } from "bun:test";
import { parse } from "node-html-parser";
import type { Scene } from "@playingwithclouds/veil-sdk";
import { parseListingCards } from "../src/listing.ts";
import { scrape } from "../src/scrape.ts";
import { enabledSites, siteByKey, siteForUrl, type SiteConfig } from "../src/sites.ts";
import { videoKey } from "../src/urls.ts";

// scatporno.org: a KVS site with its own theme, slug-keyed video pages and
// mostly download-only videos; pages saved from the live site.

const originalFetch = globalThis.fetch;
const originalSelection = process.env.KVS_SITES;

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalSelection === undefined) delete process.env.KVS_SITES;
  else process.env.KVS_SITES = originalSelection;
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

describe("scatporno", () => {
  test("is left out of search and browse until KVS_SITES names it", () => {
    delete process.env.KVS_SITES;
    expect(enabledSites().some((entry) => entry.key === "scatporno")).toBe(false);
    expect(enabledSites().length).toBeGreaterThan(0);
    process.env.KVS_SITES = "scatporno";
    expect(enabledSites().map((entry) => entry.key)).toEqual(["scatporno"]);
  });

  test("video pages are keyed by their slug, with or without .html", () => {
    expect(siteForUrl("https://scatporno.org/porno-scat/bbw-almost-vomits")?.key).toBe("scatporno");
    expect(videoKey(site("scatporno"), "https://scatporno.org/porno-scat/bbw-almost-vomits")).toBe("bbw-almost-vomits");
    expect(videoKey(site("scatporno"), "https://scatporno.org/porno-scat/bbw-almost-vomits/")).toBe("bbw-almost-vomits");
    expect(videoKey(site("scatporno"), "https://scatporno.org/porno-scat/18740-ninounini-cuckold.html")).toBe(
      "18740-ninounini-cuckold"
    );
    expect(videoKey(site("scatporno"), "https://scatporno.org/models/ninounini/")).toBe("");
    expect(videoKey(site("scatporno"), "https://scatporno.org/latest-updates/2/")).toBe("");
  });

  test("listing cards read the slug id, poster and runtime", async () => {
    const items = parseListingCards(site("scatporno"), parse(await fixture("scatporno-listing")));
    expect(items).toHaveLength(40);
    expect(items[0]).toMatchObject({
      media_type: "scene",
      title: "Fuck My Ass and Scat P1",
      external_id: "kvs-scatporno-fuck-my-ass-and-scat-p1",
      source_url: "https://scatporno.org/porno-scat/fuck-my-ass-and-scat-p1",
      poster_path: "https://scatporno.org/contents/videos_screenshots/23000/23972/430x242/1.jpg",
      duration_seconds: 662,
    });
  });

  test("a streamable page yields title, runtime, tags, download and related videos", async () => {
    const url = "https://scatporno.org/porno-scat/bbw-almost-vomits";
    const scene = await scrapeFixture("scatporno-video", url);
    expect(scene).toMatchObject({
      external_id: "kvs-scatporno-bbw-almost-vomits",
      source_url: url,
      title: "BBW ALMOST VOMITS",
      date: "2026-02-22",
      duration_seconds: 323,
      poster_path: "https://scatporno.org/contents/videos_screenshots/19000/19368/preview.jpg",
      studio: { name: "ScatPorno", source_url: "https://scatporno.org" },
      tags: ["bbw", "vomiting"],
    });
    expect(scene.downloads).toEqual([{ label: "mp4", url, format: "mp4" }]);
    expect(scene.related).toHaveLength(20);
    expect(scene.related!.every((item) => item.external_id.startsWith("kvs-scatporno-"))).toBe(true);
  });

  test("a download-only page names its model and tags but offers no stream", async () => {
    const url = "https://scatporno.org/porno-scat/18740-ninounini-cuckold-f-machine-my-lover-fucks-me-in-front-of-my-cuckold-husband.html";
    const scene = await scrapeFixture("scatporno-video-model", url);
    expect(scene.external_id).toBe(
      "kvs-scatporno-18740-ninounini-cuckold-f-machine-my-lover-fucks-me-in-front-of-my-cuckold-husband"
    );
    expect(scene.performers).toEqual([
      {
        name: "Ninounini",
        order: 0,
        external_id: "kvs-scatporno-model-ninounini",
        source_url: "https://scatporno.org/models/ninounini/",
      },
    ]);
    expect(scene.tags!.slice(0, 3)).toEqual(["retro scat", "dirty rosebut", "night scat"]);
    expect(scene.duration_seconds).toBe(1562);
    expect(scene.downloads).toEqual([]);
    expect(scene.related!.length).toBeGreaterThan(0);
  });
});
