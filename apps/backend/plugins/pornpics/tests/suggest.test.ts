import { test, expect, describe } from "bun:test";
import { suggestionsFromResponse } from "../src/suggest.ts";

describe("suggestionsFromResponse", () => {
  test("a link to a pornstar page makes a performer, anything else a plain search", () => {
    const suggestions = suggestionsFromResponse([
      { query: "isla white", link: "/pornstars/isla-white/" },
      { query: "island", link: "" },
      { query: "isla" },
    ]);
    expect(suggestions).toEqual([
      { text: "isla white", kind: "performer" },
      { text: "island", kind: "query" },
      { text: "isla", kind: "query" },
    ]);
  });

  test("skips blanks and repeats, ignoring case", () => {
    const suggestions = suggestionsFromResponse([{ query: "Isla" }, { query: "isla" }, { query: " " }, {}]);
    expect(suggestions).toEqual([{ text: "Isla", kind: "query" }]);
  });
});
