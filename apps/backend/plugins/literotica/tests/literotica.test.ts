import { afterEach, describe, expect, test } from "bun:test";
import { readFileSync } from "fs";
import { fetchStoryPage, seriesUrl, targetOf } from "../src/api.ts";
import { categoryIdOf, categoryName } from "../src/categories.ts";
import { storyList } from "../src/listing.ts";
import { findStory, readChapter } from "../src/story.ts";
import { isoDateOf, joinPages, plainPageText } from "../src/text.ts";

/** Loads a saved API answer from the fixtures folder. */
function loadFixture(fileName: string): any {
  return JSON.parse(readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8"));
}

const realFetch = globalThis.fetch;
const requestedUrls: string[] = [];

/** Answers every fetch with the JSON the router returns for the URL; unrouted URLs are 404. */
function serve(route: (url: URL) => unknown): void {
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = new URL(String(input));
    requestedUrls.push(url.pathname + url.search);
    const body = route(url);
    if (body === undefined) {
      return new Response('{"success":false}', { status: 404 });
    }
    return new Response(JSON.stringify(body), { status: 200 });
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = realFetch;
  requestedUrls.length = 0;
});

describe("text", () => {
  test("page text becomes plain paragraphs", () => {
    const text = plainPageText('Title\r\n\r\n<em>Slanted</em> &amp; "quoted"\r\n\r\n\r\n\r\nEnd<sup>1</sup>\r\n');
    expect(text).toBe('Title\n\nSlanted & "quoted"\n\nEnd1');
  });

  test("pages are joined with a paragraph break", () => {
    expect(joinPages(["One\r\n", "", "Two"])).toBe("One\n\nTwo");
  });

  test("API dates become ISO dates", () => {
    expect(isoDateOf("11/02/2026")).toBe("2026-11-02");
    expect(isoDateOf("yesterday")).toBeUndefined();
  });
});

describe("urls and categories", () => {
  test("page URLs are classified", () => {
    expect(targetOf("https://www.literotica.com/s/fucking-deal?page=2")).toEqual({ kind: "work", slug: "fucking-deal" });
    expect(targetOf("https://www.literotica.com/series/se/lush-life")).toEqual({ kind: "series", slugOrId: "lush-life" });
    expect(targetOf("https://www.literotica.com/c/science-fiction-fantasy")).toEqual({ kind: "category", slug: "science-fiction-fantasy" });
    expect(targetOf("https://www.literotica.com/authors/2charlie/works/stories")).toEqual({ kind: "author", username: "2charlie" });
    expect(targetOf("https://www.literotica.com/new/stories")).toEqual({ kind: "newest" });
    expect(targetOf("https://www.literotica.com/")).toEqual({ kind: "unknown" });
  });

  test("a series without a slug is addressed by id", () => {
    expect(seriesUrl({ id: 494977452, url: "" })).toBe("https://www.literotica.com/series/se/494977452");
    expect(seriesUrl({ id: 1, url: "lush-life" })).toBe("https://www.literotica.com/series/se/lush-life");
  });

  test("categories map slugs to ids and names", () => {
    expect(categoryIdOf("science-fiction-fantasy")).toBe(38);
    expect(categoryIdOf("nope")).toBeUndefined();
    expect(categoryName("science-fiction-fantasy")).toBe("Sci-Fi & Fantasy");
    expect(categoryName("humor-and-satire")).toBe("Humor And Satire");
  });
});

