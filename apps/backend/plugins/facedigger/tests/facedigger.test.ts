import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { agreementAccepted, identifyFace } from "../src/identify.ts";
import { parseCamMatches, similarityFrom } from "../src/parse.ts";

/** Loads a saved answer or picture from the fixtures folder. */
function loadFixture(fileName: string): Buffer {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url));
}

describe("parseCamMatches", () => {
  const result = JSON.parse(loadFixture("result.json").toString("utf8"));
  const candidates = parseCamMatches(result);

  test("reads only the cams category, best first", () => {
    expect(candidates).toHaveLength(result.results.cams.results.length);
    expect(candidates[0].name).toBe("Maya Rouse");
    expect(candidates[0].similarity).toBeCloseTo(0.67);
    for (let index = 1; index < candidates.length; index++) {
      expect(candidates[index - 1].similarity).toBeGreaterThanOrEqual(candidates[index].similarity);
    }
  });

  test("never exposes video titles or OnlyFans accounts", () => {
    const names = candidates.map((candidate) => candidate.name);
    for (const video of result.results.porn.results) expect(names).not.toContain(video.title);
    for (const account of result.results.onlyfans.results) expect(names).not.toContain(account.title);
    for (const candidate of candidates) {
      expect(Object.keys(candidate).sort()).toEqual(["image_url", "name", "profile_urls", "similarity"]);
      for (const link of candidate.profile_urls) expect(link).not.toContain("onlyfans.com");
    }
  });

  test("handles a missing cams category", () => {
    expect(parseCamMatches({ results: { porn: { results: [{ title: "x" }] } } })).toEqual([]);
  });

  test("similarityFrom reads percentages", () => {
    expect(similarityFrom("66.5%")).toBeCloseTo(0.665);
    expect(similarityFrom("junk")).toBe(0);
  });
});

describe("agreement", () => {
  test("identifyFace refuses until the agreement setting is on", async () => {
    delete process.env.FACEDIGGER_ACCEPT_AGREEMENT;
    expect(agreementAccepted()).toBe(false);
    await expect(identifyFace({ image_base64: "AAAA", mime_type: "image/jpeg" })).rejects.toThrow(
      "user agreement has not been accepted",
    );
  });
});

describe("live", () => {
  test.skipIf(!process.env.VEIL_LIVE)(
    "identifies a face among cam models",
    async () => {
      process.env.FACEDIGGER_ACCEPT_AGREEMENT = "true";
      const result = await identifyFace({
        image_base64: loadFixture("face.jpg").toString("base64"),
        mime_type: "image/jpeg",
      });
      expect(result.candidates.length).toBeGreaterThan(0);
    },
    120_000,
  );
});
