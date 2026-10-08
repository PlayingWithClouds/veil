import { test, expect, describe } from "bun:test";
import { parse } from "node-html-parser";
import { parseCards } from "../src/listing.ts";
import { buildScene } from "../src/scrape.ts";
import { pageDefinitions, pickBest, resolve } from "../src/resolve.ts";
import { extractFlashvars, extractVideoObject } from "../src/page.ts";
import { interleave, sceneList } from "../src/scenes.ts";
import { siteCategories } from "../src/tags.ts";
import { listingUrl, parseClockDuration, parseIsoDuration, slugFromHref, withPageParam } from "../src/http.ts";
import { ALL_SITES, enabledSites, siteForUrl, type Site } from "../src/sites.ts";

function site(key: string): Site {
  return ALL_SITES.find((candidate) => candidate.key === key)!;
}

async function fixture(name: string): Promise<string> {
  return Bun.file(`${import.meta.dir}/fixtures/${name}`).text();
}

// ---------------------------------------------------------------------------
// Pure helpers (no network)

describe("site routing", () => {
  test("idFromUrl reads page and embed urls of each site", () => {
    expect(site("pornhub").idFromUrl("https://www.pornhub.com/view_video.php?viewkey=6aacc98152b5f")).toBe("6aacc98152b5f");
    expect(site("pornhub").idFromUrl("https://www.pornhub.com/embed/ph5e8de50eddb01")).toBe("ph5e8de50eddb01");
    expect(site("pornhub").idFromUrl("https://www.pornhub.com/pornstar/j-mac")).toBe("");
    expect(site("redtube").idFromUrl("https://www.redtube.com/103634591")).toBe("103634591");
    expect(site("redtube").idFromUrl("https://embed.redtube.com/?id=103634591")).toBe("103634591");
    expect(site("youporn").idFromUrl("https://www.youporn.com/watch/205140451/some-slug/")).toBe("205140451");
    expect(site("tube8").idFromUrl("https://www.tube8.com/porn-video/31252491/")).toBe("31252491");
  });

  test("sceneUrl is the canonical page and round-trips through idFromUrl", () => {
    for (const candidate of ALL_SITES) {
      expect(candidate.idFromUrl(candidate.sceneUrl("12345"))).toBe("12345");
    }
  });

  test("siteForUrl matches the domain and its subdomains only", () => {
    expect(siteForUrl("https://de.pornhub.com/video")?.key).toBe("pornhub");
    expect(siteForUrl("https://www.tube8.com/cat/teens/")?.key).toBe("tube8");
    expect(siteForUrl("https://notpornhub.com/")).toBeUndefined();
    expect(siteForUrl("not a url")).toBeUndefined();
  });

  test("enabledSites honours AYLO_SITES and falls back to all four", () => {
    delete process.env.AYLO_SITES;
    expect(enabledSites().length).toBe(4);
    process.env.AYLO_SITES = "redtube, tube8";
    expect(enabledSites().map((candidate) => candidate.key)).toEqual(["redtube", "tube8"]);
    process.env.AYLO_SITES = "nonsense";
    expect(enabledSites().length).toBe(4);
    delete process.env.AYLO_SITES;
  });
});

