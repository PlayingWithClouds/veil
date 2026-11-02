import { test, expect, describe, afterEach } from "bun:test";
import { identifyPerformer, pornstarsFromSuggestHtml } from "../src/identify.ts";

const originalFetch = globalThis.fetch;
const suggestHtml = await Bun.file(`${import.meta.dir}/fixtures/suggest-mia-malkova.html`).text();
const profileHtml = await Bun.file(`${import.meta.dir}/fixtures/pornstar-mia-malkova.html`).text();
const PROFILE_URL = "https://www.eporner.com/pornstar/mia-malkova-oPgtJ/";

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Serves the saved suggest fragment and profile page, 404s anything else, and records the requests. */
function mockSite(): string[] {
  const requested: string[] = [];
  globalThis.fetch = (async (input: string | URL | Request) => {
    const url = String(input);
    requested.push(url);
    if (url.includes("/suggest/")) return new Response(suggestHtml);
    if (url === PROFILE_URL) return new Response(profileHtml);
    return new Response("", { status: 404 });
  }) as typeof fetch;
  return requested;
}

describe("pornstarsFromSuggestHtml", () => {
  test("lists the pornstar rows with their suffixed profile URLs", () => {
    expect(pornstarsFromSuggestHtml(suggestHtml)).toContainEqual({ url: PROFILE_URL, name: "Mia Malkova" });
  });
});

describe("identifyPerformer", () => {
  test("returns the filled profile for a matching name", async () => {
    mockSite();
    const { performer } = await identifyPerformer({ name: "Mia Malkova" });
    expect(performer?.source_url).toBe(PROFILE_URL);
    expect(performer?.external_id).toBe("eporner-ps-mia-malkova-oPgtJ");
    expect(performer?.name).toBe("Mia Malkova");
    expect(performer?.country).toBe("United States");
    expect(performer?.height_cm).toBe(170);
    expect(performer?.aliases).toContain("Madison Clover");
  });

  test("matches on an alias the library knows", async () => {
    mockSite();
    const { performer } = await identifyPerformer({ name: "Mia Malkova", aliases: ["Mia Bliss"] });
    expect(performer?.source_url).toBe(PROFILE_URL);
  });

  test("a known profile URL of this site is read without the search box", async () => {
    const requested = mockSite();
    const { performer } = await identifyPerformer({ name: "Mia Malkova", urls: [PROFILE_URL, "https://x.com/mia"] });
    expect(performer?.source_url).toBe(PROFILE_URL);
    expect(requested[0]).toBe(PROFILE_URL);
  });

  test("a different person is not returned", async () => {
    mockSite();
    expect(await identifyPerformer({ name: "Riley Reid" })).toEqual({});
  });

  test("a known URL whose profile belongs to someone else is not returned", async () => {
    mockSite();
    expect(await identifyPerformer({ name: "Riley Reid", urls: [PROFILE_URL] })).toEqual({});
  });
});
