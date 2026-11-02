import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { postedDate } from "../src/dates.ts";
import { parseAuthorPage, parseStoryCards, sourceOf, sourceOfUrl, storyList } from "../src/listing.ts";
import { externalIdOf, storyIdOf, storySlugOf, storyUrl } from "../src/site.ts";
import { findStory, readChapter, storyFromPage } from "../src/story.ts";
import { countWords, plainText } from "../src/text.ts";

const NET_TIMEOUT = 60_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("addresses", () => {
  test("story ids and slugs are read from story URLs", () => {
    const url = "https://www.sexstories.com/story/115749/my_bff_and_my_daddy";
    expect(storyIdOf(url)).toBe("115749");
    expect(storySlugOf(url)).toBe("my_bff_and_my_daddy");
    expect(storyIdOf("https://sexstories.com/story/115749/")).toBe("115749");
    expect(storySlugOf("https://sexstories.com/story/115749/")).toBe("");
    expect(storyIdOf("https://www.sexstories.com/profile1/x")).toBeUndefined();
    expect(storyIdOf("https://example.com/story/1/x")).toBeUndefined();
  });

  test("the slug is optional when building a story URL", () => {
    expect(storyUrl("5", "")).toBe("https://www.sexstories.com/story/5/");
    expect(storyUrl("5", "a_b")).toBe("https://www.sexstories.com/story/5/a_b");
    expect(externalIdOf("5")).toBe("sexstories-5");
  });
});

describe("text", () => {
  test("line breaks become newlines and blank lines stay paragraph breaks", () => {
    const markup = "One &quot;quoted&quot; line<br />\n<br />\nTwo<br />\nstill two <span class=\"italic\">it&#039;s</span><br />\n";
    expect(plainText(markup)).toBe('One "quoted" line\n\nTwo\nstill two it\'s');
  });

  test("typographic entities are decoded", () => {
    expect(plainText("a&nbsp;b &rsquo; &mdash; &hellip;")).toBe("a b ’ — …");
  });

  test("words are counted", () => {
    expect(countWords("one two\n\nthree")).toBe(3);
    expect(countWords("")).toBe(0);
  });
});

