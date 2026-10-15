import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { parseVideoCards, titleFromPostText } from "../src/cards.ts";
import { keywordFromQuery, listingUrl, sceneList } from "../src/listing.ts";
import { sceneFromPostPage, scrapeScene } from "../src/scene.ts";
import { hostedSourcePath, qualityFromInfo, resolve } from "../src/resolve.ts";
import { base64Encode, digitSum, playableUrl } from "../src/stream-token.ts";
import { suggest, suggestionsFromReply } from "../src/suggest.ts";
import { postIdFromUrl, decodeEntities } from "../src/http.ts";
import { clockSeconds, countFromText, isoDurationSeconds, viewsFromText } from "../src/values.ts";

const NET_TIMEOUT = 60_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

// The hosted post of the saved post page, and the source path its player script reads.
const HOSTED_POST_ID = "6ac9cd21eb1d5";
const HOSTED_SOURCE_PATH =
  "/cdn/c5/c2i1q7hzi2u3937zj3h77zt4744/8AaPAUsHfbYwQBwuH4ip-w/1791814930/m56f4e78sbi4pai3m9k3l59c9ap/g6raocm9ncyd32615ewb71bdw5r.vid";

describe("helpers", () => {
  test("post ids come from post URLs", () => {
    expect(postIdFromUrl("https://sxyprn.com/post/6ac9cd21eb1d5.html?sk=Search&so=0&ss=latest")).toBe(HOSTED_POST_ID);
    expect(postIdFromUrl("https://sxyprn.com/Blonde.html")).toBe("");
  });

  test("keywords are the query's words joined with dashes, punctuation dropped", () => {
    expect(keywordFromQuery("  Vanessa   Cage ")).toBe("Vanessa-Cage");
    expect(keywordFromQuery("milf, anal!")).toBe("milf-anal");
  });

  test("listing URLs: the first page has no page parameter, later pages carry their first offset", () => {
    expect(listingUrl({}, 0)).toBe("https://sxyprn.com/latest-videos.html");
    expect(listingUrl({}, 60)).toBe("https://sxyprn.com/latest-videos.html?page=60");
    expect(listingUrl({ query: "blonde milf" }, 30)).toBe("https://sxyprn.com/blonde-milf.html?page=30");
    expect(listingUrl({ url: "https://sxyprn.com/Binky-Beaz.html" }, 0)).toBe("https://sxyprn.com/Binky-Beaz.html");
  });

  test("runtime, count and ISO duration texts", () => {
    expect(clockSeconds("47:18")).toBe(2838);
    expect(clockSeconds("1:02:03")).toBe(3723);
    expect(clockSeconds("soon")).toBe(0);
    expect(isoDurationSeconds("PT47M18S")).toBe(2838);
    expect(isoDurationSeconds("PT1H2M3S")).toBe(3723);
    expect(isoDurationSeconds("P1D")).toBe(0);
    expect(countFromText("12.5K")).toBe(12500);
    expect(viewsFromText("2 days ago · 428 views")).toBe(428);
    expect(viewsFromText("3 days ago · 12,563 views")).toBe(12563);
    expect(viewsFromText("no count")).toBeUndefined();
  });

  test("entities decode in one pass", () => {
    expect(decodeEntities("a &amp; b &#039;c&#039; &amp;lt;")).toBe("a & b 'c' &lt;");
  });
});

describe("stream link (offline)", () => {
  test("base64 matches the standard encoding, padding included", () => {
    expect(base64Encode("72-sxyprn.com-45")).toBe("NzItc3h5cHJuLmNvbS00NQ==");
    expect(base64Encode("a")).toBe("YQ==");
    expect(base64Encode("ab")).toBe("YWI=");
    expect(base64Encode("abc")).toBe("YWJj");
  });

  test("digit sums ignore everything but digits", () => {
    expect(digitSum("m56f4e78")).toBe(5 + 6 + 4 + 7 + 8);
    expect(digitSum("abc")).toBe(0);
  });

  test("the playable URL of the saved source path, as the site's player builds it", () => {
    // Verified against the live site: this URL answers with a redirect to the video file.
    expect(playableUrl(HOSTED_SOURCE_PATH)).toBe(
      "https://sxyprn.com/cdn8/NzItc3h5cHJuLmNvbS00NQ../c5/c2i1q7hzi2u3937zj3h77zt4744/8AaPAUsHfbYwQBwuH4ip-w/1791814813/m56f4e78sbi4pai3m9k3l59c9ap/g6raocm9ncyd32615ewb71bdw5r.vid"
    );
  });

  test("a saved post page yields its hosted source path and quality", () => {
    const html = loadFixture("post-hosted.html");
    expect(hostedSourcePath(parse(html), HOSTED_POST_ID)).toBe(HOSTED_SOURCE_PATH);
    expect(hostedSourcePath(parse(html), "ffffffffffff")).toBe("");
    expect(qualityFromInfo(html)).toBe("720p");
  });

  test("a page without the source record has no hosted copy", () => {
    expect(hostedSourcePath(parse("<html><body>no player</body></html>"), HOSTED_POST_ID)).toBe("");
  });
});

