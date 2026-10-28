import { test, expect, describe, afterEach } from "bun:test";
import { parse } from "node-html-parser";
import {
  fetchPage,
  listingPageFromUrl,
  listingPageUrl,
  pageUrl,
  parseClockSeconds,
  parseSiteDate,
  playbackIdFromUrl,
  recordingStampFromUrl,
  videoIdFromUrl,
} from "../src/http.ts";
import { parseVideoCards, sceneTitle } from "../src/listing.ts";
import { buildScene, scrape } from "../src/scrape.ts";
import { sceneList } from "../src/scenes.ts";
import { resolve, videoSource } from "../src/resolve.ts";

const modelPage = await Bun.file(`${import.meta.dir}/fixtures/model.html`).text();
const videoPage = await Bun.file(`${import.meta.dir}/fixtures/scene.html`).text();
const VIDEO_URL =
  "https://www.showcamrips.com/show-cam-sex-movies/2210361-cosmictandem-chaturbate-webcam-rip-20261028-210524.html";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Serves `pages` by URL (404 otherwise) and returns the URLs requested, in order. */
function mockFetch(pages: Record<string, string>): string[] {
  const requested: string[] = [];
  globalThis.fetch = (async (input: string | URL | Request) => {
    const url = String(input);
    requested.push(url);
    const body = pages[url];
    if (body === undefined) return new Response("", { status: 404 });
    return new Response(body, { status: 200 });
  }) as typeof fetch;
  return requested;
}

describe("url helpers", () => {
  test("videoIdFromUrl reads the id of a video page only", () => {
    expect(videoIdFromUrl(VIDEO_URL)).toBe("2210361");
    expect(videoIdFromUrl("/show-cam-sex-movies/1-a.html")).toBe("1");
    expect(videoIdFromUrl("https://www.showcamrips.com/model/en/cosmictandem/")).toBe("");
  });

  test("playbackIdFromUrl accepts play, iframe and video page URLs", () => {
    expect(playbackIdFromUrl("https://www.showcamrips.com/play.php?idd=2210361")).toBe("2210361");
    expect(playbackIdFromUrl("https://www.showcamrips.com/loading_video.php?idd=2210361&vv=27")).toBe("2210361");
    expect(playbackIdFromUrl(VIDEO_URL)).toBe("2210361");
    expect(playbackIdFromUrl("https://www.showcamrips.com/en/")).toBe("");
  });

  test("listing pages parse and rebuild", () => {
    expect(listingPageFromUrl("https://www.showcamrips.com/model/en/Valery_Salmons/3-pg/")).toEqual({
      kind: "model",
      name: "Valery_Salmons",
    });
    expect(listingPageFromUrl("https://www.showcamrips.com/site/en/Chaturbate/")).toEqual({
      kind: "site",
      name: "Chaturbate",
    });
    expect(listingPageFromUrl(VIDEO_URL)).toBeUndefined();
    expect(listingPageUrl({ kind: "cat", name: "Couple" })).toBe("https://www.showcamrips.com/cat/en/Couple/");
  });

  test("pageUrl appends <n>-pg/ from the second page", () => {
    expect(pageUrl("https://www.showcamrips.com/en/", 1)).toBe("https://www.showcamrips.com/en/");
    expect(pageUrl("https://www.showcamrips.com/en/", 3)).toBe("https://www.showcamrips.com/en/3-pg/");
  });

  test("clock, date and recording stamp parsing", () => {
    expect(parseClockSeconds("00:13:13")).toBe(793);
    expect(parseClockSeconds("13:13")).toBe(793);
    expect(parseClockSeconds("nonsense")).toBe(0);
    expect(parseSiteDate("2026.10.28_21.05.24 CamRips")).toBe("2026-10-28");
    expect(parseSiteDate("")).toBe("");
    expect(recordingStampFromUrl(VIDEO_URL)).toBe("2026.10.28_21.05.24");
    expect(sceneTitle("cosmictandem", "Chaturbate", VIDEO_URL)).toBe(
      "cosmictandem Chaturbate webcam rip 2026.10.28_21.05.24"
    );
    expect(sceneTitle("", "Chaturbate", VIDEO_URL)).toBe("");
  });
});

