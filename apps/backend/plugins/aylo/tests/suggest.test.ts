import { test, expect, describe } from "bun:test";
import { suggestionsFromResponse } from "../src/suggest.ts";

describe("suggestionsFromResponse", () => {
  test("RedTube entries become plain searches", () => {
    const suggestions = suggestionsFromResponse([
      { label: "Cassie Del Isla" },
      { label: "Isla Moon" },
    ]);
    expect(suggestions).toEqual([
      { text: "Cassie Del Isla", kind: "query" },
      { text: "Isla Moon", kind: "query" },
    ]);
  });

  test("skips blanks and repeats, ignoring case", () => {
    const suggestions = suggestionsFromResponse([{ label: "Isla" }, { label: "ISLA" }, { label: " " }, {}]);
    expect(suggestions).toEqual([{ text: "Isla", kind: "query" }]);
  });
});
