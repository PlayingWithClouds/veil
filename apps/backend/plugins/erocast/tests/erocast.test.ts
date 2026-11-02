import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { audioFromTrackPage, genreNamesOf, scrapeAudio } from "../src/audio.ts";
import { decodeEntities, toSiteUrl, trackIdFromUrl } from "../src/http.ts";
import { audioList, listingPath, tracksFromSearch } from "../src/listing.ts";
import { suggest, suggestionsFromReply } from "../src/suggest.ts";
import { cardOf, tracksFromPage } from "../src/track.ts";

const NET_TIMEOUT = 60_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("track ids come from track URLs", () => {
    expect(trackIdFromUrl("https://www.erocast.me/track/15000/m4a-breaking-a-sweat-in-the-garden")).toBe("15000");
    expect(trackIdFromUrl("https://www.erocast.me/julietmonroeasmr")).toBe("");
  });

  test("the bare domain is pointed at www", () => {
    expect(toSiteUrl("https://erocast.me/track/1/x")).toBe("https://www.erocast.me/track/1/x");
    expect(toSiteUrl("https://www.erocast.me/track/1/x")).toBe("https://www.erocast.me/track/1/x");
  });

  test("entities decode in one pass", () => {
    expect(decodeEntities("a &amp;lt; b &amp; c &#39;d&#39;")).toBe("a &lt; b & c 'd'");
  });

  test("listing paths: newest by default, creators under /tracks, listing pages as they are", () => {
    expect(listingPath(undefined)).toBe("/channel/new-audios");
    expect(listingPath("https://www.erocast.me/julietmonroeasmr")).toBe("/julietmonroeasmr/tracks");
    expect(listingPath("https://erocast.me/julietmonroeasmr/tracks?page=3")).toBe("/julietmonroeasmr/tracks");
    expect(listingPath("https://www.erocast.me/tracks/tag/anal")).toBe("/tracks/tag/anal");
    expect(listingPath("https://www.erocast.me/genre/narrative")).toBe("/genre/narrative");
    expect(() => listingPath("https://www.erocast.me/a/b")).toThrow();
  });
});

describe("listing cards", () => {
  const tracks = tracksFromPage(loadFixture("channel-page.html"));

  test("every embedded track is read", () => {
    expect(tracks.map((track) => track.id)).toEqual([18824, 18823, 18822, 17746, 17745]);
  });

  test("a converted track's card carries its HLS playlist", () => {
    const card = cardOf(tracks[3]);
    expect(card.media_type).toBe("audio");
    expect(card.external_id).toBe("erocast-17746");
    expect(card.source_url).toStartWith("https://www.erocast.me/track/17746/");
    expect(card.duration_seconds).toBeGreaterThan(0);
    expect(card.downloads).toHaveLength(1);
    expect(card.downloads?.[0].format).toBe("hls");
    expect(card.downloads?.[0].url).toEndWith("/track.m3u8");
  });

  test("a track still being converted has no download", () => {
    const card = cardOf(tracks[0]);
    expect(card.title).toBe("[F4M] Bred By Him Part 3 The Window");
    expect(card.date).toBe("2026-11-02");
    expect(card.downloads).toBeUndefined();
  });

  test("the grey placeholder is no poster", () => {
    expect(cardOf(tracks[0]).poster_path).toBeUndefined();
  });

  test("search replies list their tracks", () => {
    const found = tracksFromSearch(loadFixture("search.json"));
    expect(found).toHaveLength(3);
    expect(cardOf(found[1]).title).toStartWith("Teasing you while we smoke");
  });
});

describe("audio page", () => {
  test("a converted track: title, voice, tags, cover, duration, HLS download", () => {
    const audio = audioFromTrackPage(loadFixture("track-page.html"), "15000");
    expect(audio.type).toBe("audio");
    expect(audio.external_id).toBe("erocast-15000");
    expect(audio.source_url).toBe("https://www.erocast.me/track/15000/m4a-breaking-a-sweat-in-the-garden");
    expect(audio.title).toBe("M4A breaking a sweat in the garden");
    expect(audio.date).toBe("2025-10-25");
    expect(audio.duration_seconds).toBe(618);
    expect(audio.cover_path).toContain("wasabisys.com");
    expect(audio.performers?.[0].external_id).toBe("hansleftasscheek");
    expect(audio.tags).toEqual(["bunny finn", "yuurivoice"]);
    expect(audio.downloads).toEqual([
      { label: "HLS", url: "https://erocast.s3.us-east-2.wasabisys.com/1508720/track.m3u8", format: "hls" },
    ]);
  });

  test("genres come first and the description keeps its line breaks and decoded entities", () => {
    const html = loadFixture("track-pending.html");
    expect(genreNamesOf(html)).toEqual(["Fdom", "Narrative"]);
    const audio = audioFromTrackPage(html, "18824");
    expect(audio.tags).toEqual(["Fdom", "Narrative", "hotwife", "cuckold", "f4m", "breeding"]);
    expect(audio.details).toContain("\n\n10-minute preview");
    expect(audio.details).toContain("utm_source=erocast&utm_medium=social");
    expect(audio.downloads).toEqual([]);
  });

  test("a page without the requested track fails", () => {
    expect(() => audioFromTrackPage(loadFixture("track-page.html"), "1")).toThrow();
  });

  test("a URL that is no track page fails", async () => {
    await expect(scrapeAudio("https://www.erocast.me/julietmonroeasmr")).rejects.toThrow();
  });
});

describe("suggestions", () => {
  test("creators come first as performers, then track titles, each once", () => {
    const suggestions = suggestionsFromReply(loadFixture("suggest.json"));
    expect(suggestions[0].kind).toBe("performer");
    expect(suggestions.some((entry) => entry.kind === "query")).toBe(true);
    const texts = suggestions.map((entry) => entry.text.toLowerCase());
    expect(new Set(texts).size).toBe(texts.length);
  });
});

describe("live site", () => {
  test(
    "the newest uploads page, trimmed by offset and limit",
    async () => {
      const first = await audioList({ limit: 5 });
      expect(first.items).toHaveLength(5);
      const shifted = await audioList({ limit: 3, offset: 2 });
      expect(shifted.items.map((item) => item.external_id)).toEqual(first.items.slice(2, 5).map((item) => item.external_id));
    },
    NET_TIMEOUT,
  );

  test(
    "a search spanning a page boundary",
    async () => {
      const result = await audioList({ query: "teasing", limit: 6, offset: 17 });
      expect(result.items).toHaveLength(6);
      expect(result.items[0].media_type).toBe("audio");
    },
    NET_TIMEOUT,
  );

  test(
    "a converted track scrapes with a playable playlist",
    async () => {
      const [result] = await scrapeAudio("https://erocast.me/track/15000/anything");
      if (result.type !== "audio") throw new Error("expected audio");
      expect(result.audio.title).toBe("M4A breaking a sweat in the garden");
      const playlist = await fetch(result.audio.downloads?.[0].url ?? "");
      expect(await playlist.text()).toStartWith("#EXTM3U");
    },
    NET_TIMEOUT,
  );

  test(
    "suggestions",
    async () => {
      const result = await suggest({ query: "teas", limit: 5 });
      expect(result.suggestions.length).toBeGreaterThan(0);
    },
    NET_TIMEOUT,
  );
});