describe("story:find", () => {
  test("a standalone work is a one-chapter story with its text", async () => {
    serve(() => loadFixture("work-standalone.json"));
    const [result] = await findStory("https://www.literotica.com/s/fucking-deal");
    if (result.type !== "story") throw new Error("expected a story");
    const story = result.story;
    expect(story.external_id).toBe(`story-${loadFixture("work-standalone.json").submission.id}`);
    expect(story.source_url).toBe("https://www.literotica.com/s/fucking-deal");
    expect(story.title).toBe("Fucking Deal");
    expect(story.performers?.[0].name).toBe("Nakedcraving");
    expect(story.date).toBe("2026-11-02");
    expect(story.word_count).toBe(2506);
    expect(story.complete).toBe(true);
    expect(story.chapters?.length).toBe(1);
    expect(story.chapters?.[0].text?.startsWith("The Fucking Deal\n\nWas No Art")).toBe(true);
    expect(story.chapters?.[0].text).not.toContain("\r");
  });

  test("a series URL lists its works as chapters, without text", async () => {
    serve((url) => {
      if (url.pathname === "/api/3/series/stonerager-chronicles") return loadFixture("series.json");
      if (url.pathname === "/api/3/series/494107485/works") return loadFixture("series-works.json");
      return undefined;
    });
    const [result] = await findStory("https://www.literotica.com/series/se/stonerager-chronicles");
    if (result.type !== "story") throw new Error("expected a story");
    const story = result.story;
    expect(story.external_id).toBe("series-494107485");
    expect(story.source_url).toBe("https://www.literotica.com/series/se/stonerager-chronicles");
    expect(story.title).toBe("Stonerager Chronicles");
    expect(story.details).toBe("22nd Century Astronaut finds himself nearly 1100 years in the future.");
    expect(story.complete).toBe(false);
    expect(story.tags?.[0]).toBe("Sci-Fi & Fantasy");
    expect(story.chapters?.map((chapter) => chapter.title)).toEqual([
      "Stonerager Chronicles - Ch. 01",
      "Stonerager Chronicles - Ch. 02",
      "Stonerager Chronicles - Ch. 03",
    ]);
    expect(story.chapters?.[0].url).toBe("https://www.literotica.com/s/stonerager-chronicles-ch-01");
    expect(story.chapters?.[0].text).toBeUndefined();
    expect(story.word_count).toBe(loadFixture("series-works.json").reduce((sum: number, work: any) => sum + work.words_count, 0));
  });

  test("a work of a series resolves to the series and keeps the text it already has", async () => {
    const openedPage = loadFixture("work-series-member.json");
    openedPage.meta.pages_count = 1;
    const works = loadFixture("series-works.json");
    works[2].id = openedPage.submission.id;
    serve((url) => {
      if (url.pathname === "/api/3/stories/stonerager-chronicles-bk-2-ch-12") return openedPage;
      if (url.pathname === "/api/3/series/494107485") return loadFixture("series.json");
      if (url.pathname === "/api/3/series/494107485/works") return works;
      return undefined;
    });
    const [result] = await findStory("https://www.literotica.com/s/stonerager-chronicles-bk-2-ch-12");
    if (result.type !== "story") throw new Error("expected a story");
    expect(result.story.external_id).toBe("series-494107485");
    expect(result.story.chapters?.[0].text).toBeUndefined();
    expect(result.story.chapters?.[2].text?.startsWith("Stonerager Chronicles\n\nBy 2Charlie")).toBe(true);
  });

  test("numeric tags are kept as text and repeats dropped", async () => {
    const page = loadFixture("work-standalone.json");
    page.submission.tags = [{ tag: 2024 }, { tag: "Couples" }, { tag: "couples" }, { tag: "" }];
    serve(() => page);
    const [result] = await findStory("https://www.literotica.com/s/fucking-deal");
    if (result.type !== "story") throw new Error("expected a story");
    expect(result.story.tags).toEqual(["Erotic Couplings", "2024", "Couples"]);
  });

  test("other URLs are refused", async () => {
    serve(() => undefined);
    await expect(findStory("https://www.literotica.com/c/mind-control")).rejects.toThrow("Not a Literotica story");
  });
});