describe("listing cards", () => {
  test("a model page yields one stub per card with runtime, date and cam site", () => {
    const items = parseVideoCards(parse(modelPage));
    expect(items.map((item) => item.external_id)).toEqual([
      "showcamrips-2210361",
      "showcamrips-2210333",
      "showcamrips-2209040",
    ]);
    expect(items[0]).toMatchObject({
      title: "cosmictandem Chaturbate webcam rip 2026.10.28_21.05.24",
      media_type: "scene",
      source_url: VIDEO_URL,
      poster_path: "https://www.showpicmoi.com/img/2210361.jpg",
      duration_seconds: 13 * 60 + 13,
      date: "2026-10-28",
      studio: { name: "Chaturbate" },
    });
  });

  test("the placeholder card of an unknown search has no title and is skipped", () => {
    const html = `<ul class="content"><li id="li0">
      <a href="https://www.showcamrips.com/show-cam-sex-movies/748977-.html" sitename="Stripchat" data-id="748977" class="moiclick1" title=""></a>
      <span class="dur">00:09:22</span></li></ul>`;
    expect(parseVideoCards(parse(html))).toEqual([]);
  });
});

describe("video page", () => {
  const scene = buildScene(parse(videoPage), VIDEO_URL);

  test("title, studio, performer, tags and runtime come from the page and URL", () => {
    expect(scene.external_id).toBe("showcamrips-2210361");
    expect(scene.title).toBe("cosmictandem Chaturbate webcam rip 2026.10.28_21.05.24");
    expect(scene.studio).toEqual({ name: "Chaturbate" });
    expect(scene.performers).toEqual([
      {
        name: "cosmictandem",
        order: 0,
        source_url: "https://www.showcamrips.com/model/en/cosmictandem/",
        external_id: "showcamrips-model-cosmictandem",
      },
    ]);
    expect(scene.tags).toEqual(["Live cam", "Couple"]);
    expect(scene.duration_seconds).toBe(793);
    expect(scene.date).toBe("2026-10-28");
    expect(scene.poster_path).toBe("https://www.showpicmoi.com/img/2210361.jpg");
  });

  test("the download is the play page, resolved later", () => {
    expect(scene.downloads).toEqual([
      { label: "mp4", url: "https://www.showcamrips.com/play.php?idd=2210361", format: "mp4" },
    ]);
  });

  test("related lists the Related block, then the model's other shows, without the video itself", () => {
    expect(scene.related!.map((item) => item.external_id)).toEqual([
      "showcamrips-2192361",
      "showcamrips-2186361",
      "showcamrips-2210333",
    ]);
  });

  test("scrape rejects URLs that are not video pages", async () => {
    await expect(scrape("https://www.showcamrips.com/en/")).rejects.toThrow("unsupported");
  });
});

describe("resolve", () => {
  test("videoSource reads the player's mp4 and ignores pages without one", () => {
    const html = '<div><video id="myVideo" controls src="https://vidi.suzavido.top/videos4/1/a_.mp4"></video></div>';
    expect(videoSource(html)).toBe("https://vidi.suzavido.top/videos4/1/a_.mp4");
    expect(videoSource("<html>loading</html>")).toBe("");
  });

  test("resolves the play page with the site referer and returns the playback headers", async () => {
    let seenReferer = "";
    globalThis.fetch = (async (_input: string | URL | Request, init?: RequestInit) => {
      seenReferer = (init?.headers as Record<string, string>).Referer;
      return new Response('<video id="myVideo" controls src="https://vidi.suzavido.top/videos4/1/a_.mp4"></video>');
    }) as typeof fetch;
    const result = await resolve(VIDEO_URL);
    expect(seenReferer).toBe("https://www.showcamrips.com/");
    expect(result.url).toBe("https://vidi.suzavido.top/videos4/1/a_.mp4");
    expect(result.mime_type).toBe("video/mp4");
    expect(result.headers?.Referer).toBe("https://www.showcamrips.com/");
  });

  test("a play page without a video is an error", async () => {
    mockFetch({ "https://www.showcamrips.com/play.php?idd=5": "<html></html>" });
    await expect(resolve("https://www.showcamrips.com/play.php?idd=5")).rejects.toThrow("no video");
  });
});

