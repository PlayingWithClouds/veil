import { test, expect, describe } from "bun:test";
import { suggestionsFromResponse } from "../src/suggest.ts";

describe("suggestionsFromResponse", () => {
  test("link types become suggestion kinds", () => {
    const suggestions = suggestionsFromResponse([
      { query: "cassie del isla", link_type: "search" },
      { query: "isla moon", link_type: "pornstar" },
      { query: "isla studio", link_type: "channel" },
      { query: "island" },
    ]);
    expect(suggestions).toEqual([
      { text: "cassie del isla", kind: "query" },
      { text: "isla moon", kind: "performer" },
      { text: "isla studio", kind: "studio" },
      { text: "island", kind: "query" },
    ]);
  });

  test("skips blanks and repeats, ignoring case", () => {
    const suggestions = suggestionsFromResponse([{ query: "Isla" }, { query: "isla" }, { query: "" }, {}]);
    expect(suggestions).toEqual([{ text: "Isla", kind: "query" }]);
  });
});
