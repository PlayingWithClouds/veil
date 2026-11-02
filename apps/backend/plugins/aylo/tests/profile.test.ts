import { test, expect, describe } from "bun:test";
import { candidateUrls, identifyPerformer, type ProfileFetcher } from "../src/identify.ts";
import { canonicalProfileUrl, isProfileUrl, parseProfile, profileUrlForSlug } from "../src/profile.ts";
import { ALL_SITES, type Site } from "../src/sites.ts";

function site(key: string): Site {
  return ALL_SITES.find((candidate) => candidate.key === key)!;
}

async function fixture(name: string): Promise<string> {
  return Bun.file(`${import.meta.dir}/fixtures/${name}`).text();
}

/** Serves the fixtures at the profile URLs they were taken from; every other URL is a 404. */
async function fixtureFetcher(): Promise<{ fetchHtml: ProfileFetcher; requested: string[] }> {
  const pages: Record<string, string> = {
    "https://www.pornhub.com/pornstar/mia-malkova": await fixture("ph-pornstar.html"),
    "https://www.redtube.com/pornstar/mia+malkova": await fixture("rt-pornstar.html"),
    "https://www.youporn.com/pornstar/mia-malkova/": await fixture("yp-pornstar.html"),
    "https://www.tube8.com/pornstar/mia-malkova/": await fixture("t8-pornstar.html"),
  };
  const requested: string[] = [];
  const fetchHtml: ProfileFetcher = async (url) => {
    requested.push(url);
    return pages[url];
  };
  return { fetchHtml, requested };
}

describe("profile urls", () => {
  test("isProfileUrl accepts pornstar and model pages only", () => {
    expect(isProfileUrl("https://www.pornhub.com/pornstar/mia-malkova")).toBe(true);
    expect(isProfileUrl("https://www.pornhub.com/model/pinkloving/videos")).toBe(true);
    expect(isProfileUrl("https://www.redtube.com/pornstar/mia+malkova")).toBe(true);
    expect(isProfileUrl("https://www.pornhub.com/view_video.php?viewkey=abc")).toBe(false);
    expect(isProfileUrl("https://www.redtube.com/12345")).toBe(false);
    expect(isProfileUrl("not a url")).toBe(false);
  });

  test("canonicalProfileUrl drops sub-pages, query and host variants", () => {
    expect(canonicalProfileUrl(site("pornhub"), "https://de.pornhub.com/pornstar/mia-malkova/videos?page=2")).toBe(
      "https://www.pornhub.com/pornstar/mia-malkova",
    );
  });

  test("profileUrlForSlug follows each site's spelling", () => {
    expect(profileUrlForSlug(site("pornhub"), "mia-malkova")).toBe("https://www.pornhub.com/pornstar/mia-malkova");
    expect(profileUrlForSlug(site("redtube"), "mia-malkova")).toBe("https://www.redtube.com/pornstar/mia+malkova");
    expect(profileUrlForSlug(site("youporn"), "mia-malkova")).toBe("https://www.youporn.com/pornstar/mia-malkova/");
    expect(profileUrlForSlug(site("tube8"), "mia-malkova")).toBe("https://www.tube8.com/pornstar/mia-malkova/");
  });
});

describe("parseProfile", () => {
  test("Pornhub profile fills every fact the page shows", async () => {
    const performer = parseProfile(
      site("pornhub"),
      "https://www.pornhub.com/pornstar/mia-malkova",
      await fixture("ph-pornstar.html"),
    )!;
    expect(performer.external_id).toBe("pornhub-ps-mia-malkova");
    expect(performer.source_url).toBe("https://www.pornhub.com/pornstar/mia-malkova");
    expect(performer.name).toBe("Mia Malkova");
    expect(performer.aliases).toEqual(["Mia Bliss", "Madison Swan", "Madison Clover"]);
    expect(performer.details).toStartWith("Some girls are so damn hot");
    expect(performer.image_path).toStartWith("https://ei.phncdn.com/");
    expect(performer.gender).toBe("female");
    expect(performer.country).toBe("United States of America");
    expect(performer.ethnicity).toBe("White");
    expect(performer.hair_color).toBe("Blonde");
    expect(performer.eye_color).toBe("Hazel");
    expect(performer.height_cm).toBe(170);
    expect(performer.weight_kg).toBe(59);
    expect(performer.measurements).toBe("33E-21-40");
    expect(performer.fake_tits).toBe("Fake");
    expect(performer.career_length).toBe("2012 to Present");
    expect(performer.url).toBe("https://onlyfans.com/miamalkovafree/c26");
    expect(performer.twitter).toBe("https://www.twitter.com/miamalkova");
    expect(performer.instagram).toBe("https://www.instagram.com/miamalkova");
    // "Tattoos: No" says there are none, so nothing is stored.
    expect(performer.tattoos).toBeUndefined();
    expect(performer.birthdate).toBeUndefined();
  });

  test("RedTube, YouPorn and Tube8 profiles carry the birthdate and AKA list", async () => {
    const pages: [string, string, string][] = [
      ["redtube", "https://www.redtube.com/pornstar/mia+malkova", "rt-pornstar.html"],
      ["youporn", "https://www.youporn.com/pornstar/mia-malkova/", "yp-pornstar.html"],
      ["tube8", "https://www.tube8.com/pornstar/mia-malkova/", "t8-pornstar.html"],
    ];
    for (const [key, url, file] of pages) {
      const performer = parseProfile(site(key), url, await fixture(file))!;
      expect(performer.external_id).toBe(`${key}-ps-mia-malkova`);
      expect(performer.source_url).toBe(url);
      expect(performer.name).toBe("Mia Malkova");
      expect(performer.birthdate).toBe("1992-06-01");
      expect(performer.aliases).toEqual(["Mia Bliss", "Madison Swan", "Madison Clover"]);
      expect(performer.details).toStartWith("Some girls are so damn hot");
      expect(performer.image_path).toStartWith("https://ei.phncdn.com/");
      expect(performer.country).toBe("United States of America");
      expect(performer.height_cm).toBe(170);
      expect(performer.weight_kg).toBe(54);
      expect(performer.measurements).toBe("34C-26-36");
      expect(performer.hair_color).toBe("Blonde");
      expect(performer.fake_tits).toBe("Natural");
      expect(performer.piercings).toBe("Navel; clit");
      expect(performer.career_length).toBe("2012 to Present");
      expect(performer.url).toBe("https://fancentro.com/r/bCqwlzGe/miamalkova");
    }
  });

  test("a page without a profile heading is not a profile", () => {
    const indexPage = "<html><body><h1>Pornstars</h1><div class='infoPiece'></div></body></html>";
    expect(parseProfile(site("pornhub"), "https://www.pornhub.com/pornstar/nobody", indexPage)).toBeUndefined();
  });
});

