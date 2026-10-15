import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { base64ToBytes } from "@playingwithclouds/veil-sdk";
import { identifyFace } from "../src/identify.ts";
import { parseCandidates } from "../src/parse.ts";

/** Loads a saved answer or picture from the fixtures folder. */
function loadFixture(fileName: string): Buffer {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url));
}

describe("parseCandidates", () => {
  const candidates = parseCandidates(loadFixture("candidates.html").toString("utf8"));

  test("reads every candidate block, best first", () => {
    expect(candidates.length).toBeGreaterThan(10);
    for (let index = 1; index < candidates.length; index++) {
      expect(candidates[index - 1].similarity).toBeGreaterThanOrEqual(candidates[index].similarity);
    }
  });

  test("reads name, similarity, image and profile links", () => {
    const top = candidates[0];
    expect(top.name).toBe("Lexi Lore");
    expect(top.similarity).toBe(1);
    expect(top.image_url).toBe("https://pornstarbyface.com/ImgFiles/Lexi%20Lore/1.jpg");
    expect(top.profile_urls).toEqual(["https://vrporn.com/pornstars/lexi-lore/"]);
  });

  test("an error answer throws its message", () => {
    expect(() => parseCandidates('{"error":true,"errorMsg":"No face found"}')).toThrow("No face found");
  });
});

describe("base64ToBytes", () => {
  test("matches Buffer for all padding lengths", () => {
    for (const text of ["", "a", "ab", "abc", "abcd", "ÿ\u0000\u0001"]) {
      const encoded = Buffer.from(text, "latin1").toString("base64");
      expect(Array.from(base64ToBytes(encoded))).toEqual(Array.from(Buffer.from(text, "latin1")));
    }
  });
});

describe("live", () => {
  test.skipIf(!process.env.VEIL_LIVE)(
    "identifies a performer's own photo",
    async () => {
      const result = await identifyFace({
        image_base64: loadFixture("face.jpg").toString("base64"),
        mime_type: "image/jpeg",
      });
      expect(result.candidates[0].name).toBe("Lexi Lore");
    },
    60_000,
  );
});
