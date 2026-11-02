import { test, expect, describe } from "bun:test";
import { siteByKey, type SiteConfig } from "../src/sites.ts";
import { parseBirthdate, parseHeightCm, parseProfile, parseWeightKg, withoutHeadingTail } from "../src/profile.ts";
import { fullestProfile, identifyWith, knownProfiles, slugGuesses, type ProfileFetcher } from "../src/identify.ts";

/** The configured site with this key; fails the test when missing. */
function site(key: string): SiteConfig {
  const found = siteByKey(key);
  if (!found) throw new Error(`no site ${key}`);
  return found;
}

/** A saved model page. */
async function fixture(name: string): Promise<string> {
  return Bun.file(`${import.meta.dir}/fixtures/${name}.html`).text();
}

describe("parseProfile", () => {
  test("reads an info list with aliases, birthdate, measurements and eye/hair color", async () => {
    const html = await fixture("profile-xgroovy");
    const performer = parseProfile(site("xgroovy"), "https://xgroovy.com/pornstars/blue-angel/", html);
    expect(performer).toMatchObject({
      type: "performer",
      external_id: "kvs-xgroovy-model-blue-angel",
      source_url: "https://xgroovy.com/pornstars/blue-angel/",
      name: "Blue Angel",
      aliases: ["Viva Style", "Vivien Piaf", "Vivien Toth", "Angel B", "Sissy Blake", "Leila", "Beatrice"],
      birthdate: "1988-06-21",
      country: "Hungary",
      measurements: "32A-23-35",
      height_cm: 170,
      weight_kg: 52,
      eye_color: "blue",
      hair_color: "blonde",
    });
    expect(performer.image_path).toMatch(/\/contents\/models\//);
  });

  test("reads a themed list with gender, biography and a lazy profile photo", async () => {
    const html = await fixture("profile-fapgoat");
    const performer = parseProfile(site("fapgoat"), "https://fapgoat.com/models/annabel-redd/", html);
    expect(performer).toMatchObject({
      name: "Annabel Redd",
      gender: "female",
      birthdate: "1997-06-16",
      country: "United States",
      height_cm: 162,
      weight_kg: 73,
      measurements: "30E-26-38",
      hair_color: "Red",
      eye_color: "Blue",
    });
    expect(performer.details).toContain("introverted reader");
  });

  test("reads JSON-LD Person: birthdate, height, description, image", async () => {
    const html = await fixture("profile-fuqster");
    const performer = parseProfile(site("fuqster"), "https://fuqster.com/models/alyx-star/", html);
    expect(performer).toMatchObject({
      name: "Alyx Star",
      birthdate: "1998-09-07",
      height_cm: 172,
      image_path: "https://cdnstatic.fuqster.com/contents/models/714/alyx-star.jpg?v=1.1",
    });
    expect(performer.details).toContain("American adult film actress");
  });

  test("reads imperial units, the official site as twitter and the bio from a title attribute", async () => {
    const html = await fixture("profile-fpo");
    const performer = parseProfile(site("fpo"), "https://www.fpo.xxx/models/remy-lacroix/", html);
    expect(performer).toMatchObject({
      name: "Remy LaCroix",
      country: "United States",
      height_cm: 157,
      weight_kg: 48,
      twitter: "https://twitter.com/remymeow",
    });
    expect(performer.details).toContain("five-man gang bang");
  });

  test("reads stat boxes and the cover biography", async () => {
    const html = await fixture("profile-severeporn");
    const performer = parseProfile(site("severeporn"), "https://severeporn.com/models/angela-white/", html);
    expect(performer).toMatchObject({ name: "Angela White", country: "Australia", height_cm: 163, weight_kg: 58 });
    expect(performer.details).toContain("Australian porn star");
  });

  test("a page that lists nothing yields only the name, with N/A and 0 left out", async () => {
    const html = await fixture("profile-bare");
    const performer = parseProfile(site("neporn"), "https://neporn.com/models/katie-kinz/", html);
    expect(performer.name).toBe("Katie Kinz");
    expect(performer.external_id).toBe("kvs-neporn-model-katie-kinz");
    expect(performer.country).toBeUndefined();
    expect(performer.height_cm).toBeUndefined();
    expect(performer.details).toBeUndefined();
  });
});

describe("value parsers", () => {
  test("parseBirthdate reads ISO, month-first and day-first dates and ignores partial ones", () => {
    expect(parseBirthdate("1993-07-17")).toBe("1993-07-17");
    expect(parseBirthdate("September 01, 2004, 22 y.o.")).toBe("2004-09-01");
    expect(parseBirthdate("15 January 1984")).toBe("1984-01-15");
    expect(parseBirthdate("2009")).toBeUndefined();
  });

  test("parseHeightCm reads cm, feet and inches and rejects placeholders", () => {
    expect(parseHeightCm("162cm")).toBe(162);
    expect(parseHeightCm("5'2\"")).toBe(157);
    expect(parseHeightCm("5 feet 4 inches (163 cm)")).toBe(163);
    expect(parseHeightCm("165")).toBe(165);
    expect(parseHeightCm("62cm (2ft)")).toBeUndefined();
  });

  test("parseWeightKg reads kg and pounds", () => {
    expect(parseWeightKg("73kg")).toBe(73);
    expect(parseWeightKg("105lb")).toBe(48);
    expect(parseWeightKg("128 pounds (58 kg)")).toBe(58);
  });

  test("withoutHeadingTail cuts the video-list suffix themes append to the name", () => {
    expect(withoutHeadingTail("Remy LaCroix's New Porn Videos")).toBe("Remy LaCroix");
    expect(withoutHeadingTail("Riley Reid xxx videos")).toBe("Riley Reid");
    expect(withoutHeadingTail("Alyx Star")).toBe("Alyx Star");
  });
});

describe("identifyWith", () => {
  const xgroovy = site("xgroovy");
  const fapgoat = site("fapgoat");

  /** Serves the saved pages by URL; any other URL is a missing page. */
  async function savedPages(): Promise<ProfileFetcher> {
    const pages: Record<string, string> = {
      "https://xgroovy.com/pornstars/blue-angel/": await fixture("profile-xgroovy"),
      "https://fapgoat.com/models/annabel-redd/": await fixture("profile-fapgoat"),
    };
    return async (candidate) => {
      const html = pages[candidate.url];
      if (html === undefined) return undefined;
      return { html, url: candidate.url };
    };
  }

  test("builds profile URLs from the name and aliases on each site", () => {
    const guesses = slugGuesses({ name: "Blue Angel", aliases: ["Vivien Piaf", "Blue Angel"] }, xgroovy);
    expect(guesses.map((guess) => guess.url)).toEqual([
      "https://xgroovy.com/pornstars/blue-angel/",
      "https://xgroovy.com/pornstars/vivien-piaf/",
    ]);
  });

  test("picks the model pages of this plugin's sites out of the known URLs", () => {
    const known = knownProfiles({
      name: "Blue Angel",
      urls: ["https://www.eporner.com/pornstar/blue-angel/", "https://xgroovy.com/pornstars/blue-angel/?x=1"],
    });
    expect(known.map((candidate) => candidate.site.key)).toEqual(["xgroovy"]);
  });

  test("returns the profile when the name matches", async () => {
    const result = await identifyWith({ name: "Blue Angel" }, [fapgoat, xgroovy], await savedPages());
    expect(result.performer?.external_id).toBe("kvs-xgroovy-model-blue-angel");
    expect(result.performer?.birthdate).toBe("1988-06-21");
  });

  test("matches through an alias of the library's performer", async () => {
    const result = await identifyWith({ name: "Blue Angel", aliases: ["Sissy Blake"] }, [xgroovy], await savedPages());
    expect(result.performer?.name).toBe("Blue Angel");
  });

  test("returns nothing when the page's name is another person", async () => {
    const result = await identifyWith({ name: "Annabel Redd" }, [xgroovy], async (candidate) => ({
      html: await fixture("profile-xgroovy"),
      url: candidate.url.replace("annabel-redd", "blue-angel"),
    }));
    expect(result).toEqual({});
  });

  test("returns nothing when the birthdates differ", async () => {
    const result = await identifyWith({ name: "Blue Angel", birthdate: "1990-01-01" }, [xgroovy], await savedPages());
    expect(result).toEqual({});
  });

  test("returns nothing when the site has no page, or redirected elsewhere", async () => {
    expect(await identifyWith({ name: "Nobody Known" }, [xgroovy, fapgoat], await savedPages())).toEqual({});
    const redirected = await identifyWith({ name: "Blue Angel" }, [xgroovy], async (candidate) => ({
      html: await fixture("profile-xgroovy"),
      url: "https://xgroovy.com/",
    }));
    expect(redirected).toEqual({});
  });

  test("reads a known profile URL of the site before guessing", async () => {
    const result = await identifyWith(
      { name: "Annabel Redd", urls: ["https://fapgoat.com/models/annabel-redd/"] },
      [],
      await savedPages(),
    );
    expect(result.performer?.external_id).toBe("kvs-fapgoat-model-annabel-redd");
  });
});

describe("fullestProfile", () => {
  test("prefers the profile with the most facts and breaks ties by order", async () => {
    const bare = parseProfile(site("neporn"), "https://neporn.com/models/katie-kinz/", await fixture("profile-bare"));
    const rich = parseProfile(site("fapgoat"), "https://fapgoat.com/models/annabel-redd/", await fixture("profile-fapgoat"));
    const other = parseProfile(site("neporn"), "https://neporn.com/models/katie-kinz/", await fixture("profile-bare"));
    expect(fullestProfile([bare, undefined, rich])).toBe(rich);
    expect(fullestProfile([bare, other])).toBe(bare);
    expect(fullestProfile([undefined])).toBeUndefined();
  });
});
