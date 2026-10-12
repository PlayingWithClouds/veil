import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { parseVideoCards } from "../src/cards.ts";
import { listingUrl, searchPathFromQuery, sceneList } from "../src/listing.ts";
import { sceneFromWatchPage, scrapeScene } from "../src/scene.ts";
import { bestSource, resolve } from "../src/resolve.ts";
import { suggest, suggestionsFromBody } from "../src/suggest.ts";
import { decodeEntities, videoIdFromUrl } from "../src/http.ts";
import { clockSeconds, countFromText } from "../src/values.ts";

const NET_TIMEOUT = 60_000;
// The live tests go through FlareSolverr, so they run only when it is configured.
const hasSolver = Boolean(process.env.FLARESOLVERR_URL);

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("videoIdFromUrl reads the owner_video id of a watch URL", () => {
    expect(videoIdFromUrl("https://noodlemagazine.com/watch/-152603790_456239053")).toBe("-152603790_456239053");
    expect(videoIdFromUrl("https://noodlemagazine.com/video/creampie")).toBe("");
  });

  test("search paths join the words with +, and listings take 0-based pages", () => {
    expect(searchPathFromQuery("woodman  casting")).toBe("/video/woodman+casting");
    expect(listingUrl({ query: "woodman casting" }, 1)).toBe("https://noodlemagazine.com/video/woodman+casting?p=1");
    expect(listingUrl({}, 0)).toBe("https://noodlemagazine.com/home?range=month&p=0");
    expect(listingUrl({ url: "https://noodlemagazine.com/video/creampie" }, 2)).toBe(
      "https://noodlemagazine.com/video/creampie?p=2"
    );
  });

  test("clock and view-count texts", () => {
    expect(clockSeconds("13:40")).toBe(820);
    expect(clockSeconds("1:02:03")).toBe(3723);
    expect(clockSeconds("soon")).toBe(0);
    expect(countFromText("685.92K")).toBe(685920);
    expect(countFromText("1.2M")).toBe(1200000);
    expect(countFromText("lots")).toBeUndefined();
  });

  test("decodeEntities handles named, numeric and double-encoded entities", () => {
    expect(decodeEntities("a &amp; b &quot;c&quot; &#39;d&#39; &amp;lt;")).toBe("a & b \"c\" 'd' &lt;");
  });
});

describe("listing pages (offline)", () => {
  test("the monthly top list has 24 videos with identity, poster, runtime and views", () => {
    const items = parseVideoCards(parse(loadFixture("home.html")));
    expect(items.length).toBe(24);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(24);
    const first = items[0];
    expect(first.media_type).toBe("scene");
    expect(first.external_id).toBe("noodlemagazine--152603790_456239053");
    expect(first.source_url).toBe("https://noodlemagazine.com/watch/-152603790_456239053");
    expect(first.poster_path).toBe("https://cdn2.pvvstream.pro/videos/-152603790/456239053/preview_320.jpg");
    expect(first.duration_seconds).toBe(820);
    expect(first.view_count).toBe(685920);
  });

  test("the second page continues with other videos", () => {
    const firstIds = new Set(parseVideoCards(parse(loadFixture("home.html"))).map((item) => item.external_id));
    const second = parseVideoCards(parse(loadFixture("home-page2.html")));
    expect(second.length).toBe(24);
    expect(second.some((item) => firstIds.has(item.external_id))).toBe(false);
  });

  test("a lazily loaded poster takes its image from data-src, not the gif placeholder", () => {
    const items = parseVideoCards(parse(loadFixture("search-woodman-casting.html")));
    const lazy = items.find((item) => item.title.startsWith("Cutie came to a porn casting"));
    expect(lazy?.poster_path).toBe("https://cdn2.pvvstream.pro/videos/-132440213/456239334/preview_320.jpg");
    for (const item of items) expect(item.poster_path?.startsWith("data:")).toBe(false);
  });

  test("a search page lists the videos that match the words", () => {
    const items = parseVideoCards(parse(loadFixture("search-woodman-casting.html")));
    expect(items.length).toBe(24);
    for (const item of items) expect(item.title.toLowerCase()).toContain("woodman");
  });
});

