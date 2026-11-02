import { test, expect, afterEach } from "bun:test";
import { readFileSync } from "node:fs";
import { sceneList, profilePageUrl } from "../src/scenes.ts";

const profilePage = readFileSync(new URL("./fixtures/profile-videos-mia-malkova.html", import.meta.url), "utf8");
const emptyPage = "<!doctype html><html><body><div class=\"row video-list\"></div></body></html>";
const originalFetch = globalThis.fetch;
const requested: string[] = [];
const PROFILE = "https://www.tnaflix.com/profile/mia-malkova";

// Serves the fixture for page 1, a re-numbered copy for page 2, and (like the site) an empty 200 afterwards.
function mockProfile(): void {
  requested.length = 0;
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = String(input);
    requested.push(url);
    if (url === PROFILE) return new Response(profilePage);
    if (url === `${PROFILE}?page=2`) return new Response(profilePage.replace(/video(\d+)/g, "video9$1").replace(/data-vid="(\d+)"/g, 'data-vid="9$1"'));
    return new Response(emptyPage);
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("profile page URLs add ?page=N after the first page", () => {
  const pageUrl = profilePageUrl(`${PROFILE}/?x=1`);
  expect(pageUrl(1)).toBe(PROFILE);
  expect(pageUrl(4)).toBe(`${PROFILE}?page=4`);
});

test("a profile url lists that page's videos with the normal external ids", async () => {
  mockProfile();
  const result = await sceneList({ url: PROFILE, limit: 3 });
  expect(result.items.map((item) => item.external_id)).toEqual([
    "tnaflix-26253395",
    "tnaflix-26252455",
    "tnaflix-26246906",
  ]);
  expect(result.items[0].title).toContain("Mia Malkova");
  expect(result.items[0].source_url).toContain("/video26253395");
  expect(requested).toEqual([PROFILE]);
});

test("offset continues on the next page and stops past the end", async () => {
  mockProfile();
  const second = await sceneList({ url: PROFILE, offset: 3, limit: 2 });
  expect(second.items).toHaveLength(2);
  expect(second.items[0].external_id).toMatch(/^tnaflix-9\d+$/);

  const beyond = await sceneList({ url: PROFILE, offset: 6, limit: 5 });
  expect(beyond.items).toEqual([]);
});

test("without a limit every page is listed", async () => {
  mockProfile();
  const result = await sceneList({ url: PROFILE });
  expect(result.items).toHaveLength(6);
});
