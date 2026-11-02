import { test, expect, afterEach } from "bun:test";
import { readFileSync } from "node:fs";
import { sceneList, pageUrl } from "../src/scenes.ts";

const profilePage = readFileSync(new URL("./fixtures/actress-videos-yua-mikami.html", import.meta.url), "utf8");
const emptyPage = "<!doctype html><html><body><div class=\"grid\"></div></body></html>";
const ACTRESS = "https://missav.ws/en/actresses/Yua%20Mikami";
const originalFetch = globalThis.fetch;
const requested: string[] = [];

// Serves the fixture for page 1, re-coded copies for page 2, and (like the site) an empty 200 afterwards.
function mockProfile(): void {
  requested.length = 0;
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = String(input);
    requested.push(url);
    if (url === ACTRESS) return new Response(profilePage);
    if (url === `${ACTRESS}?page=2`) return new Response(profilePage.replace(/ofje-(\d+)/g, "pfje-$1"));
    return new Response(emptyPage);
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("page URLs: profile url, search and newest feed", () => {
  expect(pageUrl(`${ACTRESS}/?x=1`, "", 1)).toBe(ACTRESS);
  expect(pageUrl(ACTRESS, "ignored", 3)).toBe(`${ACTRESS}?page=3`);
  expect(pageUrl(undefined, "a b", 2)).toBe("https://missav.ws/en/search/a%20b?page=2");
  expect(pageUrl(undefined, "", 1)).toBe("https://missav.ws/en/new");
});

test("an actress url lists that page's videos with the normal external ids", async () => {
  mockProfile();
  const result = await sceneList({ url: ACTRESS, limit: 3 });
  expect(result.items.map((item) => item.external_id)).toEqual(["missav-ofje-667", "missav-ofje-637", "missav-ofje-629"]);
  expect(requested).toEqual([ACTRESS]);
});

test("offset continues on the next page and stops past the end", async () => {
  mockProfile();
  const second = await sceneList({ url: ACTRESS, offset: 3, limit: 2 });
  expect(second.items.map((item) => item.external_id)).toEqual(["missav-pfje-667", "missav-pfje-637"]);

  const beyond = await sceneList({ url: ACTRESS, offset: 6, limit: 5 });
  expect(beyond.items).toEqual([]);
});

test("without a limit every page is listed", async () => {
  mockProfile();
  const result = await sceneList({ url: ACTRESS });
  expect(result.items).toHaveLength(6);
});
