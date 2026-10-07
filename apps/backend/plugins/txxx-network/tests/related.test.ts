import { test, expect } from "bun:test";
import { relatedApiPath } from "../src/http.ts";
import { relatedItems } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.json`).json();

test("relatedApiPath buckets the id like the SPA's videos_related2 call", () => {
  expect(relatedApiPath("21796149", 41)).toBe(
    "/api/json/videos_related2/432000/41/21000000/21796000/21796149.all.1.json"
  );
});

test("relatedItems maps the related payload in API order, skipping the scene and duplicates", () => {
  const related = relatedItems("hclips.com", "21796149", fixture.videos);

  expect(related.map((item) => item.external_id)).toEqual(["txxx-21446347", "txxx-20652313"]);
  expect(related[0]).toMatchObject({
    title: "Valentines With Asian Stepsis And Bestie - Valentine S Day, Ricky Spanish And Kyler Quinn",
    media_type: "scene",
    source_url:
      "https://hclips.com/videos/21446347/valentines-with-asian-stepsis-and-bestie-valentine-s-day-ricky-spanish-and-kyler-quinn/",
    date: "2026-01-26",
    poster_path: "https://tn.txxx.tube/contents/videos_screenshots/21446000/21446347/288x162/1.jpg",
    preview_video: "https://vp2.txxx.com/c12/videos/21446000/21446347/21446347_tr.mp4",
    duration_seconds: 541,
  });
});

test("relatedItems tolerates a missing videos array", () => {
  expect(relatedItems("txxx.com", "1", undefined)).toEqual([]);
});
