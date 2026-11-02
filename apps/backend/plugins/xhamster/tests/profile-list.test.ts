import { test, expect, afterEach } from "bun:test";
import { readFileSync } from "node:fs";
import { sceneList, profilePageUrl } from "../src/scenes.ts";

const profilePage = readFileSync(new URL("./fixtures/profile-mia-malkova.html", import.meta.url), "utf8");
const originalFetch = globalThis.fetch;
const requested: string[] = [];

// Serves the profile fixture for page 1, a re-labelled copy for page 2, 404 afterwards.
function mockProfile(): void {
  requested.length = 0;
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = String(input);
    requested.push(url);
    if (url === "https://xhamster.com/pornstars/mia-malkova") return new Response(profilePage);
    if (url === "https://xhamster.com/pornstars/mia-malkova/2") {
      return new Response(profilePage.replace(/xhE3x6w|xhHyYzE|13270207/g, (id) => `p2${id}`));
    }
    return new Response("not found", { status: 404 });
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("profile page URLs append the page number after the first page", () => {
  const pageUrl = profilePageUrl("https://xhamster.com/pornstars/mia-malkova/?x=1");
  expect(pageUrl(1)).toBe("https://xhamster.com/pornstars/mia-malkova");
  expect(pageUrl(3)).toBe("https://xhamster.com/pornstars/mia-malkova/3");
});

test("a pornstar url lists that page's videos with the normal external ids", async () => {
  mockProfile();
  const result = await sceneList({ url: "https://xhamster.com/pornstars/mia-malkova", limit: 3 });
  expect(result.items.map((item) => item.external_id)).toEqual([
    "xhamster-xhE3x6w",
    "xhamster-xhHyYzE",
    "xhamster-13270207",
  ]);
  expect(requested).toEqual(["https://xhamster.com/pornstars/mia-malkova"]);
});

test("offset continues on the next page and stops past the end", async () => {
  mockProfile();
  const second = await sceneList({ url: "https://xhamster.com/pornstars/mia-malkova", offset: 3, limit: 2 });
  expect(second.items.map((item) => item.external_id)).toEqual(["xhamster-p2xhE3x6w", "xhamster-p2xhHyYzE"]);

  const beyond = await sceneList({ url: "https://xhamster.com/pornstars/mia-malkova", offset: 6, limit: 5 });
  expect(beyond.items).toEqual([]);
});

test("without a limit every page is listed", async () => {
  mockProfile();
  const result = await sceneList({ url: "https://xhamster.com/pornstars/mia-malkova" });
  expect(result.items).toHaveLength(6);
});