describe("story:chapter", () => {
  test("every page of the work is fetched and joined", async () => {
    const first = loadFixture("work-series-member.json");
    first.meta.pages_count = 2;
    serve((url) => {
      if (url.searchParams.get("params") === '{"contentPage":2}') return loadFixture("work-page-2.json");
      return first;
    });
    const chapter = await readChapter("https://www.literotica.com/s/stonerager-chronicles-bk-2-ch-12");
    expect(chapter.title).toBe("Stonerager Chronicles Bk. 2 Ch. 12");
    expect(chapter.url).toBe("https://www.literotica.com/s/stonerager-chronicles-bk-2-ch-12");
    expect(chapter.text).toContain("Book 2, Chapter 12");
    expect(chapter.text).toContain("Pheebs stopped him cold");
    expect(requestedUrls.length).toBe(2);
  });

  test("a later page is requested with contentPage", async () => {
    serve(() => loadFixture("work-page-2.json"));
    await fetchStoryPage("some-story", 3);
    expect(requestedUrls[0]).toBe("/api/3/stories/some-story?params=%7B%22contentPage%22%3A3%7D");
  });
});

describe("story:list", () => {
  test("newest stories are searched by date, audio is dropped and a series is one entry", async () => {
    serve(() => loadFixture("search.json"));
    const { items } = await storyList({});
    expect(requestedUrls[0]).toBe("/api/3/search/stories?params=%7B%22sort%22%3A%22date%22%2C%22page%22%3A1%7D");
    expect(items.every((item) => item.media_type === "story")).toBe(true);
    expect(items.map((item) => item.title)).not.toContain("Hot Boss Wants You in her Office");
    const series = items.find((item) => item.external_id.startsWith("series-"));
    expect(series?.count).toBeGreaterThan(1);
    const alone = items.find((item) => item.title === "Friends, Enemies, Office Sex");
    expect(alone?.source_url).toBe("https://www.literotica.com/s/friends-enemies-office-sex");
    expect(alone?.date).toBe("2024-10-23");
    expect(alone?.rating).toBe(9);
  });

  test("every card credits its author as studio, with the identity story:find gives its performer", async () => {
    serve(() => loadFixture("search.json"));
    const { items } = await storyList({});
    expect(items.every((item) => item.studio?.name && item.studio.external_id === item.studio.name)).toBe(true);
    for (const item of items) {
      expect(item.studio?.source_url).toBe(`https://www.literotica.com/authors/${encodeURIComponent(item.studio!.name)}`);
    }
  });

  test("a query searches by relevance", async () => {
    serve(() => loadFixture("search.json"));
    await storyList({ query: "office" });
    expect(requestedUrls[0]).toBe("/api/3/search/stories?params=%7B%22q%22%3A%22office%22%2C%22page%22%3A1%7D");
  });

  test("a category page filters by its id, newest first", async () => {
    serve(() => loadFixture("search.json"));
    await storyList({ url: "https://www.literotica.com/c/science-fiction-fantasy" });
    expect(decodeURIComponent(requestedUrls[0])).toBe('/api/3/search/stories?params={"sort":"date","categories":[38],"page":1}');
  });

  test("an author page filters by the author", async () => {
    serve(() => loadFixture("search.json"));
    await storyList({ url: "https://www.literotica.com/authors/2charlie/works/stories" });
    expect(decodeURIComponent(requestedUrls[0])).toBe('/api/3/search/stories?params={"sort":"date","author":"2charlie","page":1}');
  });

  test("offset and limit pick the pages and the window inside them", async () => {
    const page = loadFixture("search.json");
    page.data = Array.from({ length: 50 }, (_, index) => ({ ...page.data[2], id: 1000 + index, url: `story-${index}`, series: [] }));
    serve(() => page);
    const { items } = await storyList({ offset: 45, limit: 10 });
    expect(requestedUrls.length).toBe(2);
    expect(items.length).toBe(10);
    expect(items[0].source_url).toBe("https://www.literotica.com/s/story-45");
  });

  test("a small limit still yields stories when audio tops the page", async () => {
    serve(() => loadFixture("search.json"));
    const { items } = await storyList({ query: "office", limit: 2 });
    expect(items.length).toBe(2);
    expect(items.every((item) => item.media_type === "story")).toBe(true);
  });

  test("unsupported pages are refused", async () => {
    serve(() => undefined);
    await expect(storyList({ url: "https://tags.literotica.com/harem/" })).rejects.toThrow("Cannot list stories");
    await expect(storyList({ url: "https://www.literotica.com/c/no-such" })).rejects.toThrow("Unknown Literotica category");
  });
});
