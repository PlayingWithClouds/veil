import { test, expect, afterEach } from "bun:test";
import { readFileSync } from "node:fs";
import { parse } from "node-html-parser";
import { sceneList, profilePageUrl } from "../src/scenes.ts";
import { parseProfileCards } from "../src/listing.ts";

// A trimmed profile page: the promoted block (js-forced-layout, same on every page) then the results.
const profilePage = readFileSync(new URL("./fixtures/profile-videos-ariella-ferrera.html", import.meta.url), "utf8");
const PROFILE = "https://spankbang.com/1l/pornstar/ariella+ferrera";
const originalFetch = globalThis.fetch;
const requested: string[] = [];

// Past the last page SpankBang shows only the promoted block.
function promotedOnly(): string {
  return profilePage.slice(0, profilePage.indexOf('<div class="mb-6">')) + "</body></html>";
}

// Serves the fixture for page 1, re-labelled results for page 2, promoted-only afterwards.
function mockProfile(): void {
  requested.length = 0;
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = String(input);
    requested.push(url);
    if (url === `${PROFILE}/`) return new Response(profilePage);
    if (url === `${PROFILE}/2/`) {
      return new Response(profilePage.replace(/\/(a533i|a4mit|a4mul)\//g, (_, code) => `/p2${code}/`));
    }
    return new Response(promotedOnly());
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("profile cards leave out the promoted block", () => {
  const cards = parseProfileCards(parse(profilePage));
  expect(cards.map((card) => card.external_id)).toEqual(["spankbang-a533i", "spankbang-a4mit", "spankbang-a4mul"]);
  expect(cards[0].title).toContain("Ariella Ferrera");
  expect(parseProfileCards(parse(promotedOnly()))).toEqual([]);
});

test("profile page URLs append the page number after the first page", () => {
  const pageUrl = profilePageUrl(`${PROFILE}/?q=hd`);
  expect(pageUrl(1)).toBe(`${PROFILE}/`);
  expect(pageUrl(3)).toBe(`${PROFILE}/3/`);
});

test("offset continues on the next page and stops past the end", async () => {
  mockProfile();
  const first = await sceneList({ url: `${PROFILE}/`, limit: 3 });
  expect(first.items).toHaveLength(3);
  expect(requested).toEqual([`${PROFILE}/`]);

  const second = await sceneList({ url: `${PROFILE}/`, offset: 3, limit: 2 });
  expect(second.items.map((item) => item.external_id)).toEqual(["spankbang-p2a533i", "spankbang-p2a4mit"]);

  const beyond = await sceneList({ url: `${PROFILE}/`, offset: 6, limit: 5 });
  expect(beyond.items).toEqual([]);
});

test("without a limit every page is listed", async () => {
  mockProfile();
  const result = await sceneList({ url: `${PROFILE}/` });
  expect(result.items).toHaveLength(6);
});
