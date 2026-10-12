import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { performerFromBio, redirectTarget } from "../src/bio.ts";
import { listingUrl, pageSizeFrom, performerItemsFromPage, performerList } from "../src/listing.ts";
import { scrapePerformer } from "../src/bio.ts";
import { slugFromUrl } from "../src/http.ts";
import {
  countryFromPlace,
  heightCm,
  isoDateFromEnglish,
  measurementsFrom,
  weightKg,
} from "../src/values.ts";

const NET_TIMEOUT = 30_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("slugFromUrl reads the performer from profile, bio and feed URLs, with or without a locale", () => {
    expect(slugFromUrl("https://www.freeones.com/sara-jay/bio")).toBe("sara-jay");
    expect(slugFromUrl("https://www.freeones.com/sara-jay")).toBe("sara-jay");
    expect(slugFromUrl("https://www.freeones.com/de/sara-jay/feed")).toBe("sara-jay");
    expect(slugFromUrl("https://www.freeones.com/performers")).toBe("performers");
    expect(slugFromUrl("https://www.freeones.com/")).toBe("");
  });

  test("values: dates, measurements, heights and countries", () => {
    expect(isoDateFromEnglish("November 14, 1977")).toBe("1977-11-14");
    expect(isoDateFromEnglish("N/A")).toBeUndefined();
    expect(measurementsFrom("81", "E", "71 cm", "111 cm")).toBe("81E-71-111 cm");
    expect(measurementsFrom("", "", "", "")).toBeUndefined();
    expect(heightCm("12 cm")).toBeUndefined();
    expect(heightCm("154 cm")).toBe(154);
    expect(weightKg("67 kg")).toBe(67);
    expect(countryFromPlace("Cincinnati , OH - Ohio , United States")).toBe("United States");
  });

  test("redirectTarget unwraps the affiliate redirect of an official website", () => {
    expect(redirectTarget("http://refer.ccbill.com/cgi-bin/clicks.cgi?CA=925830-0000&amp;PA=1757651&amp;HTML=http://www.sarajay.com")).toBe(
      "http://www.sarajay.com"
    );
    expect(redirectTarget("https://example.com/")).toBe("https://example.com/");
  });

  test("listingUrl browses A to Z or searches by name, with 1-based pages", () => {
    expect(listingUrl(undefined, 2, 24)).toBe(
      "https://www.freeones.com/performers?s=name&o=asc&l=24&p=2"
    );
    expect(listingUrl("sara jay", 1, 10)).toBe("https://www.freeones.com/performers?q=sara+jay&l=10&p=1");
    expect(pageSizeFrom(undefined)).toBe(24);
    expect(pageSizeFrom(500)).toBe(100);
  });
});

describe("performer list page (offline)", () => {
  test("the A-Z page lists 24 performers with names, slugs and portraits", () => {
    const items = performerItemsFromPage(parse(loadFixture("performers.html")));
    expect(items.length).toBe(24);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(24);
    const first = items[0];
    expect(first.media_type).toBe("performer");
    expect(first.title).toBe("1 Cook 1 Barmaid");
    expect(first.external_id).toBe("freeones-1-cook-1-barmaid");
    expect(first.source_url).toBe("https://www.freeones.com/1-cook-1-barmaid");
    expect(first.poster_path).toStartWith("https://thumbs.freeones.com/photo/");
  });

  test("cards without a portrait keep their name and get no poster", () => {
    const items = performerItemsFromPage(parse(loadFixture("performers.html")));
    const claimable = items.find((item) => item.external_id === "freeones-a-j-cook");
    expect(claimable?.title).toBe("A.J. Cook");
    expect(claimable?.poster_path).toBeUndefined();
  });

  test("a name search and the second page each yield their own performers", () => {
    const search = performerItemsFromPage(parse(loadFixture("performers-search-sara.html")));
    expect(search.length).toBe(24);
    const second = performerItemsFromPage(parse(loadFixture("performers-page2.html")));
    const firstIds = new Set(performerItemsFromPage(parse(loadFixture("performers.html"))).map((item) => item.external_id));
    expect(second.length).toBe(24);
    expect(second.some((item) => firstIds.has(item.external_id))).toBe(false);
  });
});