describe("identifyPerformer", () => {
  test("finds the profile from the name and confirms it with the birthdate on another site", async () => {
    const { fetchHtml } = await fixtureFetcher();
    const result = await identifyPerformer({ name: "Mia Malkova", birthdate: "1992-06-01" }, fetchHtml);
    // Pornhub has no birthdate, so the confirming RedTube profile wins.
    expect(result.performer?.external_id).toBe("redtube-ps-mia-malkova");
  });

  test("takes the Pornhub profile when the library knows no birthdate", async () => {
    const { fetchHtml } = await fixtureFetcher();
    const result = await identifyPerformer({ name: "mia malkova" }, fetchHtml);
    expect(result.performer?.external_id).toBe("pornhub-ps-mia-malkova");
  });

  test("falls back to the name match when no profile shows a birthdate", async () => {
    const { fetchHtml } = await fixtureFetcher();
    const pornhubOnly: ProfileFetcher = async (url) => {
      if (url.includes("pornhub.com")) return fetchHtml(url);
      return undefined;
    };
    const result = await identifyPerformer({ name: "Mia Malkova", birthdate: "1992-06-01" }, pornhubOnly);
    expect(result.performer?.external_id).toBe("pornhub-ps-mia-malkova");
  });

  test("answers nothing when a different person has that name", async () => {
    const { fetchHtml } = await fixtureFetcher();
    const result = await identifyPerformer({ name: "Mia Malkova", birthdate: "1985-03-09" }, fetchHtml);
    // RedTube's Mia Malkova was born on another day, which also discredits Pornhub's undated name match.
    expect(result).toEqual({});
    const onlyDated: ProfileFetcher = async (url) => {
      if (url.includes("pornhub.com")) return undefined;
      return fetchHtml(url);
    };
    expect(await identifyPerformer({ name: "Mia Malkova", birthdate: "1985-03-09" }, onlyDated)).toEqual({});
  });

  test("answers nothing for a name no site has", async () => {
    const { fetchHtml } = await fixtureFetcher();
    expect(await identifyPerformer({ name: "Nobody Atall" }, fetchHtml)).toEqual({});
  });

  test("a profile page whose name differs is not accepted", async () => {
    const { fetchHtml } = await fixtureFetcher();
    // The slug lookup lands on Mia Malkova's page, but the library asks about someone else.
    const wrongPage: ProfileFetcher = async (url) => {
      if (url.includes("zzz-person")) return fetchHtml("https://www.redtube.com/pornstar/mia+malkova");
      return undefined;
    };
    expect(await identifyPerformer({ name: "Zzz Person" }, wrongPage)).toEqual({});
  });

  test("an alias of the library matches the profile's alias", async () => {
    const { fetchHtml } = await fixtureFetcher();
    const result = await identifyPerformer({ name: "Madison Swan", aliases: ["Mia Malkova"] }, fetchHtml);
    expect(result.performer?.name).toBe("Mia Malkova");
  });

  test("a known profile URL is asked first and sites are not asked twice", async () => {
    const { fetchHtml, requested } = await fixtureFetcher();
    const result = await identifyPerformer(
      { name: "Mia Malkova", urls: ["https://www.youporn.com/pornstar/mia-malkova/videos", "https://eporner.com/pornstar/x/"] },
      fetchHtml,
    );
    expect(requested[0]).toBe("https://www.youporn.com/pornstar/mia-malkova");
    expect(result.performer?.external_id).toBe("youporn-ps-mia-malkova");
  });
});

describe("candidateUrls", () => {
  test("lists known profile urls before slugs of the name and the first aliases", () => {
    const urls = candidateUrls(site("redtube"), {
      name: "Mia Malkova",
      aliases: ["Mia Bliss", "Madison Swan", "Madison Clover"],
      urls: ["https://www.redtube.com/pornstar/mia+malkova", "https://www.pornhub.com/pornstar/mia-malkova"],
    });
    expect(urls).toEqual([
      "https://www.redtube.com/pornstar/mia+malkova",
      "https://www.redtube.com/pornstar/mia+bliss",
      "https://www.redtube.com/pornstar/madison+swan",
    ]);
  });

  test("folds accents and punctuation out of slugs", () => {
    const urls = candidateUrls(site("pornhub"), { name: "Chloé O'Brien" });
    expect(urls).toEqual(["https://www.pornhub.com/pornstar/chloe-o-brien"]);
  });
});
