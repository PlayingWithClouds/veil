import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractPerformers, extractStudio, extractUploader } from "../src/scrape.ts";

const channelPage = parse(await Bun.file(`${import.meta.dir}/fixtures/credits-channel.html`).text());
const creatorPage = parse(await Bun.file(`${import.meta.dir}/fixtures/credits-creator.html`).text());

test("extractPerformers reads the pornstar links of the video-tags block in order", () => {
  expect(extractPerformers(channelPage)).toEqual([
    { name: "Jean Pallett", source_url: "https://spankbang.com/cx9/pornstar/jean+pallett/" },
    { name: "Crystal White", source_url: "https://spankbang.com/5ms/pornstar/crystal+white/" },
  ]);
});

test("extractPerformers is empty without pornstar credits", () => {
  expect(extractPerformers(creatorPage)).toEqual([]);
});

test("extractStudio reads the channel", () => {
  expect(extractStudio(channelPage)).toEqual({
    name: "SalsaXXX",
    source_url: "https://spankbang.com/b2/channel/salsaxxx/",
  });
});

test("the creator profile is the uploader, not a studio", () => {
  expect(extractStudio(creatorPage)).toBeUndefined();
  expect(extractUploader(creatorPage)).toEqual({
    name: "Miguelmegadotado",
    url: "https://spankbang.com/4wpt/creator/miguelmegadotado/",
  });
});

test("a video with a channel and a creator gets both", () => {
  const page = parse(`<div data-testid="video-tags">
    <a href="/4wpt/creator/someone/">someone</a>
    <a href="/iq/channel/mofos/">MOFOS</a>
  </div>`);
  expect(extractStudio(page)?.source_url).toBe("https://spankbang.com/iq/channel/mofos/");
  expect(extractUploader(page)?.url).toBe("https://spankbang.com/4wpt/creator/someone/");
});
