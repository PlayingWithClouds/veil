import { test, expect, describe, afterEach } from "bun:test";
import { readFileSync } from "fs";
import { identifyPerformer, slugFromName } from "../src/identify.ts";

const originalFetch = globalThis.fetch;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Serves saved pages by URL, 404s everything else, and records the requests. */
function mockSite(pages: Record<string, string>): string[] {
  const requested: string[] = [];
  globalThis.fetch = (async (input: string | URL | Request) => {
    const url = String(input);
    requested.push(url);
    const body = pages[url];
    if (body === undefined) return new Response("", { status: 404 });
    return new Response(body);
  }) as typeof fetch;
  return requested;
}

const search = loadFixture("performers-search-sara.html");
const SEARCH_RILEY = "https://www.freeones.com/performers?q=Riley+Reid&l=10&p=1";
const SEARCH_LEXI = "https://www.freeones.com/performers?q=Lexi+Belle&l=10&p=1";
const pages = {
  "https://www.freeones.com/sara-jay/bio": loadFixture("bio-sara-jay.html"),
  "https://www.freeones.com/lexi-belle/bio": loadFixture("bio-lexi-belle.html"),
};

describe("identifyPerformer", () => {
  test("slugFromName hyphenates the folded name", () => {
    expect(slugFromName("Sara Jay")).toBe("sara-jay");
    expect(slugFromName("Zoë  O'Neil")).toBe("zoe-o-neil");
  });

  test("the name's slug gives the profile for a matching name", async () => {
    mockSite(pages);
    const { performer } = await identifyPerformer({ name: "Sara Jay" });
    expect(performer?.source_url).toBe("https://www.freeones.com/sara-jay");
    expect(performer?.external_id).toBe("freeones-sara-jay");
    expect(performer?.aliases).toContain("Sarah Jay");
  });

  test("an alias the profile lists is enough", async () => {
    mockSite(pages);
    const { performer } = await identifyPerformer({ name: "Sarah Jay", urls: ["https://www.freeones.com/sara-jay"] });
    expect(performer?.name).toBe("Sara Jay");
  });

  test("a known URL whose profile is another person is not returned", async () => {
    mockSite({ ...pages, [SEARCH_RILEY]: search });
    expect(await identifyPerformer({ name: "Riley Reid", urls: ["https://www.freeones.com/sara-jay/bio"] })).toEqual({});
  });

  test("a different birthdate refuses a name match", async () => {
    mockSite({ ...pages, [SEARCH_LEXI]: search });
    expect(await identifyPerformer({ name: "Lexi Belle", birthdate: "1900-01-01" })).toEqual({});
  });

  test("falls back to the site's search when the name's slug is not a profile", async () => {
    const requested = mockSite({ ...pages, "https://www.freeones.com/performers?q=Sarah+Jay&l=10&p=1": search });
    const { performer } = await identifyPerformer({ name: "Sarah Jay", aliases: ["Sara Jay"] });
    expect(performer?.source_url).toBe("https://www.freeones.com/sara-jay");
    expect(requested[0]).toBe("https://www.freeones.com/sarah-jay/bio");
  });

  test("an unknown name finds nothing", async () => {
    mockSite({ ...pages, "https://www.freeones.com/performers?q=Nobody+Known&l=10&p=1": search });
    expect(await identifyPerformer({ name: "Nobody Known" })).toEqual({});
  });
});