describe("watch page (offline)", () => {
  const watch = parse(loadFixture("watch.html"));
  const scene = sceneFromWatchPage(watch, "-152603790_456239053");

  test("the scene carries the page's metadata", () => {
    expect(scene.type).toBe("scene");
    expect(scene.external_id).toBe("noodlemagazine--152603790_456239053");
    expect(scene.source_url).toBe("https://noodlemagazine.com/watch/-152603790_456239053");
    expect(scene.title).toStartWith("Реальный русский инцест");
    expect(scene.date).toBe("2020-07-04");
    expect(scene.duration_seconds).toBe(820);
    expect(scene.view_count).toBe(35757453);
    expect(scene.poster_path).toBe("https://cdn2.pvvstream.pro/videos/-152603790/456239053/preview_800.jpg");
    expect(scene.tags).toContain("incest");
  });

  test("related videos are the page's other videos, without the scene itself", () => {
    expect(scene.related?.length).toBe(24);
    for (const item of scene.related ?? []) {
      expect(item.external_id).not.toBe(scene.external_id);
    }
  });

  test("the one download is the watch page: the signed links expire, so resolve fetches one at play time", () => {
    expect(scene.downloads?.map((download) => download.url)).toEqual([scene.source_url]);
  });
});

describe("stream resolution (offline)", () => {
  test("the highest quality of the player's playlist wins", () => {
    const source = bestSource(loadFixture("watch.html"));
    expect(source?.height).toBe(720);
    expect(source?.url).toStartWith("https://cdn.pvvstream.pro/videos/-152603790/456239053/vid_720p.mp4");
  });

  test("a page without a playlist falls back to its video element", () => {
    const html = '<video src="https://cdn.example.com/a/vid_360p.mp4?x=1&amp;y=2"></video>';
    expect(bestSource(html)?.url).toBe("https://cdn.example.com/a/vid_360p.mp4?x=1&y=2");
  });
});

describe("search suggestions (offline)", () => {
  test("the autocomplete reply, even inside the solver's <pre> block, gives its results as queries", () => {
    const suggestions = suggestionsFromBody(loadFixture("autocomplete-teen.html"));
    expect(suggestions[0]).toEqual({ text: "teen", kind: "query" });
    expect(suggestions.length).toBe(10);
  });
});

// ---------------------------------------------------------------------------
// Live integration against noodlemagazine.com (through FlareSolverr).

describe.if(hasSolver)("scene:list (live, needs FlareSolverr)", () => {
  test(
    "returns a bounded page of video stubs",
    async () => {
      const { items } = await sceneList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) expect(item.external_id).toStartWith("noodlemagazine-");
    },
    2 * NET_TIMEOUT
  );

  test(
    "offsets past the first page do not repeat it",
    async () => {
      const first = await sceneList({ limit: 5 });
      const later = await sceneList({ limit: 5, offset: 24 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(later.items.length).toBe(5);
      expect(later.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test(
    "a query returns videos that match it",
    async () => {
      const { items } = await sceneList({ query: "woodman casting", limit: 5 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.every((item) => item.title.toLowerCase().includes("woodman"))).toBe(true);
    },
    2 * NET_TIMEOUT
  );
});

describe.if(hasSolver)("scene:find, stream:resolve and search:suggest (live, needs FlareSolverr)", () => {
  test(
    "a listed video scrapes to a scene and resolves to an mp4",
    async () => {
      const { items } = await sceneList({ limit: 1 });
      const [result] = await scrapeScene(items[0].source_url);
      expect(result.type).toBe("scene");
      if (result.type !== "scene") throw new Error("expected a scene");
      expect(result.scene.external_id).toBe(items[0].external_id);
      const resolved = await resolve(items[0].source_url);
      expect(resolved.mime_type).toBe("video/mp4");
      expect(resolved.url).toContain("pvvstream.pro");
    },
    3 * NET_TIMEOUT
  );

  test(
    "suggestions come back for a partly typed query",
    async () => {
      const { suggestions } = await suggest({ query: "teen", limit: 5 });
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions.length).toBeLessThanOrEqual(5);
    },
    2 * NET_TIMEOUT
  );
});