describe("listing pages (offline)", () => {
  test("the newest posts: 30 cards with identity, title, poster, preview, runtime and views", () => {
    const items = parseVideoCards(parse(loadFixture("listing-latest.html")));
    expect(items.length).toBe(30);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(30);
    for (const item of items) {
      expect(item.external_id).toMatch(/^sxyprn-[0-9a-f]+$/);
      expect(item.source_url).toBe(`https://sxyprn.com/post/${item.external_id.slice("sxyprn-".length)}.html`);
      expect(item.title).not.toBe("");
      expect(item.title).not.toContain("#");
      expect(item.title).not.toContain("http");
      expect(item.media_type).toBe("scene");
    }
  });

  test("a card's media fields come from the card: poster, preview clip, runtime and views", () => {
    const first = parseVideoCards(parse(loadFixture("listing-latest.html")))[0];
    expect(first.title).not.toBe("");
    expect(first.poster_path?.startsWith("https://")).toBe(true);
    expect(first.preview_video?.endsWith("vidthumb.mp4")).toBe(true);
    expect(first.duration_seconds).toBeGreaterThan(0);
    expect(first.view_count).toBeGreaterThanOrEqual(0);
  });

  test("the keyword page lists its own posts, 30 per page", () => {
    const items = parseVideoCards(parse(loadFixture("listing-blonde.html")));
    expect(items.length).toBe(30);
  });

  test("a card's post text keeps the words and drops the markup around them", () => {
    const html =
      "<b class=\"post_b_text_el\">NEW</b> Cool <a class=\"hash_link tdn transition\" label=\"Milf\" href=\"/Milf.html\">#Milf</a> " +
      "scene <a href='https://vidara.so/v/x' target='_blank' title='>External Link!<' class='extlink_icon extlink'>vidara.so</a> " +
      "<b class='sub_cat_s' data-subkey='Only+Fans'> <a class=\"hash_link\" label=\"x\">#x</a> </b> <span class='sub_el'>+</span> " +
      "https://cuty.io/abc &amp; more";
    expect(titleFromPostText(html)).toBe("Cool scene & more");
  });
});

describe("post page (offline)", () => {
  const root = parse(loadFixture("post-hosted.html"));
  const scene = sceneFromPostPage(root, HOSTED_POST_ID);

  test("the scene carries the post's own metadata", () => {
    expect(scene.type).toBe("scene");
    expect(scene.external_id).toBe("sxyprn-6ac9cd21eb1d5");
    expect(scene.source_url).toBe("https://sxyprn.com/post/6ac9cd21eb1d5.html");
    expect(scene.title).toBe("Binky Beaz And Vanessa Cage Prime Cavity Search");
    expect(scene.date).toBe("2026-10-10");
    expect(scene.duration_seconds).toBe(2838);
    expect(scene.view_count).toBe(428);
    expect(scene.poster_path).toBe(
      "https://b2.trafficdeposit.com/pivi/1/6/vid/MGNAFjvmmWl34b7uPNLANw/1791813600/5fe8b4a3935ca/6ac9cd21eb1d5/full.jpg"
    );
  });

  test("tags are the hashtags plus the category, and performers link to their keyword pages", () => {
    expect(scene.tags).toContain("Hardcore");
    expect(scene.tags).toContain("MYLF");
    expect(scene.performers?.map((performer) => performer.name)).toContain("Vanessa Cage");
    expect(scene.performers?.find((performer) => performer.name === "Binky Beaz")?.source_url).toBe(
      "https://sxyprn.com/Binky-Beaz.html"
    );
  });

  test("the uploader is the studio, identified by its blog", () => {
    expect(scene.studio).toEqual({
      name: "MeetMyLove",
      external_id: "5fe8b4a3935ca",
      source_url: "https://sxyprn.com/blog/5fe8b4a3935ca/0.html",
    });
  });

  test("related posts are the page's other cards: 24 cards, one repeated, the post itself left out", () => {
    expect(scene.related?.length).toBe(22);
    for (const item of scene.related ?? []) {
      expect(item.external_id).not.toBe(scene.external_id);
    }
  });

  test("the download is the post page, resolved when it plays: the playable link expires", () => {
    expect(scene.downloads).toEqual([
      { label: "sxyprn", url: "https://sxyprn.com/post/6ac9cd21eb1d5.html", format: "mp4" },
    ]);
  });

  test("a post that only links to other hosts has no download", () => {
    const withoutPlayer = loadFixture("post-hosted.html").replace(/data-vnfo=/g, "data-removed=");
    expect(sceneFromPostPage(parse(withoutPlayer), HOSTED_POST_ID).downloads).toBeUndefined();
  });
});

