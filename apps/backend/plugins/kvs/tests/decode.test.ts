import { test, expect, describe } from "bun:test";
import { decodeVideoUrl, isObfuscated, licenseToken } from "../src/decode.ts";

// Expected values come from the reference implementation of kt_player's
// hash swap (yt-dlp's KVS extractor) and, for the last case, from a live
// yesporn.vip page whose decoded link the server accepted (the raw one 404s).

describe("decodeVideoUrl", () => {
  test("leaves plain URLs untouched", () => {
    const url = "https://www.analdin.com/get_file/12/4daa43101e61f4e76c652ac1a08cb599/803000/803178/803178.mp4/";
    expect(isObfuscated(url)).toBe(false);
    expect(decodeVideoUrl(url, "$463045615276685")).toBe(url);
  });

  test("unshuffles the hash segment and keeps the query", () => {
    const obfuscated =
      "function/0/https://www.example.com/get_file/1/0123456789abcdef0123456789abcdef/1000/1234/1234.mp4/?br=100";
    expect(isObfuscated(obfuscated)).toBe(true);
    expect(decodeVideoUrl(obfuscated, "$463045615276685")).toBe(
      "https://www.example.com/get_file/1/028433675c59d2b9baacf110f876dee4/1000/1234/1234.mp4/?br=100"
    );
  });

  test("matches the reference for another license", () => {
    const obfuscated =
      "function/0/https://www.example.com/get_file/12/4daa43101e61f4e76c652ac1a08cb599/803000/803178/803178.mp4/";
    expect(decodeVideoUrl(obfuscated, "$481703115892229")).toBe(
      "https://www.example.com/get_file/12/e9aa804551dc27c4166e19aaf1cb0346/803000/803178/803178.mp4/"
    );
  });

  test("decodes a real yesporn.vip link, keeping hash characters past the first 32", () => {
    const obfuscated =
      "function/0/https://yesporn.vip/get_file/5/7d6fd114c9e7289c81e78749465d532f735912fc1a/75000/75378/75378_trim.mp4/";
    expect(decodeVideoUrl(obfuscated, "$494000116297929")).toBe(
      "https://yesporn.vip/get_file/5/4d84d12e7ff61595939c878cd4e72617735912fc1a/75000/75378/75378_trim.mp4/"
    );
  });

  test("rejects a license code too short to derive a token", () => {
    const obfuscated = "function/0/https://x.com/get_file/1/0123456789abcdef0123456789abcdef/1/1/1.mp4/";
    expect(() => decodeVideoUrl(obfuscated, "$123")).toThrow();
  });
});

describe("licenseToken", () => {
  test("yields four digits per key digit", () => {
    const token = licenseToken("$463045615276685");
    expect(token.length).toBe(32);
    expect(token.every((digit) => digit >= 0 && digit <= 9)).toBe(true);
  });
});
