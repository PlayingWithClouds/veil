import { test, expect } from "bun:test";
import { candidateUrls, isActressUrl, parsePerformerPage, performerIdentify, type PageFetcher } from "../src/performers.ts";

const fixtures = `${import.meta.dir}/fixtures`;
const yuaHtml = await Bun.file(`${fixtures}/actress-yua-mikami.html`).text();
const miaHtml = await Bun.file(`${fixtures}/actress-mia-malkova.html`).text();
const missingHtml = await Bun.file(`${fixtures}/actress-missing.html`).text();

const yuaUrl = "https://missav.ws/en/actresses/Yua%20Mikami";

/** Serves the given pages by URL and records every URL asked; anything else is an empty "no such actress" page. */
function fakeSite(pages: Record<string, string>): { fetchHtml: PageFetcher; asked: string[] } {
  const asked: string[] = [];
  const fetchHtml: PageFetcher = async (url) => {
    asked.push(url);
    return pages[url] ?? missingHtml;
  };
  return { fetchHtml, asked };
}

test("isActressUrl accepts profiles and rejects the index, ranking and video pages", () => {
  expect(isActressUrl(yuaUrl)).toBe(true);
  expect(isActressUrl("https://missav.ws/dm44/en/actresses/Nozomi%20Ishihara")).toBe(true);
  expect(isActressUrl("https://missav.ws/en/actresses")).toBe(false);
  expect(isActressUrl("https://missav.ws/en/actresses/ranking")).toBe(false);
  expect(isActressUrl("https://missav.ws/en/mide-900")).toBe(false);
});

test("parsePerformerPage fills the profile from the hero block", () => {
  const performer = parsePerformerPage(yuaHtml, "https://missav.ws/en/actresses/%E4%B8%89%E4%B8%8A%E6%82%A0%E4%BA%9C");

  expect(performer).toMatchObject({
    type: "performer",
    external_id: "missav-actress-Yua Mikami",
    source_url: yuaUrl,
    name: "Yua Mikami",
    aliases: ["三上悠亜"],
    birthdate: "1993-08-16",
    height_cm: 159,
    measurements: "33F-23-35",
    image_path: "https://fourhoi.com/actress/1030262-t.jpg",
  });
});

test("parsePerformerPage keeps only the name when the page shows no stats", () => {
  const performer = parsePerformerPage(miaHtml, "https://missav.ws/en/actresses/Mia%20Malkova");

  expect(performer?.name).toBe("Mia Malkova");
  expect(performer?.aliases).toEqual(["ミア・マルコヴァ"]);
  expect(performer?.birthdate).toBeUndefined();
  expect(performer?.height_cm).toBeUndefined();
  expect(performer?.image_path).toBeUndefined();
});

test("parsePerformerPage returns null for a name without an actress", () => {
  expect(parsePerformerPage(missingHtml, "https://missav.ws/en/actresses/Zzqx%20Nobody")).toBeNull();
});

test("candidateUrls lists known missav profiles first, then one page per name", () => {
  expect(
    candidateUrls({
      name: "Yua Mikami",
      aliases: ["三上悠亜"],
      urls: ["https://example.com/a", "https://missav.ws/dm218/en/actresses/Yua%20Mikami"],
    }),
  ).toEqual([yuaUrl, "https://missav.ws/en/actresses/%E4%B8%89%E4%B8%8A%E6%82%A0%E4%BA%9C"]);
});

test("performerIdentify returns the profile of a matching name", async () => {
  const { fetchHtml, asked } = fakeSite({ [yuaUrl]: yuaHtml });
  const result = await performerIdentify({ name: "Yua Mikami" }, fetchHtml);

  expect(result.performer?.external_id).toBe("missav-actress-Yua Mikami");
  expect(asked).toEqual([yuaUrl]);
});

test("performerIdentify finds the profile through a Japanese alias", async () => {
  const japaneseUrl = "https://missav.ws/en/actresses/%E4%B8%89%E4%B8%8A%E6%82%A0%E4%BA%9C";
  const { fetchHtml } = fakeSite({ [japaneseUrl]: yuaHtml });
  const result = await performerIdentify({ name: "Mikami Yua", aliases: ["三上悠亜"] }, fetchHtml);

  expect(result.performer?.name).toBe("Yua Mikami");
});

test("performerIdentify returns nothing when the site has no such actress", async () => {
  const { fetchHtml } = fakeSite({});
  expect(await performerIdentify({ name: "Zzqx Nobody" }, fetchHtml)).toEqual({});
});

test("performerIdentify rejects a page whose name is someone else's", async () => {
  // The site answered a different actress's page for the asked-for URL.
  const { fetchHtml } = fakeSite({ "https://missav.ws/en/actresses/Mia%20Malkova": yuaHtml });
  expect(await performerIdentify({ name: "Mia Malkova" }, fetchHtml)).toEqual({});
});

test("performerIdentify rejects a matching name with a different birthdate", async () => {
  const { fetchHtml } = fakeSite({ [yuaUrl]: yuaHtml });
  expect(await performerIdentify({ name: "Yua Mikami", birthdate: "1990-01-01" }, fetchHtml)).toEqual({});
  const same = await performerIdentify({ name: "Yua Mikami", birthdate: "1993-08-16" }, fetchHtml);
  expect(same.performer?.birthdate).toBe("1993-08-16");
});
