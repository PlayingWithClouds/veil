import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { identifyPerformer, isProfileUrl, parseProfile, pornstarUrlsFromSuggestions, type ProfileFetcher } from "../src/profile.ts";

/** Reads a saved page or answer from the fixtures folder. */
function fixture(name: string): string {
  return readFileSync(new URL(`./fixtures/${name}`, import.meta.url), "utf8");
}

const PROFILE_URL = "https://www.xvideos.com/pornstars/mia_malkova-model";

/** A site that knows Mia Malkova's pornstar page and nothing else; records the pages it served. */
function fakeSite(served: string[] = []): ProfileFetcher {
  return {
    suggest: async () => JSON.parse(fixture("suggest.json")),
    page: async (url) => {
      served.push(url);
      return { html: fixture("pornstar-profile.html"), url };
    },
  };
}

describe("profile page", () => {
  const performer = parseProfile(fixture("pornstar-profile.html"), PROFILE_URL);

  test("reads name, gender, country and picture", () => {
    expect(performer).toMatchObject({
      type: "performer",
      external_id: "xvideos-pornstars-mia_malkova-model",
      source_url: PROFILE_URL,
      name: "Mia Malkova",
      gender: "female",
      country: "USA",
    });
    expect(performer?.image_path).toStartWith("https://profile-cdn77.xvideos-cdn.com/");
    expect(performer?.image_path).toContain("pp_big.jpg");
    expect(performer?.images).toEqual([{ type: "profile", file_path: performer?.image_path as string }]);
  });

  test("a page without a profile heading is not a performer", () => {
    expect(parseProfile("<html><body>nothing</body></html>", PROFILE_URL)).toBeUndefined();
  });

  test("recognizes profile urls", () => {
    expect(isProfileUrl("https://www.xvideos.com/pornstars/nicole_aniston")).toBe(true);
    expect(isProfileUrl("https://de.xvideos.com/profiles/someone")).toBe(true);
    expect(isProfileUrl("https://www.xvideos.com/video.abc/x")).toBe(false);
    expect(isProfileUrl("https://www.xnxx.com/pornstar/x")).toBe(false);
  });
});

describe("identify", () => {
  test("lists suggested pornstars named exactly as asked", () => {
    const answer = JSON.parse(fixture("suggest.json"));
    expect(pornstarUrlsFromSuggestions(answer, { name: "mia malkova" })).toEqual([PROFILE_URL]);
    expect(pornstarUrlsFromSuggestions(answer, { name: "Mia Khalifa" })).toEqual([]);
  });

  test("returns the profile of a matching name", async () => {
    const result = await identifyPerformer({ name: "Mia Malkova" }, fakeSite());
    expect(result.performer?.name).toBe("Mia Malkova");
    expect(result.performer?.source_url).toBe(PROFILE_URL);
  });

  test("matches through an alias", async () => {
    const answer = await identifyPerformer({ name: "Mia M", aliases: ["Mia Malkova"] }, fakeSite());
    expect(answer.performer?.name).toBe("Mia Malkova");
  });

  test("returns nothing for another name", async () => {
    expect(await identifyPerformer({ name: "Mia Khalifa" }, fakeSite())).toEqual({});
  });

  test("reads a known pornstar url of this site without searching", async () => {
    const served: string[] = [];
    const site: ProfileFetcher = { ...fakeSite(served), suggest: async () => { throw new Error("no search expected"); } };
    const result = await identifyPerformer({ name: "Mia Malkova", urls: [PROFILE_URL, "https://www.eporner.com/pornstar/mia-malkova/"] }, site);
    expect(result.performer?.name).toBe("Mia Malkova");
    expect(served).toEqual([PROFILE_URL]);
  });

  test("a known url of someone else's page is refused", async () => {
    const result = await identifyPerformer({ name: "Riley Reid", urls: [PROFILE_URL] }, fakeSite());
    expect(result).toEqual({});
  });

  test("a pornstar slug that redirects to a user account is refused", async () => {
    const site: ProfileFetcher = {
      ...fakeSite(),
      page: async (url) => ({ html: fixture("pornstar-profile.html"), url: url.replace("/pornstars/", "/profiles/") }),
    };
    expect(await identifyPerformer({ name: "Mia Malkova" }, site)).toEqual({});
  });
});
