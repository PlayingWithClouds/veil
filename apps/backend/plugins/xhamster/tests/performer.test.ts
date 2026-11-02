import { test, expect, describe } from "bun:test";
import {
  isPornstarUrl,
  parsePornstarPage,
  pornstarSlugFromUrl,
  slugFromName,
} from "../src/performer.ts";
import { candidateSlugs, identifyPerformer, type PageFetcher } from "../src/identify.ts";

const fixture = (name: string) => Bun.file(`${import.meta.dir}/fixtures/${name}`).text();
const angela = await fixture("pornstar-angela-white.html");
const riley = await fixture("pornstar-riley-reid.html");
const mia = await fixture("pornstar-mia-malkova.html");

/** A fetcher serving fixtures by slug and recording the URLs asked for; unknown slugs are 404s. */
function fakeFetcher(pages: Record<string, string>, asked: string[] = []): PageFetcher {
  return async (url) => {
    asked.push(url);
    const slug = pornstarSlugFromUrl(url);
    return pages[slug];
  };
}

describe("pornstar URLs", () => {
  test("only profile pages count, not the listings", () => {
    expect(isPornstarUrl("https://xhamster.com/pornstars/mia-malkova")).toBe(true);
    expect(isPornstarUrl("https://xhamster.com/pornstars/mia-malkova/")).toBe(true);
    expect(isPornstarUrl("https://xhamster.com/pornstars/all/countries/us")).toBe(false);
    expect(isPornstarUrl("https://xhamster.com/pornstars")).toBe(false);
    expect(isPornstarUrl("https://xhamster.com/videos/foo-xh1")).toBe(false);
  });

  test("slugFromName folds accents and punctuation", () => {
    expect(slugFromName("Mia Malkova")).toBe("mia-malkova");
    expect(slugFromName("Cléo  D'Arc!")).toBe("cleo-d-arc");
  });
});

describe("parsePornstarPage", () => {
  test("fills the fields the page shows", () => {
    const performer = parsePornstarPage(angela, "https://xhamster.com/pornstars/angela-white");
    expect(performer).toMatchObject({
      type: "performer",
      external_id: "xhamster-ps-angela-white",
      source_url: "https://xhamster.com/pornstars/angela-white",
      name: "Angela White",
      gender: "female",
      country: "Australia",
      ethnicity: "White",
      hair_color: "Brown",
      eye_color: "Blue",
      height_cm: 160,
    });
    expect(performer?.details).toContain("Angela White");
    expect(performer?.image_path).toContain("xhcdn.com");
    expect(performer?.images).toEqual([{ type: "profile", file_path: performer?.image_path as string }]);
  });

  test("aliases come from the comma line, without the own name", () => {
    const performer = parsePornstarPage(riley, "https://xhamster.com/pornstars/riley-reid");
    expect(performer?.aliases).toEqual(["Molly", "Riley Ried", "Riley Reed", "Paige Riley"]);
  });

  test("a sparse profile still has name, gender, country and picture", () => {
    const performer = parsePornstarPage(mia, "https://xhamster.com/pornstars/mia-malkova");
    expect(performer?.name).toBe("Mia Malkova");
    expect(performer?.gender).toBe("female");
    expect(performer?.country).toBe("United States");
    expect(performer?.image_path).toBeTruthy();
    expect(performer?.height_cm).toBeUndefined();
  });

  test("a page without profile state is no performer", () => {
    expect(parsePornstarPage("<html><body>404</body></html>", "https://xhamster.com/pornstars/x")).toBeUndefined();
  });
});

describe("identifyPerformer", () => {
  test("candidate slugs: known site URL first, then name, then aliases, no duplicates", () => {
    expect(
      candidateSlugs({
        name: "Mia Malkova",
        aliases: ["Madison Swan", "Mia Malkova"],
        urls: ["https://www.eporner.com/pornstar/mia-malkova/", "https://xhamster.com/pornstars/mia-m"],
      })
    ).toEqual(["mia-m", "mia-malkova", "madison-swan"]);
  });

  test("returns the profile when the name matches", async () => {
    const asked: string[] = [];
    const result = await identifyPerformer({ name: "Angela White" }, fakeFetcher({ "angela-white": angela }, asked));
    expect(result.performer?.name).toBe("Angela White");
    expect(asked).toEqual(["https://xhamster.com/pornstars/angela-white"]);
  });

  test("an alias of the profile matches too", async () => {
    const result = await identifyPerformer({ name: "Riley Reed" }, fakeFetcher({ "riley-reed": riley }));
    expect(result.performer?.external_id).toBe("xhamster-ps-riley-reid");
  });

  test("a profile of another person under the same slug is rejected", async () => {
    const result = await identifyPerformer({ name: "Angela White" }, fakeFetcher({ "angela-white": mia }));
    expect(result).toEqual({});
  });

  test("a known birthdate is not held against a profile that states none", async () => {
    const result = await identifyPerformer(
      { name: "Angela White", birthdate: "1985-03-18" },
      fakeFetcher({ "angela-white": angela })
    );
    expect(result.performer?.name).toBe("Angela White");
  });

  test("no profile at any slug gives an empty result", async () => {
    expect(await identifyPerformer({ name: "Nobody Atall" }, fakeFetcher({}))).toEqual({});
  });
});
