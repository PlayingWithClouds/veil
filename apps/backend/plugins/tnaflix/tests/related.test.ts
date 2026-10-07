import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractRelated } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();

test("extractRelated parses the related grid in site order and skips the scene itself", () => {
  const related = extractRelated(parse(fixture), "tnaflix-26135734");

  expect(related.map((item) => item.external_id)).toEqual(["tnaflix-10880491", "tnaflix-1731016"]);
  expect(related[0]).toMatchObject({
    title: "Emo slut pov ass fucked",
    media_type: "scene",
    source_url: "https://www.tnaflix.com/amateur-porn/Emo-slut-pov-ass-fucked/video10880491",
    poster_path: "https://img.tnaflix.com/a16:8q80w356r/194/10/88/10880491/thumbs/10.jpg",
    duration_seconds: 375,
  });
  expect(related[0].preview_video).toContain("/10880491/trailer.mp4");
  expect(related[1].title).toBe("Man is fucking cute gal - video 17");
  expect(related[1].duration_seconds).toBe(308);
});
