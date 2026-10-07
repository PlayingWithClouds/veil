import { test, expect, afterEach } from "bun:test";
import { followablePageUrl, sceneList } from "../src/scenes.ts";

// Any hqporner video-card grid works as a page listing; reuse the related one.
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

test("followablePageUrl paginates actress and category pages", () => {
  expect(followablePageUrl("https://hqporner.com/actress/little-caprice", 1)).toBe(
    "https://hqporner.com/actress/little-caprice"
  );
  expect(followablePageUrl("https://hqporner.com/actress/little-caprice/4", 2)).toBe(
    "https://hqporner.com/actress/little-caprice/2"
  );
  expect(followablePageUrl("/category/milf", 3)).toBe("https://hqporner.com/category/milf/3");
});

test("followablePageUrl rejects videos and other sites", () => {
  expect(followablePageUrl("https://hqporner.com/hdporn/128024-yes.html", 1)).toBe("");
  expect(followablePageUrl("https://example.com/actress/little-caprice", 1)).toBe("");
});

test("sceneList with a url lists the page's cards until a page fails", async () => {
  const requested = mockFetch({ "https://hqporner.com/actress/little-caprice": grid });
  const result = await sceneList({ url: "https://hqporner.com/actress/little-caprice", limit: 40 });

  expect(result.items.length).toBeGreaterThan(0);
  expect(result.items.every((item) => item.external_id.startsWith("hqporner-"))).toBe(true);
  expect(requested).toEqual(["https://hqporner.com/actress/little-caprice", "https://hqporner.com/actress/little-caprice/2"]);
});

test("sceneList with an unlistable or unreachable url throws", async () => {
  mockFetch({});
  await expect(sceneList({ url: "https://hqporner.com/hdporn/1-x.html" })).rejects.toThrow("cannot list");
  await expect(sceneList({ url: "https://hqporner.com/category/gone" })).rejects.toThrow("no videos");
});
