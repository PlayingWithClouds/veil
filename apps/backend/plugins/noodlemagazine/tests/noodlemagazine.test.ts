import { test, expect, describe, afterEach } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { parseVideoCards } from "../src/cards.ts";
import { listingPath, searchPathFromQuery, sceneList } from "../src/listing.ts";
import { pageTags, sceneFromWatchPage, scrapeScene } from "../src/scene.ts";
import { bestSource, chooseSource, requestedHeight, resolve } from "../src/resolve.ts";
import { suggest, suggestionsFromBody } from "../src/suggest.ts";
import { decodeEntities, videoIdFromUrl } from "../src/http.ts";
import { fetchFromMirrors, mirrorPath, usableMirrors } from "../src/mirrors.ts";
import { clockSeconds, countFromText } from "../src/values.ts";

const NET_TIMEOUT = 60_000;
// The copies behind Cloudflare need FlareSolverr; mat6tube.com does not.
const hasSolver = Boolean(process.env.FLARESOLVERR_URL);
// Live tests hit the sites; run them with LIVE=1.
const live = process.env.LIVE === "1";

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("videoIdFromUrl reads the owner_video id of a watch URL on any mirror, or a relative one", () => {
    expect(videoIdFromUrl("https://noodlemagazine.com/watch/-152603790_456239053")).toBe("-152603790_456239053");
    expect(videoIdFromUrl("https://ukdevilz.com/watch/-152603790_456239053")).toBe("-152603790_456239053");
    expect(videoIdFromUrl("https://mat6tube.com/watch/-113656837_456242801#720p")).toBe("-113656837_456242801");
    expect(videoIdFromUrl("/watch/-45836141_456239627")).toBe("-45836141_456239627");
    expect(videoIdFromUrl("https://noodlemagazine.com/video/creampie")).toBe("");
    expect(videoIdFromUrl("https://example.com/watch/-1_2")).toBe("");
  });

  test("search paths join the words with +, and listings take 0-based pages", () => {
    expect(searchPathFromQuery("woodman  casting")).toBe("/video/woodman+casting");
    expect(listingPath({ query: "woodman casting" }, 1)).toBe("/video/woodman+casting?p=1");
    expect(listingPath({}, 0)).toBe("/home?range=month&p=0");
  });

  test("a site page on any mirror lists by its path", () => {
    expect(listingPath({ url: "https://noodlemagazine.com/video/creampie" }, 2)).toBe("/video/creampie?p=2");
    expect(listingPath({ url: "https://mat6tube.com/video/creampie?p=0" }, 1)).toBe("/video/creampie?p=1");
    expect(() => listingPath({ url: "https://example.com/video/creampie" }, 0)).toThrow();
  });

  test("mirrorPath accepts the mirrors' hosts only", () => {
    expect(mirrorPath("https://www.ukdevilz.com/watch/-1_2?x=1")).toBe("/watch/-1_2?x=1");
    expect(mirrorPath("https://noodlemagazine.com.evil.test/watch/-1_2")).toBeUndefined();
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

describe("mirrors", () => {
  const originalFetch = globalThis.fetch;
  const originalSolver = process.env.FLARESOLVERR_URL;
  afterEach(() => {
    globalThis.fetch = originalFetch;
    if (originalSolver === undefined) delete process.env.FLARESOLVERR_URL;
    else process.env.FLARESOLVERR_URL = originalSolver;
  });

  /** Answers plain requests per host; FlareSolverr requests with a solved page of the host they ask for. */
  function stubSites(plain: Record<string, { status: number; body: string }>, solved: Record<string, string>): string[] {
    const requested: string[] = [];
    globalThis.fetch = (async (input: string | URL | Request, init?: RequestInit) => {
      const url = String(input);
      if (url.startsWith("http://solver.test")) {
        const target = new URL((JSON.parse(String(init?.body)) as { url: string }).url);
        requested.push(`solver ${target.hostname}`);
        const response = solved[target.hostname];
        if (response === undefined) return new Response(JSON.stringify({ status: "error", message: "timeout" }));
        return new Response(JSON.stringify({ status: "ok", solution: { response } }));
      }
      const host = new URL(url).hostname;
      requested.push(host);
      const answer = plain[host] ?? { status: 503, body: "" };
      return new Response(answer.body, { status: answer.status });
    }) as typeof fetch;
    return requested;
  }

  const sitePage = (marker: string) => `<html><link href="/static/extend/dark/a.css">${marker}</html>`;
  const challenge = { status: 403, body: "<html>Just a moment...</html>" };

  test("without a solver only the copy that answers plain requests is used", async () => {
    delete process.env.FLARESOLVERR_URL;
    expect(usableMirrors(false).map((mirror) => mirror.origin)).toEqual(["https://mat6tube.com"]);
    const requested = stubSites({ "mat6tube.com": { status: 200, body: sitePage("mat6tube") } }, {});
    const page = await fetchFromMirrors("/home?range=month&p=0");
    expect(page.origin).toBe("https://mat6tube.com");
    expect(requested).toEqual(["mat6tube.com"]);
  });

  test("with a solver the first copy wins, and a failing one passes to the next", async () => {
    process.env.FLARESOLVERR_URL = "http://solver.test/v1";
    const requested = stubSites(
      { "noodlemagazine.com": challenge, "ukdevilz.com": challenge, "mat6tube.com": { status: 200, body: sitePage("m") } },
      { "ukdevilz.com": sitePage("ukdevilz") }
    );
    const page = await fetchFromMirrors("/home?range=month&p=0");
    expect(page.origin).toBe("https://ukdevilz.com");
    expect(page.body).toContain("ukdevilz");
    expect(requested).toEqual(["noodlemagazine.com", "solver noodlemagazine.com", "ukdevilz.com", "solver ukdevilz.com"]);
  });

  test("a page that is not the site's (or fails the check) is passed over", async () => {
    process.env.FLARESOLVERR_URL = "http://solver.test/v1";
    stubSites(
      { "mat6tube.com": { status: 200, body: sitePage('<meta property="og:title" content="x">window.playlist') } },
      { "noodlemagazine.com": "<html>blocked</html>", "ukdevilz.com": sitePage("no video here") }
    );
    const page = await fetchFromMirrors("/watch/-1_2", (body) => body.includes("window.playlist"));
    expect(page.origin).toBe("https://mat6tube.com");
  });

  test("when no copy answers, the error names each one", async () => {
    delete process.env.FLARESOLVERR_URL;
    stubSites({ "mat6tube.com": { status: 500, body: "" } }, {});
    await expect(fetchFromMirrors("/home")).rejects.toThrow(/mat6tube\.com: HTTP 500/);
  });
});

describe("listing pages (offline)", () => {
  test("the monthly top list has 24 videos with identity, poster, preview, runtime and views", () => {
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
    expect(items.filter((item) => item.preview_video?.includes("tr_240p.mp4")).length).toBeGreaterThan(0);
  });

  test("a copy's page gives the same identities, on the canonical origin", () => {
    const items = parseVideoCards(parse(loadFixture("mat6tube-recent.html")));
    expect(items.length).toBe(24);
    for (const item of items) {
      expect(item.source_url).toStartWith("https://noodlemagazine.com/watch/");
      expect(item.external_id).toStartWith("noodlemagazine-");
      expect(item.poster_path).toStartWith("https://");
    }
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
  });

  test("tags use the English translation of the tag links", () => {
    expect(scene.tags).toContain("porn");
    expect(scene.tags).toContain("amateur home");
    expect(scene.tags?.some((tag) => /[а-я]/i.test(tag))).toBe(false);
  });

  test("without tag links the video:tag meta list is used", () => {
    const page = parse('<meta property="video:tag" content="teen, anal ,  ">');
    expect(pageTags(page)).toEqual(["teen", "anal"]);
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

  test("a copy's watch page gives the same scene identity and English tags", () => {
    const copy = sceneFromWatchPage(parse(loadFixture("mat6tube-watch.html")), "-113656837_456242801");
    expect(copy.external_id).toBe("noodlemagazine--113656837_456242801");
    expect(copy.source_url).toBe("https://noodlemagazine.com/watch/-113656837_456242801");
    expect(copy.title).toStartWith("Добро пожаловать домой");
    expect(copy.duration_seconds).toBe(2567);
    expect(copy.tags).toContain("stepmom");
    expect(copy.tags).toContain("milf");
    expect(copy.related?.length).toBe(24);
  });
});

describe("stream resolution (offline)", () => {
  test("the highest quality of the player's playlist wins", () => {
    const source = bestSource(loadFixture("watch.html"));
    expect(source?.height).toBe(720);
    expect(source?.url).toStartWith("https://cdn.pvvstream.pro/videos/-152603790/456239053/vid_720p.mp4");
  });

  test("a quality named in the URL's fragment is picked when the page has it", () => {
    const html = loadFixture("mat6tube-watch.html");
    expect(requestedHeight("https://mat6tube.com/watch/-113656837_456242801#480p")).toBe(480);
    expect(requestedHeight("https://noodlemagazine.com/watch/-113656837_456242801")).toBe(0);
    expect(chooseSource(html, 480)?.url).toContain("vid_480p.mp4");
    expect(chooseSource(html, 1080)?.height).toBe(720);
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
// Live integration: through the first copy that answers (mat6tube.com without
// FlareSolverr). Run with LIVE=1.

describe.if(live)("live, through the mirrors", () => {
  test(
    "lists, pages and searches videos",
    async () => {
      const first = await sceneList({ limit: 5 });
      expect(first.items.length).toBe(5);
      for (const item of first.items) expect(item.external_id).toStartWith("noodlemagazine-");
      const later = await sceneList({ limit: 5, offset: 24 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(later.items.some((item) => firstIds.has(item.external_id))).toBe(false);
      const { items } = await sceneList({ query: "woodman casting", limit: 5 });
      expect(items.every((item) => item.title.toLowerCase().includes("woodman"))).toBe(true);
    },
    3 * NET_TIMEOUT
  );

  test(
    "a listed video scrapes to a scene and resolves to an mp4, also from a copy's URL",
    async () => {
      const { items } = await sceneList({ limit: 1 });
      const [result] = await scrapeScene(items[0].source_url);
      if (result.type !== "scene") throw new Error("expected a scene");
      expect(result.scene.external_id).toBe(items[0].external_id);
      const copyUrl = items[0].source_url.replace("https://noodlemagazine.com", "https://mat6tube.com");
      const resolved = await resolve(copyUrl);
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

test.if(live && hasSolver)(
  "live: with FlareSolverr the first copy, noodlemagazine.com, serves",
  async () => {
    const page = await fetchFromMirrors("/home?range=month&p=0");
    expect(page.origin).toBe("https://noodlemagazine.com");
  },
  2 * NET_TIMEOUT
);