describe("url and time helpers", () => {
  test("withPageParam sets, replaces and omits the page parameter", () => {
    expect(withPageParam("https://x.com/a?search=teen", 1)).toBe("https://x.com/a?search=teen");
    expect(withPageParam("https://x.com/a?search=teen", 3)).toBe("https://x.com/a?search=teen&page=3");
    expect(withPageParam("https://x.com/a/", 2)).toBe("https://x.com/a/?page=2");
    expect(withPageParam("https://x.com/a?page=2&c=4", 5)).toBe("https://x.com/a?c=4&page=5");
    expect(withPageParam("https://x.com/a?page=2", 1)).toBe("https://x.com/a");
  });

  test("listingUrl points entity roots at their full video list on the canonical host", () => {
    expect(listingUrl(site("pornhub"), "https://de.pornhub.com/pornstar/j-mac")).toBe("https://www.pornhub.com/pornstar/j-mac/videos");
    expect(listingUrl(site("pornhub"), "https://www.pornhub.com/video?c=4")).toBe("https://www.pornhub.com/video?c=4");
    expect(listingUrl(site("youporn"), "https://www.youporn.com/channel/reality-kings/")).toBe("https://www.youporn.com/channel/reality-kings/");
  });

  test("durations parse from clock text and ISO-8601", () => {
    expect(parseClockDuration("12:48")).toBe(768);
    expect(parseClockDuration("\n 05:45\n")).toBe(345);
    expect(parseClockDuration("1:02:03")).toBe(3723);
    expect(parseClockDuration("")).toBe(0);
    expect(parseIsoDuration("PT00H12M48S")).toBe(768);
    expect(parseIsoDuration("PT1217S")).toBe(1217);
    expect(parseIsoDuration(undefined)).toBe(0);
  });

  test("slugFromHref decodes the last path segment", () => {
    expect(slugFromHref("/pornstar/katty+west")).toBe("katty west");
    expect(slugFromHref("/category/teens/")).toBe("teens");
  });

  test("interleave takes one item from each list in turn", () => {
    const item = (id: string) => ({ title: id, media_type: "scene" as const, source_url: id, external_id: id });
    const merged = interleave([[item("a1"), item("a2"), item("a3")], [item("b1")], [item("c1"), item("c2")]]);
    expect(merged.map((entry) => entry.external_id)).toEqual(["a1", "b1", "c1", "a2", "c2", "a3"]);
  });
});

