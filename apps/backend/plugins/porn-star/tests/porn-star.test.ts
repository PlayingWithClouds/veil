import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { attributeOf, decodeEntities, durationSeconds, resolveUrl } from "../src/http.ts";
import { cardsFromRoot, durationFromAlt, externalIdFromPath } from "../src/cards.ts";
import { sceneFromPage, scrapeScene } from "../src/scene.ts";
import { sceneList } from "../src/catalog.ts";

const NET_TIMEOUT = 30_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("durationSeconds reads SS, MM:SS and HH:MM:SS", () => {
    expect(durationSeconds("45")).toBe(45);
    expect(durationSeconds("37:06")).toBe(37 * 60 + 6);
    expect(durationSeconds("00:33:19")).toBe(33 * 60 + 19);
    expect(durationSeconds("x")).toBeUndefined();
  });

  test("resolveUrl resolves root-relative paths against the site", () => {
    expect(resolveUrl("/video/penthouse/24107/")).toBe("https://www.porn-star.com/video/penthouse/24107/");
    expect(resolveUrl("https://cdn.example/a.mp4")).toBe("https://cdn.example/a.mp4");
  });

  test("decodeEntities handles named and numeric entities", () => {
    expect(decodeEntities("Dap &amp; Fist &#39;x&#39;")).toBe("Dap & Fist 'x'");
  });

  test("externalIdFromPath keeps the site and id parts of a video path", () => {
    expect(externalIdFromPath("/video/penthouse/24107/")).toBe("porn-star-penthouse-24107");
    expect(externalIdFromPath("/video/pornworld/5329513-45326/")).toBe("porn-star-pornworld-5329513-45326");
    expect(externalIdFromPath("/river_lynn.html")).toBeUndefined();
  });

  test("durationFromAlt reads the runtime in a card's alt text", () => {
    expect(durationFromAlt("Teenage Auditions 14. (Video duration: 00:33:19)")).toBe(33 * 60 + 19);
    expect(durationFromAlt("Kay Carter - Video preview from Penthouse")).toBeUndefined();
  });
});

describe("video catalog page (offline)", () => {
  const root = parse(loadFixture("videos-page.html"));
  const items = cardsFromRoot(root);

  test("cardsFromRoot yields one scene stub per video card", () => {
    expect(items.length).toBe(24);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(24);
  });

  test("a query keeps only the cards whose names or description match", () => {
    const matches = cardsFromRoot(root, "redhead");
    expect(matches.map((item) => item.external_id)).toEqual(["porn-star-pornworld-5329513-45326"]);
    expect(cardsFromRoot(root, "no such word at all")).toEqual([]);
  });

  test("a card carries its performers, poster, runtime and URL", () => {
    const first = items[0];
    expect(first.media_type).toBe("scene");
    expect(first.title).toBe("Nicole Murkovski");
    expect(first.external_id).toBe("porn-star-pornworld-5329513-45326");
    expect(first.source_url).toBe("https://www.porn-star.com/video/pornworld/5329513-45326/");
    expect(first.poster_path).toBe("https://www.porn-star.com/tgpx/downloads/158990-530298.jpg");
    expect(first.duration_seconds).toBe(37 * 60 + 6);
  });
});

describe("video page (offline)", () => {
  const results = sceneFromPage("/video/penthouse/24107/", loadFixture("video-detail.html"));
  const [result] = results;

  test("the root is a scene keyed by site and id", () => {
    expect(results.length).toBe(1);
    expect(result.type).toBe("scene");
    if (result.type !== "scene") throw new Error("expected a scene");
    expect(result.scene.external_id).toBe("porn-star-penthouse-24107");
    expect(result.scene.source_url).toBe("https://www.porn-star.com/video/penthouse/24107/");
  });

  test("performers come from the headline and the studio from its link", () => {
    if (result.type !== "scene") throw new Error("expected a scene");
    const scene = result.scene;
    expect(scene.title).toBe("River Lynn");
    expect(scene.performers).toEqual([{ name: "River Lynn" }]);
    expect(scene.studio).toEqual({
      name: "Penthouse",
      external_id: "porn-star-studio-penthouse",
      source_url: "https://www.porn-star.com/destination/penthouse/",
    });
  });

  test("the playable mp4 and poster come from the player", () => {
    if (result.type !== "scene") throw new Error("expected a scene");
    expect(result.scene.downloads).toEqual([
      { label: "mp4", url: "https://www.porn-star.com/media/penthouse-24107.mp4", format: "mp4" },
    ]);
    expect(result.scene.poster_path).toBe("https://www.porn-star.com/tgpx/downloads/159027-1280720.jpg");
  });

  test("the related videos are stubs of this site", () => {
    if (result.type !== "scene") throw new Error("expected a scene");
    const related = result.scene.related ?? [];
    expect(related.length).toBeGreaterThan(0);
    for (const item of related) expect(item.external_id).toStartWith("porn-star-penthouse-");
  });

  test("a page that is not a video yields nothing", () => {
    expect(sceneFromPage("/river_lynn.html", loadFixture("video-detail.html"))).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// Live integration against porn-star.com.

describe("scene:list (live)", () => {
  test(
    "returns a bounded set of scene stubs",
    async () => {
      const { items } = await sceneList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) {
        expect(item.media_type).toBe("scene");
        expect(item.external_id).toStartWith("porn-star-");
        expect(item.source_url).toStartWith("https://www.porn-star.com/video/");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "offset pages do not repeat the first page",
    async () => {
      const first = await sceneList({ limit: 5 });
      const second = await sceneList({ limit: 5, offset: 30 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(second.items.length).toBe(5);
      expect(second.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test(
    "a query filters the newest catalog pages",
    async () => {
      const { items } = await sceneList({ query: "blonde", limit: 5 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(5);
      for (const item of items) expect(item.external_id).toStartWith("porn-star-");
    },
    4 * NET_TIMEOUT
  );
});

describe("scene:find (live)", () => {
  test(
    "scrapes a listed video with its mp4",
    async () => {
      const { items } = await sceneList({ limit: 1 });
      const [result] = await scrapeScene(items[0].source_url);
      expect(result.type).toBe("scene");
      if (result.type !== "scene") throw new Error("expected a scene");
      expect(result.scene.external_id).toBe(items[0].external_id);
      expect(result.scene.downloads?.[0].url).toEndWith(".mp4");
    },
    2 * NET_TIMEOUT
  );
});