describe("performer bio page (offline)", () => {
  const sara = performerFromBio(loadFixture("bio-sara-jay.html"), "sara-jay");

  test("identity and names", () => {
    expect(sara.type).toBe("performer");
    expect(sara.external_id).toBe("freeones-sara-jay");
    expect(sara.source_url).toBe("https://www.freeones.com/sara-jay");
    expect(sara.name).toBe("Sara Jay");
    expect(sara.aliases).toEqual(["Sarah Jay", "Sarah Jaye"]);
  });

  test("personal details", () => {
    expect(sara.birthdate).toBe("1977-11-14");
    expect(sara.death_date).toBeUndefined();
    expect(sara.country).toBe("United States");
    expect(sara.ethnicity).toBe("Caucasian");
    expect(sara.eye_color).toBe("Hazel");
    expect(sara.hair_color).toBe("Blonde");
    expect(sara.weight_kg).toBe(67);
    expect(sara.career_length).toBe("2000-2024");
    expect(sara.fake_tits).toBe("Fake");
    expect(sara.measurements).toBe("81E-71-111 cm");
  });

  test("the site's 12 cm placeholder height is not stored", () => {
    expect(sara.height_cm).toBeUndefined();
  });

  test("tattoos and piercings keep their locations", () => {
    expect(sara.tattoos).toStartWith("Yes: Oriental char on back of neck");
    expect(sara.piercings).toBe("Yes: Tongue; industrial");
  });

  test("biography, website, socials and portrait", () => {
    expect(sara.details).toStartWith("As an AVN and UrbanX Hall of Fame inductee");
    expect(sara.details).toContain("\n\n");
    expect(sara.url).toBe("http://www.sarajay.com");
    expect(sara.twitter).toBe("http://twitter.com/SaraJayXXX");
    expect(sara.instagram).toBe("http://instagram.com/officialsarajay");
    expect(sara.image_path).toStartWith("https://thumbs.freeones.com/photo/");
  });

  test("a page without a height placeholder keeps the real height", () => {
    const lexi = performerFromBio(loadFixture("bio-lexi-belle.html"), "lexi-belle");
    expect(lexi.name).toBe("Lexi Belle");
    expect(lexi.aliases).toEqual(["Lexi Williams", "Lexi Bell", "Lexi B."]);
    expect(lexi.instagram).toBe("http://instagram.com/omgitslexi");
    expect(lexi.twitter).toBe("http://twitter.com/OMGitsLexi");
    expect(lexi.measurements).toBe("81C-60-83 cm");
  });
});

// ---------------------------------------------------------------------------
// Live integration against freeones.com.

describe("performer:list (live)", () => {
  test(
    "returns a bounded page of performer stubs",
    async () => {
      const { items } = await performerList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) {
        expect(item.media_type).toBe("performer");
        expect(item.external_id).toStartWith("freeones-");
        expect(item.source_url).toStartWith("https://www.freeones.com/");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "offset pages do not repeat the first page",
    async () => {
      const first = await performerList({ limit: 5 });
      const second = await performerList({ limit: 5, offset: 5 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(second.items.length).toBe(5);
      expect(second.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test(
    "a query returns performers whose names match it",
    async () => {
      const { items } = await performerList({ query: "sara jay", limit: 5 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.some((item) => item.title === "Sara Jay")).toBe(true);
    },
    2 * NET_TIMEOUT
  );
});

describe("performer:find (live)", () => {
  test(
    "scrapes a listed performer into one performer record",
    async () => {
      const { items } = await performerList({ limit: 1 });
      const results = await scrapePerformer(items[0].source_url);
      expect(results.length).toBe(1);
      const [result] = results;
      if (result.type !== "performer") throw new Error("expected a performer");
      expect(result.performer.external_id).toBe(items[0].external_id);
      expect(result.performer.name.trim()).not.toBe("");
    },
    2 * NET_TIMEOUT
  );
});
