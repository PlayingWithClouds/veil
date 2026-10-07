import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractPerformerCredits, extractStudio, extractPageTags, mergeTags } from "../src/scrape.ts";

const uploaderPage = parse(await Bun.file(`${import.meta.dir}/fixtures/credits-uploader.html`).text());
const channelPage = parse(await Bun.file(`${import.meta.dir}/fixtures/credits-channel.html`).text());

test("extractStudio reads the uploader profile as the studio", () => {
  expect(extractStudio(uploaderPage)).toEqual({
    name: "glazers",
    source_url: "https://www.eporner.com/profile/glazers/",
  });
});

test("extractStudio reads a channel with the studio:find external_id", () => {
  expect(extractStudio(channelPage)).toEqual({
    name: "Vixen",
    source_url: "https://www.eporner.com/channel/vixen/",
    external_id: "eporner-ch-vixen",
  });
});

test("extractStudio prefers the channel over the uploader", () => {
  const page = parse(`<div id="video-info-tags"><ul>
    <li class="vit-uploader"><a href="/profile/someone/">someone</a></li>
    <li class="vit-channel"><a href="/channel/pure-taboo/">Pure Taboo</a></li>
  </ul></div>`);
  expect(extractStudio(page)?.source_url).toBe("https://www.eporner.com/channel/pure-taboo/");
});

test("extractStudio returns undefined without credits", () => {
  expect(extractStudio(parse(`<div id="video-info-tags"><ul></ul></div>`))).toBeUndefined();
});

test("extractPerformerCredits keeps the pornstar credits in order", () => {
  expect(extractPerformerCredits(channelPage)).toEqual([
    { name: "Mia Malkova", source_url: "https://www.eporner.com/pornstar/mia-malkova-oPgtJ/" },
    { name: "Emily Willis", source_url: "https://www.eporner.com/pornstar/emily-willis/" },
    { name: "Johnny Sins", source_url: "https://www.eporner.com/pornstar/johnny-sins-cgLkt/" },
  ]);
  expect(extractPerformerCredits(uploaderPage).map((performer) => performer.name)).toEqual(["Riley Reid"]);
});

test("extractPageTags reads categories then tags, skipping uploader and pornstar links", () => {
  expect(extractPageTags(uploaderPage)).toEqual([
    "Cumshot",
    "Blowjob",
    "Group Sex",
    "Riley reid anal",
    "Blowjob",
    "321sexchat",
  ]);
});

test("mergeTags keeps the API keywords first and dedupes case-insensitively", () => {
  expect(mergeTags(["blowjob", "Riley Reid"], extractPageTags(uploaderPage))).toEqual([
    "blowjob",
    "Riley Reid",
    "Cumshot",
    "Group Sex",
    "Riley reid anal",
    "321sexchat",
  ]);
});