describe("sceneList", () => {
  test("an empty answer is asked for again", async () => {
    let calls = 0;
    globalThis.fetch = (async () => {
      calls++;
      if (calls < 3) return new Response("", { status: 200 });
      return new Response("<html>ok</html>", { status: 200 });
    }) as typeof fetch;
    expect(await fetchPage("https://www.showcamrips.com/en/")).toBe("<html>ok</html>");
    expect(calls).toBe(3);
  });

  test("a model search continues on the model's own pages", async () => {
    const requested = mockFetch({
      "https://www.showcamrips.com/search.php?Src=cosmictandem&lg=en": modelPage,
      "https://www.showcamrips.com/model/en/cosmictandem/2-pg/": modelPage.replace(/2209040|2210333|2210361/g, "9$&"),
    });
    const result = await sceneList({ query: "cosmictandem", limit: 5 });
    expect(result.items).toHaveLength(5);
    expect(requested).toEqual([
      "https://www.showcamrips.com/search.php?Src=cosmictandem&lg=en",
      "https://www.showcamrips.com/model/en/cosmictandem/2-pg/",
    ]);
  });

  test("a query nobody matches finds nothing", async () => {
    mockFetch({
      "https://www.showcamrips.com/search.php?Src=blonde&lg=en": "<html></html>",
      "https://www.showcamrips.com/en/": "<html><div class='az'></div></html>",
    });
    expect((await sceneList({ query: "blonde" })).items).toEqual([]);
  });

  test("a query naming a cam site lists that site's page", async () => {
    const home = '<div class="az"><a href="https://www.showcamrips.com/site/en/Chaturbate/">Chaturbate</a></div>';
    const requested = mockFetch({
      "https://www.showcamrips.com/search.php?Src=chaturbate&lg=en": "<html></html>",
      "https://www.showcamrips.com/en/": home,
      "https://www.showcamrips.com/site/en/Chaturbate/": modelPage,
    });
    const result = await sceneList({ query: "chaturbate", limit: 2 });
    expect(result.items).toHaveLength(2);
    expect(requested).toContain("https://www.showcamrips.com/site/en/Chaturbate/");
  });

  test("a url lists that page and rejects other pages", async () => {
    mockFetch({ "https://www.showcamrips.com/cat/en/Couple/": modelPage });
    const result = await sceneList({ url: "https://www.showcamrips.com/cat/en/Couple/", limit: 3, offset: 1 });
    expect(result.items.map((item) => item.external_id)).toEqual(["showcamrips-2210333", "showcamrips-2209040"]);
    await expect(sceneList({ url: VIDEO_URL })).rejects.toThrow("cannot list");
  });
});

// Live checks against the real site; chained so they survive catalog changes.
const NETWORK_TIMEOUT_MS = 60_000;

describe("live", () => {
  test(
    "browse, scrape and resolve the newest recording",
    async () => {
      const { items } = await sceneList({ limit: 5 });
      expect(items.length).toBeGreaterThan(0);

      const [result] = await scrape(items[0].source_url);
      const found = (result as { scene: { downloads: { url: string }[]; related: unknown[]; title: string } }).scene;
      expect(found.title.length).toBeGreaterThan(0);
      expect(found.related.length).toBeGreaterThan(0);

      const resolved = await resolve(found.downloads[0].url);
      expect(resolved.url).toStartWith("http");
      const head = await fetch(resolved.url, { headers: { ...resolved.headers, Range: "bytes=0-1" } });
      expect(head.ok).toBe(true);
    },
    NETWORK_TIMEOUT_MS
  );

  test(
    "searching a listed model's name finds recordings",
    async () => {
      // The search index lags behind brand-new models, so any of the newest will do.
      const newest = (await sceneList({ limit: 10 })).items;
      let found = 0;
      for (const item of newest) {
        const { items } = await sceneList({ query: item.title.split(" ")[0], limit: 5 });
        found += items.length;
      }
      expect(found).toBeGreaterThan(0);
    },
    NETWORK_TIMEOUT_MS * 2
  );
});
