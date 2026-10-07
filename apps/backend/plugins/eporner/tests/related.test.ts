import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractRelated } from "../src/scrape.ts";
import { parseClockDuration } from "../src/listing.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();

test("extractRelated parses the related grid in site order and skips the scene itself", () => {
  const related = extractRelated(parse(fixture), "eporner-qMt524baxFN");

  expect(related.map((item) => item.external_id)).toEqual(["eporner-mklZZn2M2ea", "eporner-qNitFgZULLc"]);
  expect(related[0]).toMatchObject({
    media_type: "scene",
    title:
      "Busty Single Woman Inserts A Soft Anal Plug Into Her Tight Anus And Records A Slutty Game Before Devoted Camera Fans",
    source_url:
      "https://www.eporner.com/video-mklZZn2M2ea/busty-single-woman-inserts-a-soft-anal-plug-into-her-tight-anus-and-records-a-slutty-game-before-devoted-camera-fans/",
    poster_path: "https://static-eu-cdn.eporner.com/thumbs/static4/1/18/184/18400412/8_240.jpg",
    duration_seconds: 120 * 60 + 6,
  });
  expect(related[1].title).toBe("Latina Ass Shake With Buttplug Show");
  expect(related[1].duration_seconds).toBe(4 * 60 + 16);
});

test("parseClockDuration handles mm:ss, long minutes and h:mm:ss", () => {
  expect(parseClockDuration("4:16")).toBe(256);
  expect(parseClockDuration("120:06")).toBe(7206);
  expect(parseClockDuration("1:02:03")).toBe(3723);
  expect(parseClockDuration("")).toBe(0);
});
