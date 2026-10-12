import { test, expect, describe } from "bun:test";
import { suggestionsFromResponse } from "../src/suggest.ts";

describe("suggestionsFromResponse", () => {
  test("lists searches first, then performers, categories and channels", () => {
    const suggestions = suggestionsFromResponse({
      keywords: ["Isla Summers", "Island Long"],
      models: [{ title: "Isla Biza" }],
      categories: [{ title: "Anal" }],
      channels: [{ title: "Isla Studio" }],
    });
    expect(suggestions).toEqual([
      { text: "Isla Summers", kind: "query" },
      { text: "Island Long", kind: "query" },
      { text: "Isla Biza", kind: "performer" },
      { text: "Anal", kind: "tag" },
      { text: "Isla Studio", kind: "studio" },
    ]);
  });

  test("skips blanks and repeats across groups, ignoring case", () => {
    const suggestions = suggestionsFromResponse({
      keywords: ["Isla Biza", "  ", "isla biza"],
      models: [{ title: "ISLA BIZA" }, {}],
    });
    expect(suggestions).toEqual([{ text: "Isla Biza", kind: "query" }]);
  });

  test("an empty answer has no suggestions", () => {
    expect(suggestionsFromResponse({})).toEqual([]);
  });
});
