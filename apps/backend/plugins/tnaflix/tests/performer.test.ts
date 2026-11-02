import { test, expect, describe } from "bun:test";
import { isProfileUrl, parseProfilePage, profileSlugFromUrl, slugFromName } from "../src/performer.ts";
import { candidates, identifyPerformer, type PageFetcher } from "../src/identify.ts";

const fixture = (name: string) => Bun.file(`${import.meta.dir}/fixtures/${name}`).text();
const mia = await fixture("profile-mia-malkova.html");
const danny = await fixture("profile-danny-d.html");
const sophia = await fixture("profile-sophia-laure.html");
const member = await fixture("profile-member.html");

/** A fetcher serving fixtures by slug and recording the URLs asked for; unknown slugs are 404s. */
function fakeFetcher(pages: Record<string, string>, asked: string[] = []): PageFetcher {
  return async (url) => {
    asked.push(url);
    return pages[profileSlugFromUrl(url)];
  };
}

describe("profile URLs", () => {
  test("only /profile/<slug> pages count", () => {
    expect(isProfileUrl("https://www.tnaflix.com/profile/mia-malkova")).toBe(true);
    expect(isProfileUrl("https://www.tnaflix.com/channel/brazzers")).toBe(false);
    expect(isProfileUrl("https://www.tnaflix.com/profile")).toBe(false);
  });

  test("slugFromName folds accents and punctuation", () => {
    expect(slugFromName("Mia Malkova")).toBe("mia-malkova");
    expect(slugFromName("Cléo D'Arc")).toBe("cleo-d-arc");
  });
});

describe("parseProfilePage", () => {
  test("reads name, canonical URL, stable id and the performer photo", () => {
    const performer = parseProfilePage(mia, "https://www.tnaflix.com/profile/mia-malkova");
    expect(performer).toMatchObject({
      type: "performer",
      external_id: "tnaflix-profile-mia-malkova",
      source_url: "https://www.tnaflix.com/profile/mia-malkova",
      name: "Mia Malkova",
    });
    expect(performer?.image_path).toContain("/pics/alpha/pornstars/");
    expect(performer?.images).toEqual([{ type: "profile", file_path: performer?.image_path as string }]);
  });

  test("a placeholder avatar gives no picture", () => {
    expect(parseProfilePage(danny, "https://www.tnaflix.com/profile/danny-d")?.image_path).toBeUndefined();
  });

  test("a page without a profile header is no performer", () => {
    expect(parseProfilePage("<html><body>404</body></html>", "https://www.tnaflix.com/profile/x")).toBeUndefined();
  });
});

describe("identifyPerformer", () => {
  test("candidates: known site URLs first, then name and alias slugs, no duplicates", () => {
    expect(
      candidates({
        name: "Sophia Laure",
        aliases: ["Sophia Laure", "Sofia L"],
        urls: ["https://www.tnaflix.com/profile/sophia-laure-130239", "https://www.eporner.com/pornstar/x/"],
      })
    ).toEqual([
      { slug: "sophia-laure-130239", known: true },
      { slug: "sophia-laure", known: false },
      { slug: "sofia-l", known: false },
    ]);
  });

  test("returns the profile when the name slug matches and the photo is a performer's", async () => {
    const asked: string[] = [];
    const result = await identifyPerformer({ name: "Mia Malkova" }, fakeFetcher({ "mia-malkova": mia }, asked));
    expect(result.performer?.name).toBe("Mia Malkova");
    expect(asked).toEqual(["https://www.tnaflix.com/profile/mia-malkova"]);
  });

  test("a profile of another name under the slug is rejected", async () => {
    expect(await identifyPerformer({ name: "Riley Reid" }, fakeFetcher({ "riley-reid": mia }))).toEqual({});
  });

  test("a member account without a performer photo is not guessed to be the performer", async () => {
    expect(await identifyPerformer({ name: "kelsie62" }, fakeFetcher({ kelsie62: member }))).toEqual({});
  });

  test("a profile URL the library already knows is accepted without a photo, once the name agrees", async () => {
    const pages = { "danny-d": danny };
    const known = await identifyPerformer(
      { name: "Danny D", urls: ["https://www.tnaflix.com/profile/danny-d"] },
      fakeFetcher(pages)
    );
    expect(known.performer?.external_id).toBe("tnaflix-profile-danny-d");
    expect(await identifyPerformer({ name: "Danny D" }, fakeFetcher(pages))).toEqual({});
  });

  test("a slug with an id suffix is read off a known URL", async () => {
    const result = await identifyPerformer(
      { name: "Sophia Laure", urls: ["https://www.tnaflix.com/profile/sophia-laure-130239"] },
      fakeFetcher({ "sophia-laure-130239": sophia })
    );
    expect(result.performer?.source_url).toBe("https://www.tnaflix.com/profile/sophia-laure-130239");
  });

  test("no profile at any slug gives an empty result", async () => {
    expect(await identifyPerformer({ name: "Nobody Atall" }, fakeFetcher({}))).toEqual({});
  });
});
