import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractRelated } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();

test("extractRelated reads only the related block, in site order, without the scene itself", () => {
  const related = extractRelated(parse(fixture), "xhamster-xhuku55");

  expect(related.map((item) => item.external_id)).toEqual(["xhamster-xhaXqmd", "xhamster-xhuY0GB"]);
  expect(related[0]).toMatchObject({
    title: "Sexy Petite Tiffany Tatum Gets Banged the Hardest by a Monster Cock of Rob Diesel",
    media_type: "scene",
    source_url:
      "https://xhamster.com/videos/sexy-petite-tiffany-tatum-gets-banged-the-hardest-by-a-monster-cock-of-rob-diesel-xhaXqmd",
    duration_seconds: 672,
  });
  expect(related[0].poster_path).toContain("xhcdn.com");
  expect(related[0].preview_video).toEndWith(".mp4");
  expect(related[1].title).toStartWith("Tiffany Tatum's puny assets");
});

test("extractRelated returns nothing when the page has no related block", () => {
  expect(extractRelated(parse("<html><body></body></html>"), "xhamster-x")).toEqual([]);
});
