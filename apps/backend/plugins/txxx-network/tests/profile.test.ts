import { test, expect, describe } from "bun:test";
import { modelApiPath, modelDir, parseBirthDate, parseMeasurements, parseModel, type ApiModel, type SiteLanguage } from "../src/profile.ts";
import { identifyWith, knownGuesses, nameGuesses, type ModelLookup } from "../src/identify.ts";

const rileyResponse = await Bun.file(`${import.meta.dir}/fixtures/model-riley-reid.json`).json();
const language: SiteLanguage = await Bun.file(`${import.meta.dir}/fixtures/lang-en.json`).json();
const riley: ApiModel = rileyResponse.model;

describe("parseModel", () => {
  test("fills the profile from the model record and the site's language file", () => {
    const performer = parseModel("txxx.com", riley, language);
    expect(performer).toMatchObject({
      type: "performer",
      external_id: "txxx-model-riley-reid@txxx.com",
      source_url: "https://txxx.com/models/riley-reid/",
      name: "Riley Reid",
      aliases: ["Paige Riley", "Riley Reed", "Reily Reid", "Page Riley", "Ashley Mathews"],
      birthdate: "1991-07-09",
      country: "USA",
      gender: "female",
      ethnicity: "Caucasian",
      hair_color: "Brown",
      eye_color: "Green",
      weight_kg: 45,
      measurements: "32A-23-34",
      career_length: "2010-2023",
      image_path: "https://tn.txxx.tube/contents/models/41056/r240x400_1.jpg",
    });
    expect(performer.details).toContain("XBIZ Award");
    expect(performer.details).toContain("James Deen & Erik Everhard");
    expect(performer.tattoos).toContain("kanji");
  });

  test("leaves out zero heights, unknown codes and empty text", () => {
    const performer = parseModel("txxx.com", riley, language);
    expect(performer.height_cm).toBeUndefined();
    expect(performer.piercings).toBeUndefined();
    expect(performer.url).toBeUndefined();
    const unknownCodes = parseModel("txxx.com", { ...riley, hair: "0", eye: "0", country: "" }, language);
    expect(unknownCodes.hair_color).toBeUndefined();
    expect(unknownCodes.eye_color).toBeUndefined();
    expect(unknownCodes.country).toBeUndefined();
  });

  test("reads records whose fields are null", () => {
    const sparse = { model_id: "1", title: "Some Model", dir: "some-model", alias: null, description: null, hair: null } as unknown as ApiModel;
    expect(parseModel("txxx.com", sparse, language)).toEqual({
      type: "performer",
      external_id: "txxx-model-some-model@txxx.com",
      source_url: "https://txxx.com/models/some-model/",
      name: "Some Model",
    });
  });

  test("files the website as twitter, instagram or url", () => {
    expect(parseModel("txxx.com", { ...riley, website: "https://twitter.com/rileyreid" }, language).twitter).toBe("https://twitter.com/rileyreid");
    expect(parseModel("txxx.com", { ...riley, website: "https://www.instagram.com/x/" }, language).instagram).toBe("https://www.instagram.com/x/");
    expect(parseModel("txxx.com", { ...riley, website: "https://example.com/" }, language).url).toBe("https://example.com/");
  });
});

describe("value parsers", () => {
  test("modelDir reads the slug of a model page on any network site", () => {
    expect(modelDir("https://txxx.com/models/riley-reid/")).toBe("riley-reid");
    expect(modelDir("https://hclips.com/models/riley-reid")).toBe("riley-reid");
    expect(modelDir("https://txxx.com/videos/123/some-video/")).toBe("");
  });

  test("modelApiPath buckets by the first two letters", () => {
    expect(modelApiPath("riley-reid")).toBe("/api/json/model/86400/ri/riley-reid.json");
  });

  test("parseBirthDate reads the site's day-month-year text", () => {
    expect(parseBirthDate("9 July, 1991")).toBe("1991-07-09");
    expect(parseBirthDate("31 December, 1978")).toBe("1978-12-31");
    expect(parseBirthDate("")).toBeUndefined();
  });

  test("parseMeasurements converts centimetre waist and hips to inches", () => {
    expect(parseMeasurements("32A-58 cm-86 cm")).toBe("32A-23-34");
    expect(parseMeasurements("34D-24-36")).toBe("34D-24-36");
    expect(parseMeasurements("")).toBeUndefined();
  });
});

describe("identify", () => {
  /** Serves Riley's record for her slug on txxx.com only. */
  const lookup: ModelLookup = async (host, dir) => {
    if (host === "txxx.com" && dir === "riley-reid") return riley;
    return undefined;
  };

  test("builds slugs from the name and aliases, and reads known model pages", () => {
    expect(nameGuesses({ name: "Riley Reid", aliases: ["Riley Reed", "Riley Reid"] }, "txxx.com")).toEqual([
      { host: "txxx.com", dir: "riley-reid" },
      { host: "txxx.com", dir: "riley-reed" },
    ]);
    expect(
      knownGuesses({
        name: "Riley Reid",
        urls: ["https://hclips.com/models/riley-reid/", "https://www.eporner.com/pornstar/riley-reid/"],
      }),
    ).toEqual([{ host: "hclips.com", dir: "riley-reid" }]);
  });

  test("finds the model on the network when the name matches", async () => {
    const found = await identifyWith({ name: "Riley Reid", birthdate: "1991-07-09" }, lookup);
    expect(found?.host).toBe("txxx.com");
    expect(found?.model.dir).toBe("riley-reid");
  });

  test("matches through an alias", async () => {
    const found = await identifyWith({ name: "Paige Riley", aliases: ["Riley Reid"] }, lookup);
    expect(found?.model.dir).toBe("riley-reid");
  });

  test("returns nothing for a model the network does not list", async () => {
    expect(await identifyWith({ name: "Nobody Known" }, lookup)).toBeUndefined();
  });

  test("returns nothing when the record is another person with the same slug", async () => {
    const otherPerson: ModelLookup = async () => ({ ...riley, title: "Riley Reid Jr" });
    expect(await identifyWith({ name: "Riley Reid" }, otherPerson)).toBeUndefined();
  });

  test("returns nothing when birthdates or genders differ", async () => {
    expect(await identifyWith({ name: "Riley Reid", birthdate: "1985-01-01" }, lookup)).toBeUndefined();
    expect(await identifyWith({ name: "Riley Reid", gender: "male" }, lookup)).toBeUndefined();
  });
});
