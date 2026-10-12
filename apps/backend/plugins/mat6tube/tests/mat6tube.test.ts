import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { cardsFromRoot } from "../src/cards.ts";
import { sceneFromPage, playlistSources } from "../src/scene.ts";
import { watchIdFromHref } from "../src/site.ts";
import { clockSeconds, countText } from "../src/text.ts";

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

const WATCH_ID = "-113656837_456242801";

describe("helpers", () => {
  test("watchIdFromHref accepts /watch/ links only, with or without query and fragment", () => {
    expect(watchIdFromHref("/watch/-113656837_456242801")).toBe(WATCH_ID);
    expect(watchIdFromHref("https://mat6tube.com/watch/-113656837_456242801#720p")).toBe(WATCH_ID);
    expect(watchIdFromHref("https://www.mat6tube.com/watch/-113656837_456242801?x=1")).toBe(WATCH_ID);
    expect(watchIdFromHref("https://mat6tube.com/video/blonde")).toBeUndefined();
    expect(watchIdFromHref("https://example.com/watch/1_2")).toBeUndefined();
  });

  test("clockSeconds and countText read the listing badges", () => {
    expect(clockSeconds("40:51")).toBe(40 * 60 + 51);
    expect(clockSeconds("1:02:03")).toBe(3723);
    expect(clockSeconds("views")).toBe(0);
    expect(countText("14.81M")).toBe(14_810_000);
    expect(countText("8.22K")).toBe(8220);
    expect(countText("1,234")).toBe(1234);
    expect(countText("")).toBe(0);
  });
});

describe("listing page (offline)", () => {
  const cards = cardsFromRoot(parse(loadFixture("listing-page.html")));

  test("one scene stub per video card", () => {
    expect(cards.length).toBe(24);
    expect(new Set(cards.map((item) => item.external_id)).size).toBe(24);
    expect(cards[0]!.external_id.startsWith("mat6tube-")).toBe(true);
  });

  test("a card carries its title, poster, preview, runtime and views", () => {
    const first = cards[0]!;
    expect(first.source_url).toBe(`https://mat6tube.com/watch/${first.external_id.slice("mat6tube-".length)}`);
    expect(first.title.length).toBeGreaterThan(0);
    expect(first.poster_path).toMatch(/^https:\/\/.*preview_320\.jpg$/);
    expect(first.preview_video).toMatch(/tr_240p\.mp4/);
    expect(first.duration_seconds).toBeGreaterThan(0);
    expect(first.view_count).toBeGreaterThan(0);
  });

  test("lazy-loaded cards take their poster from data-src, not the placeholder", () => {
    expect(cards.every((item) => item.poster_path?.startsWith("https://"))).toBe(true);
    expect(cards.some((item) => item.poster_path?.startsWith("data:"))).toBe(false);
  });
});

describe("video page (offline)", () => {
  const html = loadFixture("video-page.html");
  const scene = sceneFromPage(WATCH_ID, html);

  test("identity, title and studio", () => {
    expect(scene.external_id).toBe(`mat6tube-${WATCH_ID}`);
    expect(scene.source_url).toBe(`https://mat6tube.com/watch/${WATCH_ID}`);
    expect(scene.title.startsWith("Добро пожаловать домой")).toBe(true);
    expect(scene.studio).toEqual({ name: "mat6tube", source_url: "https://mat6tube.com" });
  });

  test("date, runtime, views and poster from the meta tags", () => {
    expect(scene.date).toBe("2022-05-31");
    expect(scene.duration_seconds).toBe(2567);
    expect(scene.view_count).toBe(15290303);
    expect(scene.poster_path).toBe("https://cdn2.pvvstream.pro/videos/-113656837/456242801/preview_800.jpg");
  });

  test("tags use the English translation where the page has one", () => {
    expect(scene.tags).toContain("stepmom");
    expect(scene.tags).toContain("sonny");
    expect(scene.tags).toContain("milf");
  });

  test("one page-relative download per playlist file, best first", () => {
    expect(scene.downloads?.map((download) => download.quality)).toEqual(["720p", "480p", "360p", "240p"]);
    expect(scene.downloads?.[0]!.url).toBe(`https://mat6tube.com/watch/${WATCH_ID}#720p`);
    expect(scene.downloads?.[0]!.format).toBe("mp4");
  });

  test("related videos exclude the video itself", () => {
    expect(scene.related?.length).toBe(24);
    expect(scene.related?.some((item) => item.external_id === `mat6tube-${WATCH_ID}`)).toBe(false);
  });

  test("playlist sources keep the file and label of each quality", () => {
    const sources = playlistSources(html);
    expect(sources.map((source) => source.label)).toEqual(["720", "480", "360", "240"]);
    expect(sources[0]!.file.startsWith("https://cdn2.pvvstream.pro/videos/-113656837/456242801/vid_720p.mp4")).toBe(true);
  });

  test("a page without a playlist yields no downloads", () => {
    expect(playlistSources("<html><body>gone</body></html>")).toEqual([]);
    expect(sceneFromPage(WATCH_ID, "<html><title>x</title></html>").downloads).toEqual([]);
  });
});
