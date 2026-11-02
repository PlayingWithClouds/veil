import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractMarkers } from "../src/scrape.ts";

const fixture = await Bun.file(`${import.meta.dir}/fixtures/skip-to-positions.html`).text();

test("the Skip to buttons become point markers", () => {
  expect(extractMarkers(parse(fixture))).toEqual([
    { tag: "Cowgirl", seconds: 144 },
    { tag: "Blowjob", seconds: 488 },
    { tag: "Doggy", seconds: 536 },
    { tag: "Missionary", seconds: 592 },
  ]);
});

test("a page without the buttons has no markers", () => {
  expect(extractMarkers(parse("<html><body><h1>x</h1></body></html>"))).toEqual([]);
});

test("multi-word positions are title-cased", () => {
  const html = '<button class="js-scene-position" data-timestamp="7">reverse cowgirl</button>';
  expect(extractMarkers(parse(html))).toEqual([{ tag: "Reverse Cowgirl", seconds: 7 }]);
});
