import { test, expect, afterEach } from "bun:test";
import { crawlPage, pageListingUrl } from "../src/pages.ts";
import { sceneList } from "../src/scenes.ts";

// Any eporner video grid works as a page listing; reuse the related-videos one.
const grid = await Bun.file(`${import.meta.dir}/fixtures/related.html`).text();
const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Serves `pages` by URL and 404s everything else, recording each request. */
function mockFetch(pages: Record<string, string>): string[] {
  const requested: string[] = [];
  globalThis.fetch = (async (input: string | URL | Request) => {
    const url = String(input);
    requested.push(url);
    const body = pages[url];
    if (body === undefined) return new Response("", { status: 404 });
    return new Response(body, { status: 200 });
  }) as typeof fetch;
  return requested;
}

test("pageListingUrl paginates channels, pornstars and categories", () => {
  expect(pageListingUrl("https://www.eporner.com/channel/vixen/", 1)).toBe("https://www.eporner.com/channel/vixen/");
  expect(pageListingUrl("https://www.eporner.com/channel/vixen/", 2)).toBe("https://www.eporner.com/channel/vixen/2/");
  expect(pageListingUrl("https://eporner.com/pornstar/riley-reid", 3)).toBe(
    "https://www.eporner.com/pornstar/riley-reid/3/"
  );
  expect(pageListingUrl("https://www.eporner.com/cat/milf/", 1)).toBe("https://www.eporner.com/cat/milf/");
});

test("pageListingUrl lists a profile's uploads on a single page", () => {
  expect(pageListingUrl("https://www.eporner.com/profile/glazers/", 1)).toBe(
    "https://www.eporner.com/profile/glazers/videos/"
  );
  expect(pageListingUrl("https://www.eporner.com/profile/glazers/", 2)).toBe("");
});

test("pageListingUrl rejects other pages and sites", () => {
  expect(pageListingUrl("https://www.eporner.com/video-abc/title/", 1)).toBe("");
  expect(pageListingUrl("https://xhamster.com/channels/vixen", 1)).toBe("");
});

test("crawlPage follows pagination until a page fails", async () => {
  const requested = mockFetch({ "https://www.eporner.com/channel/vixen/": grid });
  const items = await crawlPage("https://www.eporner.com/channel/vixen/", 30);

  expect(items.map((item) => item.external_id)).toEqual([
    "eporner-mklZZn2M2ea",
    "eporner-qMt524baxFN",
    "eporner-qNitFgZULLc",
  ]);
  expect(requested).toEqual(["https://www.eporner.com/channel/vixen/", "https://www.eporner.com/channel/vixen/2/"]);
});

test("crawlPage fails when the first page cannot be loaded", async () => {
  mockFetch({});
  await expect(crawlPage("https://www.eporner.com/channel/gone/", 30)).rejects.toThrow("could not load");
  await expect(crawlPage("https://www.eporner.com/video-abc/title/", 30)).rejects.toThrow("cannot list");
});

test("sceneList with a url lists that page, sliced to the limit", async () => {
  mockFetch({ "https://www.eporner.com/pornstar/riley-reid/": grid });
  const result = await sceneList({ url: "https://www.eporner.com/pornstar/riley-reid/", limit: 2 });

  expect(result.items.map((item) => item.external_id)).toEqual(["eporner-mklZZn2M2ea", "eporner-qMt524baxFN"]);
});
