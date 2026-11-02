import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { audioFromPage, findAudio, mediaUrlOf } from "../src/audio.ts";
import { audioList, parseUserPage } from "../src/listing.ts";
import { externalIdOf, parseAddress } from "../src/site.ts";
import { splitTitle } from "../src/titles.ts";

const NET_TIMEOUT = 60_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("addresses", () => {
  test("user and audio URLs are told apart", () => {
    expect(parseAddress("https://soundgasm.net/u/sassmastah77")).toEqual({ user: "sassmastah77", slug: undefined });
    expect(parseAddress("https://soundgasm.net/u/sassmastah77/Hmmm?x=1")).toEqual({ user: "sassmastah77", slug: "Hmmm" });
    expect(parseAddress("https://example.com/u/a/b")).toBeUndefined();
    expect(parseAddress("https://soundgasm.net/login")).toBeUndefined();
  });

  test("slugs are only unique per uploader", () => {
    expect(externalIdOf("a", "x")).not.toBe(externalIdOf("b", "x"));
  });
});

describe("titles", () => {
  test("bracketed phrases become lowercase tags and the rest the title", () => {
    expect(splitTitle("[F4M] [ASMR] Door-to-Door [Whispers] [asmr]")).toEqual({
      title: "Door-to-Door",
      tags: ["f4m", "asmr", "whispers"],
    });
  });

  test("a title of only brackets stays whole", () => {
    expect(splitTitle("[F4M] [Whispers]")).toEqual({ title: "[F4M] [Whispers]", tags: ["f4m", "whispers"] });
  });

  test("sentence-long brackets are not tags", () => {
    const split = splitTitle("[F4M] Factory [Meet the mysterious Olly Lolis and the whole crew!]");
    expect(split.title).toBe("Factory");
    expect(split.tags).toEqual(["f4m"]);
  });
});

describe("audio page", () => {
  const html = loadFixture("audio.html");

  test("the player script gives the m4a", () => {
    expect(mediaUrlOf(html)).toBe("https://media.soundgasm.net/sounds/b3a4a86e4a966510036e2c128c5b11916e60e547.m4a");
    expect(mediaUrlOf("<html></html>")).toBe("");
  });

  test("builds the audio with uploader, tags, description and download", () => {
    const audio = audioFromPage(html, "sassmastah77", "f4m-Door-to-Door-Blowjob-Whore-1");
    expect(audio?.title).toBe("Door-to-Door Blowjob Whore");
    expect(audio?.tags).toEqual(["f4m"]);
    expect(audio?.details).toStartWith("[sloppy bj from a slutty legal teen]");
    expect(audio?.external_id).toBe("soundgasm-sassmastah77/f4m-Door-to-Door-Blowjob-Whore-1");
    expect(audio?.source_url).toBe("https://soundgasm.net/u/sassmastah77/f4m-Door-to-Door-Blowjob-Whore-1");
    expect(audio?.performers?.[0]).toEqual({
      name: "sassmastah77",
      external_id: "sassmastah77",
      source_url: "https://soundgasm.net/u/sassmastah77",
    });
    expect(audio?.downloads).toEqual([
      { label: "m4a", url: "https://media.soundgasm.net/sounds/b3a4a86e4a966510036e2c128c5b11916e60e547.m4a", format: "m4a" },
    ]);
  });

  test("a page without a player yields nothing", () => {
    expect(audioFromPage("<html><body>404</body></html>", "a", "b")).toBeUndefined();
  });
});

describe("user page", () => {
  test("lists every card with clean title, id and play count", () => {
    const items = parseUserPage(loadFixture("user.html"));
    expect(items).toHaveLength(8);
    const hmmm = items.find((item) => item.title === "Hmmm….");
    expect(hmmm).toEqual({
      title: "Hmmm….",
      media_type: "audio",
      source_url: "https://soundgasm.net/u/sassmastah77/Hmmm",
      external_id: "soundgasm-sassmastah77/Hmmm",
      view_count: 25980,
    });
    expect(items.find((item) => item.title.includes("Left"))?.title).toBe("Tegan & Sara - I Won't Be Left");
    expect(items.every((item) => !item.title.includes("["))).toBe(true);
  });

  test("an unknown user (empty page) lists nothing", () => {
    expect(parseUserPage(loadFixture("empty-user.html"))).toEqual([]);
  });
});

describe("list arguments", () => {
  test("a query or missing URL lists nothing, the site has no search", async () => {
    expect(await audioList({ query: "asmr" })).toEqual({ items: [] });
    expect(await audioList({})).toEqual({ items: [] });
  });

  test("a URL that is not a Soundgasm user lists nothing", async () => {
    expect(await findAudio("https://soundgasm.net/u/sassmastah77")).toEqual([]);
    expect(await audioList({ url: "https://example.com/" })).toEqual({ items: [] });
  });
});

describe("live site", () => {
  test(
    "finds one audio and lists a user page with paging",
    async () => {
      const found = await findAudio("https://soundgasm.net/u/sassmastah77/f4m-Door-to-Door-Blowjob-Whore-1");
      expect(found[0]?.type).toBe("audio");

      const page = await audioList({ url: "https://soundgasm.net/u/sassmastah77", limit: 3, offset: 1 });
      expect(page.items).toHaveLength(3);
    },
    NET_TIMEOUT,
  );
});
