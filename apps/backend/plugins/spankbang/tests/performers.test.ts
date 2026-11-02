import { test, expect } from "bun:test";
import {
  parsePerformerPage,
  performerIdentify,
  profileLinks,
  type PageFetcher,
} from "../src/performers.ts";

const fixtures = `${import.meta.dir}/fixtures`;
const profileHtml = await Bun.file(`${fixtures}/pornstar-ariella-ferrera.html`).text();
const rileySuggestions = await Bun.file(`${fixtures}/suggest-riley-reid.html`).text();
const creatorOnlySuggestions = await Bun.file(`${fixtures}/suggest-creator-only.html`).text();

const profileUrl = "https://spankbang.com/1l/pornstar/ariella+ferrera/";

/** Serves the Ariella profile for her URL and the given suggestions for the search endpoint; records every URL asked. */
function fakeSite(suggestions: string): { fetchHtml: PageFetcher; asked: string[] } {
  const asked: string[] = [];
  const fetchHtml: PageFetcher = async (url) => {
    asked.push(url);
    if (url.includes("/api/search/search_bar_suggest")) return suggestions;
    if (url === profileUrl) return profileHtml;
    throw new Error(`unexpected fetch ${url}`);
  };
  return { fetchHtml, asked };
}

test("parsePerformerPage fills the profile from the page", () => {
  const performer = parsePerformerPage(profileHtml, profileUrl);

  expect(performer).toMatchObject({
    type: "performer",
    external_id: "spankbang-pornstar-1l",
    source_url: profileUrl,
    name: "Ariella Ferrera",
    country: "Colombia",
    ethnicity: "Latin",
    hair_color: "Brown",
    height_cm: 164,
    weight_kg: 52,
    image_path: "https://spankbang.com/pornstarimg/f/57-250.jpg",
  });
  expect(performer?.details).toStartWith("Ariella Ferrera, a sultry Colombian beauty born in Medellín");
  expect(performer?.birthdate).toBeUndefined();
  expect(performer?.gender).toBeUndefined();
});

test("parsePerformerPage returns null for a page without a profile", () => {
  expect(parsePerformerPage("<html><body><h1>Not found</h1></body></html>", profileUrl)).toBeNull();
});

test("profileLinks reads pornstar links from the suggestions, skipping creators and searches", () => {
  expect(profileLinks(rileySuggestions)).toEqual(["https://spankbang.com/1a/pornstar/riley+reid/"]);
  expect(profileLinks(creatorOnlySuggestions)).toEqual([]);
});

test("performerIdentify opens the profile named by a known URL of this site", async () => {
  const { fetchHtml, asked } = fakeSite("");
  const result = await performerIdentify(
    { name: "Ariella Ferrera", urls: ["https://example.com/x", profileUrl] },
    fetchHtml,
  );

  expect(result.performer?.external_id).toBe("spankbang-pornstar-1l");
  expect(asked).toEqual([profileUrl]);
});

test("performerIdentify returns nothing when the profile is someone else", async () => {
  const { fetchHtml } = fakeSite("");
  const result = await performerIdentify({ name: "Mia Malkova", urls: [profileUrl] }, fetchHtml);

  expect(result).toEqual({});
});

test("performerIdentify matches on an alias", async () => {
  const { fetchHtml } = fakeSite("");
  const result = await performerIdentify(
    { name: "Mia Malkova", aliases: ["Ariella Ferrera"], urls: [profileUrl] },
    fetchHtml,
  );

  expect(result.performer?.name).toBe("Ariella Ferrera");
});

test("performerIdentify finds the profile through the suggestions", async () => {
  const profile = profileHtml.replace(/Ariella Ferrera/g, "Riley Reid");
  const rileyUrl = "https://spankbang.com/1a/pornstar/riley+reid/";
  const fetchHtml: PageFetcher = async (url) => {
    if (url.includes("/api/search/search_bar_suggest")) return rileySuggestions;
    if (url === rileyUrl) return profile;
    throw new Error(`unexpected fetch ${url}`);
  };

  const result = await performerIdentify({ name: "Riley Reid" }, fetchHtml);
  expect(result.performer?.name).toBe("Riley Reid");
});

test("performerIdentify does not open unrelated suggested profiles", async () => {
  const { fetchHtml, asked } = fakeSite(rileySuggestions);
  const result = await performerIdentify({ name: "Mia Malkova" }, fetchHtml);

  expect(result).toEqual({});
  expect(asked.every((url) => url.includes("search_bar_suggest"))).toBe(true);
});

test("performerIdentify finds nothing when only a creator matches", async () => {
  const { fetchHtml } = fakeSite(creatorOnlySuggestions);
  expect(await performerIdentify({ name: "Mia Malkova" }, fetchHtml)).toEqual({});
});

test("performerIdentify keeps a name match when the page has no birthdate to contradict", async () => {
  const { fetchHtml } = fakeSite("");
  const result = await performerIdentify(
    { name: "Ariella Ferrera", birthdate: "1978-12-29", urls: [profileUrl] },
    fetchHtml,
  );
  expect(result.performer).toBeDefined();
});