describe("pickBest", () => {
  test("prefers the tallest source, HLS on a tie", () => {
    const best = pickBest([
      { url: "mp4-720", format: "mp4", height: 720 },
      { url: "mp4-1080", format: "mp4", height: 1080 },
      { url: "hls-1080", format: "hls", height: 1080 },
      { url: "hls-480", format: "hls", height: 480 },
    ]);
    expect(best?.url).toBe("hls-1080");
    expect(pickBest([])).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// Parsing against saved pages

describe("listing cards", () => {
  test("pornhub skips header-menu cards and reads title, duration, poster and channel", async () => {
    const cards = parseCards(site("pornhub"), parse(await fixture("ph-listing.html")));
    expect(cards.map((card) => card.external_id)).toEqual([
      "pornhub-6aabd415254f7",
      "pornhub-ph5b2ada45b5a8b",
      "pornhub-ph5a491e7628f29",
    ]);
    expect(cards[1]).toMatchObject({
      media_type: "scene",
      source_url: "https://www.pornhub.com/view_video.php?viewkey=ph5b2ada45b5a8b",
      duration_seconds: 628,
      studio: { name: "Team Skeet", source_url: "https://www.pornhub.com/channels/teamskeet" },
    });
    expect(cards[1].poster_path).toStartWith("https://");
    expect(cards[0].studio).toBeUndefined();
  });

  test("redtube reads numeric ids and amateur uploaders as studios", async () => {
    const cards = parseCards(site("redtube"), parse(await fixture("rt-listing.html")));
    expect(cards.map((card) => card.external_id)).toEqual(["redtube-103634591", "redtube-102262661", "redtube-102089581"]);
    expect(cards[0]).toMatchObject({
      title: "I'm already 18 years old - can you fuck me like a neighbor?",
      source_url: "https://www.redtube.com/103634591",
      duration_seconds: 1217,
      studio: { name: "DisDiger", source_url: "https://www.redtube.com/amateur/disdiger-ph" },
    });
    expect(cards[0].poster_path).toStartWith("https://ei-ph.rdtcdn.com/");
  });

  test("youporn and tube8 share the card markup", async () => {
    const youporn = parseCards(site("youporn"), parse(await fixture("yp-listing.html")));
    expect(youporn[0]).toMatchObject({
      external_id: "youporn-205140451",
      source_url: "https://www.youporn.com/watch/205140451/",
      duration_seconds: 345,
      studio: { name: "Leah Aloe" },
    });
    const tube8 = parseCards(site("tube8"), parse(await fixture("t8-listing.html")));
    expect(tube8[0]).toMatchObject({
      external_id: "tube8-31252491",
      source_url: "https://www.tube8.com/porn-video/31252491/",
      studio: { name: "MassageRooms", source_url: "https://www.tube8.com/channel/massagerooms/" },
    });
    expect(tube8.length).toBe(3);
  });
});

describe("scene pages", () => {
  test("pornhub: player config, performers, tags and related", async () => {
    const html = await fixture("ph-scene.html");
    const scene = buildScene(site("pornhub"), "6aacc98152b5f", html);
    expect(scene).toMatchObject({
      external_id: "pornhub-6aacc98152b5f",
      title: "PAWG Threesome of the Century with Kayley Gunner and Kali Roses by JMac",
      duration_seconds: 768,
      date: "2026-09-18T05:20:10+00:00",
    });
    expect(scene.performers!.map((performer) => performer.name)).toEqual(["J Mac", "Kali Roses", "Kayley Gunner"]);
    expect(scene.performers![0].source_url).toBe("https://www.pornhub.com/pornstar/j-mac");
    // The uploader is a pornstar, who is a performer rather than a studio.
    expect(scene.studio).toBeUndefined();
    expect(scene.tags).toContain("Big Ass");
    expect(scene.tags).toContain("pawg");
    expect(scene.tags).not.toContain("Pornstar");
    expect(scene.related!.map((item) => item.external_id)).toEqual([
      "pornhub-ph5e8de50eddb01",
      "pornhub-ph6135d539747f9",
      "pornhub-6a11f1602d8c3",
    ]);
    expect(scene.related![0].studio?.name).toBe("Wolf Wagner Love");
  });

  test("pornhub: mediaDefinitions come from the flashvars", async () => {
    const html = await fixture("ph-scene.html");
    expect(extractFlashvars(html).video_duration).toBe(768);
    const definitions = pageDefinitions(site("pornhub"), html);
    expect(definitions.filter((definition) => definition.format === "hls").length).toBe(4);
    expect(definitions.find((definition) => definition.format === "mp4").remote).toBe(true);
  });

  test("redtube: amateur uploader, carousel tags, popup-free performer, related", async () => {
    const scene = buildScene(site("redtube"), "103634591", await fixture("rt-scene.html"));
    expect(scene.title).toBe("I'm already 18 years old - can you fuck me like a neighbor?");
    expect(scene.duration_seconds).toBe(1217);
    expect(scene.studio).toMatchObject({ name: "DisDiger", source_url: "https://www.redtube.com/amateur/disdiger-ph" });
    expect(scene.performers).toEqual([
      {
        name: "Katty West",
        external_id: "redtube-ps-katty-west",
        source_url: "https://www.redtube.com/pornstar/katty+west",
      },
    ]);
    expect(scene.tags!.slice(0, 3)).toEqual(["18-25", "HD", "Verified Amateurs"]);
    expect(scene.tags).not.toContain("Порно на русском");
    expect(scene.related!.map((item) => item.external_id)).toEqual(["redtube-40218361", "redtube-43175361", "redtube-41086891"]);
  });

  test("redtube: remote definitions point at the quality-list endpoints", async () => {
    const definitions = pageDefinitions(site("redtube"), await fixture("rt-scene.html"));
    expect(definitions.map((definition) => [definition.format, definition.remote])).toEqual([
      ["hls", true],
      ["mp4", true],
    ]);
    expect(definitions[0].videoUrl).toStartWith("/media/hls?s=");
  });

  test("youporn: channel studio, performer chips are not tags", async () => {
    const scene = buildScene(site("youporn"), "214140131", await fixture("yp-scene.html"));
    expect(scene.studio).toMatchObject({ name: "RealityKings", source_url: "https://www.youporn.com/channel/reality-kings/" });
    expect(scene.performers!.map((performer) => performer.name)).toEqual(["Logan Pierce", "Staci Silverstone", "Tanya Tate"]);
    expect(scene.tags).toContain("Threesome");
    expect(scene.tags).not.toContain("Logan Pierce");
    expect(scene.related!.length).toBe(3);
    expect(scene.related![0].external_id).toBe("youporn-190138441");
  });

  test("tube8: channel studio and tags", async () => {
    const scene = buildScene(site("tube8"), "31252491", await fixture("t8-scene.html"));
    expect(scene).toMatchObject({
      external_id: "tube8-31252491",
      source_url: "https://www.tube8.com/porn-video/31252491/",
      duration_seconds: 763,
      date: "2016-07-23T17:18:01+00:00",
    });
    expect(scene.studio?.name).toBe("MassageRooms");
    expect(scene.tags).toContain("Massage");
    expect(scene.related!.length).toBe(3);
  });

  test("extractVideoObject finds the VideoObject flat or inside an @graph", async () => {
    expect(extractVideoObject(await fixture("rt-scene.html"))?.duration).toBe("PT1217S");
    expect(extractVideoObject(await fixture("yp-scene.html"))?.duration).toBe("PT418S");
    expect(extractVideoObject("<html></html>")).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// Live (network)

describe("scene:list (live)", () => {
  test("search interleaves all four sites", async () => {
    delete process.env.AYLO_SITES;
    const result = await sceneList({ query: "teen", limit: 12 });
    expect(result.items.length).toBe(12);
    const prefixes = new Set(result.items.map((item) => item.external_id.split("-")[0]));
    expect(prefixes.size).toBe(4);
  }, 60_000);

  for (const candidate of ALL_SITES) {
    test(`${candidate.key}: newest feed paginates without duplicates`, async () => {
      process.env.AYLO_SITES = candidate.key;
      const result = await sceneList({ limit: 45 });
      delete process.env.AYLO_SITES;
      expect(result.items.length).toBe(45);
      expect(new Set(result.items.map((item) => item.external_id)).size).toBe(45);
      expect(result.items[0].duration_seconds).toBeGreaterThan(0);
    }, 60_000);
  }

  test("scene:list:page lists a category page", async () => {
    const result = await sceneList({ url: "https://www.youporn.com/category/teens/", limit: 10 });
    expect(result.items.length).toBe(10);
    expect(result.items[0].external_id).toStartWith("youporn-");
  }, 60_000);
});

describe("categories (live)", () => {
  for (const candidate of ALL_SITES) {
    test(`${candidate.key} lists categories`, async () => {
      const categories = await siteCategories(candidate);
      expect(categories.length).toBeGreaterThan(30);
      expect(categories.every((category) => category.media_type === "tag" && category.title.length > 0)).toBe(true);
    }, 30_000);
  }
});

describe("resolve (live)", () => {
  for (const candidate of ALL_SITES) {
    test(`${candidate.key} resolves a listed scene to a stream`, async () => {
      process.env.AYLO_SITES = candidate.key;
      const { items } = await sceneList({ limit: 1 });
      delete process.env.AYLO_SITES;
      const resolved = await resolve(items[0].source_url);
      expect(resolved.url).toStartWith("https://");
      expect(["application/x-mpegURL", "video/mp4"]).toContain(resolved.mime_type);
      expect(resolved.headers?.Referer).toBe(`https://${candidate.host}/`);
      // A ranged request keeps the check cheap when the source is an MP4 file.
      const stream = await fetch(resolved.url, { headers: { ...resolved.headers, Range: "bytes=0-99" } });
      expect(stream.ok).toBe(true);
    }, 60_000);
  }
});