describe("dates", () => {
  test("a written date becomes an ISO date", () => {
    expect(postedDate("Posted Thu 22nd of October 2026 Diary")).toBe("2026-10-22");
    expect(postedDate("Posted Sat 3rd of October 2026")).toBe("2026-10-03");
  });

  test("a relative time is approximated", () => {
    const date = postedDate("Posted 3 days ago Fiction , Wife");
    expect(date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  test("text without Posted has no date", () => {
    expect(postedDate("Read 3 times")).toBeUndefined();
  });
});

describe("listing pages", () => {
  const items = parseStoryCards(loadFixture("updates.html"));

  test("every card becomes a story item", () => {
    expect(items.length).toBe(10);
    expect(items[0]).toEqual({
      title: "Mommy & Dada's Mutt",
      media_type: "story",
      source_url: "https://www.sexstories.com/story/115683/mommy_amp_dada_039_s_mutt",
      external_id: "sexstories-115683",
      date: "2026-10-01",
      rating: 10,
      view_count: 261777,
    });
  });

  test("cards with relative dates and without an author still parse", () => {
    const relative = items.find((item) => item.external_id === "sexstories-115788");
    expect(relative?.title).toBe("I CAN'T STOP FUCKING OUR NEIGHBOR'S SON");
    expect(relative?.date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(relative?.rating).toBeCloseTo(9.4);
  });

  test("non-breaking spaces in titles are plain spaces", () => {
    const storm = items.find((item) => item.external_id === "sexstories-115789");
    expect(storm?.title).toBe("The Perfect Storm: Part 8 - Sarah gets her nudist sanctuary");
  });

  test("search results use the same cards", () => {
    const results = parseStoryCards(loadFixture("search.html"));
    expect(results.length).toBe(9);
    expect(results[0].media_type).toBe("story");
  });

  test("an author page lists the stories of its table", () => {
    const stories = parseAuthorPage(loadFixture("profile.html"));
    expect(stories.length).toBe(5);
    expect(stories[0]).toEqual({
      title: "My BFF and My Daddy",
      media_type: "story",
      source_url: "https://www.sexstories.com/story/115749/",
      external_id: "sexstories-115749",
      view_count: 107913,
      rating: 10,
    });
  });
});

describe("listing sources", () => {
  test("no query and no URL lists the latest stories", () => {
    const source = sourceOf({});
    expect(source.pageUrl(2)).toBe("https://www.sexstories.com/toplists/last/2");
    expect(source.pageSize).toBe(100);
  });

  test("a query searches, joining words with underscores, on one page", () => {
    const source = sourceOf({ query: "big tits" });
    expect(source.pageUrl(1)).toBe("https://www.sexstories.com/search/1/relevance/big_tits////");
    expect(source.pageSize).toBeUndefined();
  });

  test("category, toplist, updates and author URLs are recognised", () => {
    expect(sourceOfUrl("https://www.sexstories.com/themes/121/Dark+fiction")?.pageUrl(2)).toBe(
      "https://www.sexstories.com/themes/121/Dark+fiction/s-date/p-2",
    );
    expect(sourceOfUrl("https://www.sexstories.com/genres/13/True+Story/s-rate/p-3")?.pageUrl(1)).toBe(
      "https://www.sexstories.com/genres/13/True+Story/s-date/p-1",
    );
    expect(sourceOfUrl("https://www.sexstories.com/toplists/bests30d")?.pageUrl(3)).toBe(
      "https://www.sexstories.com/toplists/bests30d/3",
    );
    expect(sourceOfUrl("https://www.sexstories.com/updates/")?.pageUrl(2)).toBe("https://www.sexstories.com/updates/2");
    expect(sourceOfUrl("https://www.sexstories.com/profile2052904/lila_hart")?.authorPage).toBe(true);
    expect(sourceOfUrl("https://www.sexstories.com/contact")).toBeUndefined();
  });
});

describe("story page", () => {
  const html = loadFixture("story.html");
  const url = "https://www.sexstories.com/story/115749/my_bff_and_my_daddy";
  const story = storyFromPage(html, "115749", url);

  test("carries the title, author, themes, date and introduction", () => {
    expect(story?.title).toBe("My BFF and My Daddy");
    expect(story?.external_id).toBe("sexstories-115749");
    expect(story?.source_url).toBe(url);
    expect(story?.performers).toEqual([
      {
        name: "lila_hart",
        external_id: "sexstories-author-2052904",
        source_url: "https://www.sexstories.com/profile2052904/lila_hart",
      },
    ]);
    expect(story?.tags).toEqual(["Diary", "Written by women", "Young"]);
    expect(story?.date).toBe("2026-10-22");
    expect(story?.details).toBe("Dark, taboo, and emotionally messy.\n\nIf that isn’t your thing, please move along.");
    expect(story?.complete).toBe(true);
  });

  test("the text is one chapter without the votes block, comments or markup", () => {
    expect(story?.chapters?.length).toBe(1);
    const chapter = story?.chapters?.[0];
    expect(chapter?.url).toBe(url);
    expect(chapter?.text).toStartWith("All chars are 18+ adults.\n\nI was sitting on my bed");
    expect(chapter?.text).toContain("\"Mr. Hale did it again,\" she said.");
    expect(chapter?.text).toContain("Authors note:");
    expect(chapter?.text).not.toContain("POSITIVE");
    expect(chapter?.text).not.toContain("Loved this story");
    expect(chapter?.text).not.toContain("<");
    expect(story?.word_count).toBe(chapter?.word_count);
    expect(story?.word_count).toBeGreaterThan(300);
  });

  test("a page without a story gives nothing", () => {
    expect(storyFromPage("<html><body><h1>sexstories.com</h1></body></html>", "9", url)).toBeUndefined();
  });

  test("a non-story address finds nothing", async () => {
    expect(await findStory("https://www.sexstories.com/themes/1/Fiction")).toEqual([]);
  });
});

describe("live site", () => {
  test(
    "the latest stories list and a story reads",
    async () => {
      const { items } = await storyList({ limit: 3 });
      expect(items.length).toBe(3);
      const results = await findStory(items[0].source_url);
      expect(results[0].type).toBe("story");
      const chapter = await readChapter(items[0].source_url);
      expect((chapter.text ?? "").length).toBeGreaterThan(100);
    },
    NET_TIMEOUT,
  );

  test(
    "a search answers with stories",
    async () => {
      const { items } = await storyList({ query: "daddy", limit: 5 });
      expect(items.length).toBe(5);
    },
    NET_TIMEOUT,
  );
});
