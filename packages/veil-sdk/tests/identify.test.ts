import { describe, expect, test } from "bun:test";
import { foldName, isSamePerformer } from "../src/identify.ts";

describe("isSamePerformer", () => {
  test("matches a name across case, accents and punctuation", () => {
    expect(foldName("Mía  Malkova!")).toBe("mia malkova");
    expect(isSamePerformer({ name: "MIA MALKOVA" }, { name: "Mía Malkova" })).toBe(true);
    expect(foldName("Ma’Lynn – O’Neil")).toBe("ma lynn o neil");
  });

  test("keeps names in other scripts", () => {
    expect(foldName("三上 悠亜・")).toBe("三上 悠亜");
    expect(isSamePerformer({ name: "三上悠亜" }, { name: "三上 悠亜" })).toBe(false);
  });

  test("matches through an alias on either side", () => {
    expect(isSamePerformer({ name: "Ada L", aliases: ["Ada Lane"] }, { name: "Ada Lane" })).toBe(true);
    expect(isSamePerformer({ name: "Ada L" }, { name: "Ada Lane", aliases: ["Ada L."] })).toBe(true);
  });

  test("refuses another name or a different birthdate", () => {
    expect(isSamePerformer({ name: "Ada Lanes" }, { name: "Ada Lane" })).toBe(false);
    expect(
      isSamePerformer({ name: "Ada Lane", birthdate: "1990-01-02" }, { name: "Ada Lane", birthdate: "1991-01-02" }),
    ).toBe(false);
    expect(isSamePerformer({ name: "Ada Lane", country: "USA" }, { name: "Ada Lane", country: "United States" })).toBe(
      true,
    );
  });
});
