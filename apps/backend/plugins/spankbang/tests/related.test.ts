import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractRelated } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();

test("extractRelated reads the bottom then sidebar related blocks, skipping the scene and duplicates", () => {
  const related = extractRelated(parse(fixture), "spankbang-a59db");

  expect(related.map((item) => item.external_id)).toEqual([
    "spankbang-a4wzb",
    "spankbang-6evis",
    "spankbang-646ly",
  ]);
  expect(related[0]).toMatchObject({
    media_type: "scene",
    title: "Goddess body line's overwhelming beautiful big boobs beauty",
    source_url: "https://spankbang.com/a4wzb/video/porn",
    poster_path: "https://tbi.sb-cd.com/t/17025527/c9/fd/w:300/t6-enh/porn.jpg",
    preview_video: "https://tbv.sb-cd.com/t/17025527/c9/fd/td.mp4",
    duration_seconds: 2460,
  });
});

test("extractRelated ignores the promoted video-list in the navigation", () => {
  const related = extractRelated(parse(fixture), "spankbang-a59db");
  expect(related.some((item) => item.external_id === "spankbang-a4ejw")).toBe(false);
});
