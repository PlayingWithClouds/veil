import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractRelated } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();

test("extractRelated parses the similar-videos grid in site order and skips the scene itself", () => {
  const related = extractRelated(parse(fixture), fixture, "hqporner-128024");

  expect(related.map((item) => item.external_id)).toEqual(["hqporner-128015", "hqporner-127787"]);
  expect(related[0]).toMatchObject({
    media_type: "scene",
    title: "present from a business partner",
    source_url: "https://hqporner.com/hdporn/128015-present_from_a_business_partner.html",
    poster_path: "https://fastporndelivery.hqporner.com/imgs/75/46/0ae064a995a0041_main.jpg",
    duration_seconds: 24 * 60 + 48,
  });
  expect(related[0].preview_images).toHaveLength(10);
  expect(related[1].title).toBe("pleasure overloaded");
});

test("extractRelated ignores the sidebar's recent-porn links", () => {
  const related = extractRelated(parse(fixture), fixture, "hqporner-0");
  expect(related.some((item) => item.external_id === "hqporner-128023")).toBe(false);
});