describe("search suggestions (offline)", () => {
  test("the live-search reply gives its keyword matches as queries, not its posts", () => {
    const suggestions = suggestionsFromReply(loadFixture("suggest-blonde.json"));
    expect(suggestions[0]).toEqual({ text: "Blonde", kind: "query" });
    expect(suggestions.map((suggestion) => suggestion.text)).toContain("Sara Blonde");
    expect(suggestions.length).toBe(8);
  });
});

// ---------------------------------------------------------------------------
// Live integration against sxyprn.com (plain requests; the site needs no solver).

describe("scene:list (live)", () => {
  test(
    "returns a bounded page of video stubs",
    async () => {
      const { items } = await sceneList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) expect(item.external_id).toStartWith("sxyprn-");
    },
    NET_TIMEOUT
  );

  test(
    "offsets past the first page do not repeat it",
    async () => {
      const first = await sceneList({ limit: 5 });
      const later = await sceneList({ limit: 5, offset: 30 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(later.items.length).toBe(5);
      expect(later.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test(
    "a keyword query returns posts",
    async () => {
      const { items } = await sceneList({ query: "blonde", limit: 5 });
      expect(items.length).toBe(5);
    },
    NET_TIMEOUT
  );
});

describe("scene:find, stream:resolve and search:suggest (live)", () => {
  test(
    "a listed post scrapes to a scene with the same identity",
    async () => {
      const { items } = await sceneList({ limit: 1 });
      const [result] = await scrapeScene(items[0].source_url);
      expect(result.type).toBe("scene");
      if (result.type !== "scene") throw new Error("expected a scene");
      expect(result.scene.external_id).toBe(items[0].external_id);
      expect(result.scene.title).not.toBe("");
    },
    NET_TIMEOUT
  );

  test(
    "a hosted post resolves to a video file the site redirects to",
    async () => {
      const { items } = await sceneList({ limit: 10 });
      let resolved: Awaited<ReturnType<typeof resolve>> | undefined;
      for (const item of items) {
        try {
          resolved = await resolve(item.source_url);
          break;
        } catch (error) {
          // Posts that only link to other hosts have no playable copy; try the next one.
          if (!String(error).includes("no hosted video")) throw error;
        }
      }
      if (resolved === undefined) throw new Error("no hosted post among the first 10");
      expect(resolved.mime_type).toBe("video/mp4");
      expect(resolved.url.startsWith("https://")).toBe(true);
      expect(resolved.url.includes("sxyprn.com/cdn8/")).toBe(false);
      const range = await fetch(resolved.url, { headers: { Range: "bytes=0-15" } });
      expect(range.status).toBe(206);
      expect(range.headers.get("content-type")).toBe("video/mp4");
    },
    3 * NET_TIMEOUT
  );

  test(
    "suggestions come back for a partly typed query",
    async () => {
      const { suggestions } = await suggest({ query: "blon", limit: 5 });
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions.length).toBeLessThanOrEqual(5);
    },
    NET_TIMEOUT
  );
});
