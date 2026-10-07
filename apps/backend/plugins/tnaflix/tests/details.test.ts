import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractStudio, extractPerformers, extractTags, extractDetails } from "../src/scrape.ts";

const channelPage = parse(await Bun.file(`${import.meta.dir}/fixtures/video-channel.html`).text());
const uploaderPage = parse(await Bun.file(`${import.meta.dir}/fixtures/video-uploader.html`).text());

/** The badge strip of a parsed fixture page. */
function badgesOf(root: ReturnType<typeof parse>) {
  return root.querySelector("div.video-detail-badges");
}

test("only a verified channel badge becomes the studio", () => {
  expect(extractStudio(badgesOf(channelPage))).toEqual({
    name: "Brazzers",
    source_url: "https://www.tnaflix.com/channel/brazzers",
  });
  expect(extractStudio(badgesOf(uploaderPage))).toBeUndefined();
  expect(extractStudio(null)).toBeUndefined();
});

test("performers are the badge-kiss profiles only", () => {
  expect(extractPerformers(badgesOf(channelPage)).map((performer) => performer.name)).toEqual([
    "Danny D",
    "Sophia Laure",
  ]);
  expect(extractPerformers(badgesOf(uploaderPage))).toEqual([]);
});

test("tags are category chips, without entity badges or title-derived search chips", () => {
  expect(extractTags(badgesOf(channelPage))).toEqual(["Blowjobs & Oral Sex", "HD Porn", "Porn Stars"]);
  expect(extractTags(badgesOf(uploaderPage))).toEqual(["HD Porn", "Porn Compilations"]);
});

test("details come from the description paragraph and skip placeholders", () => {
  expect(extractDetails(channelPage)).toBe(
    "Brazzers - Big Tits In Sports - Sophia Laure and Danny D - Sweaty Ass Workout"
  );
  // Empty paragraph falls back to JSON-LD, which only has the placeholder.
  expect(extractDetails(uploaderPage)).toBeUndefined();
});

test("details fall back to JSON-LD and drop the generated SEO blurbs", () => {
  const jsonLdOnly = parse(
    '<script type="application/ld+json">{"description":"A real &amp; decoded description"}</script>'
  );
  expect(extractDetails(jsonLdOnly)).toBe("A real & decoded description");

  const seoBlurb = parse(
    '<p class="video-detail-description"><b>Description</b>: Watch Deepthroat on  com&amp;comma; the best hardcore porn site   is home to the widest selection</p>'
  );
  expect(extractDetails(seoBlurb)).toBeUndefined();

  const nowBlurb = parse(
    '<p class="video-detail-description"><b>Description</b>: Watch Br1g1tTe LahA1e on  now! - Vintage, Anal Porn</p>'
  );
  expect(extractDetails(nowBlurb)).toBeUndefined();
});
