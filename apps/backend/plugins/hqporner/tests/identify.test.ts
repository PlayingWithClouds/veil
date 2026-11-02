import { test, expect, describe, afterEach } from "bun:test";
import { identifyPerformer, slugFromName } from "../src/identify.ts";

const originalFetch = globalThis.fetch;
const miaPage = await Bun.file(`${import.meta.dir}/fixtures/actress-mia-malkova.html`).text();
const missingPage = await Bun.file(`${import.meta.dir}/fixtures/actress-missing.html`).text();
const girlsPage = await Bun.file(`${import.meta.dir}/fixtures/girls.html`).text();

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/**
 * Serves the saved pages like the site does: /girls, Mia's page, and the
 * "Can't find" page with status 200 for every other actress slug.
 */
function mockSite(): string[] {
  const requested: string[] = [];
  globalThis.fetch = (async (input: string | URL | Request) => {
    const url = String(input);
    requested.push(url);
    if (url === "https://hqporner.com/girls") return new Response(girlsPage);
    if (url === "https://hqporner.com/actress/mia-malkova") return new Response(miaPage);
    if (url.startsWith("https://hqporner.com/actress/")) return new Response(missingPage);
    return new Response("", { status: 404 });
  }) as typeof fetch;
  return requested;
}

describe("identifyPerformer", () => {
  test("slugFromName hyphenates the folded name", () => {
    expect(slugFromName("Mia Malkova")).toBe("mia-malkova");
    expect(slugFromName("Zoë O'Neil")).toBe("zoe-o-neil");
  });

  test("the name's slug gives the actress page for a matching name", async () => {
    mockSite();
    const { performer } = await identifyPerformer({ name: "Mia Malkova" });
    expect(performer?.name).toBe("Mia Malkova");
    expect(performer?.external_id).toBe("hqporner-actress-mia-malkova");
    expect(performer?.source_url).toBe("https://hqporner.com/actress/mia-malkova");  });

  test("a known profile URL is tried before the name's slug", async () => {
    const requested = mockSite();
    const { performer } = await identifyPerformer({
      name: "Mia Malkova",
      urls: ["https://www.freeones.com/mia-malkova", "https://hqporner.com/actress/mia-malkova"],
    });
    expect(performer?.name).toBe("Mia Malkova");
    expect(requested).toEqual(["https://hqporner.com/actress/mia-malkova"]);
  });

  test("a known URL that is another person's page is not returned", async () => {
    mockSite();
    expect(await identifyPerformer({ name: "Riley Reid", urls: ["https://hqporner.com/actress/mia-malkova"] })).toEqual({});
  });

  test("a slug the site has no actress for is not guessed from the name", async () => {
    mockSite();
    expect(await identifyPerformer({ name: "Nobody Known" })).toEqual({});
  });
});
