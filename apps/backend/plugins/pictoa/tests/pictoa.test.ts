import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { albumIdFromUrl, albumsFromRoot } from "../src/cards.ts";
import { searchPathFromQuery, galleryList } from "../src/listing.ts";
import { albumPageUrl, canonicalAlbumUrl, imageUrlsOnPage, pageCount, scrapeGallery, titleOf } from "../src/album.ts";
import { resolveUrl } from "../src/http.ts";

const NET_TIMEOUT = 30_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("albumIdFromUrl reads the trailing id of an album URL", () => {
    expect(albumIdFromUrl("https://www.pictoa.com/albums/alexia-latina-teen-3710577.html")).toBe("3710577");
    expect(albumIdFromUrl("https://www.pictoa.com/albums/alexia-latina-teen-3710577-p2.html")).toBeUndefined();
    expect(albumIdFromUrl("https://www.pictoa.com/albums/alexia-latina-teen-3710577/83678506.html")).toBeUndefined();
  });

  test("searchPathFromQuery turns the query into a search path", () => {
    expect(searchPathFromQuery("Wood Sex")).toBe("/s/wood-sex/");
    expect(searchPathFromQuery("  ")).toBe("");
    expect(searchPathFromQuery(undefined)).toBe("");
  });

  test("album page URLs and canonical URLs", () => {
    const album = "https://www.pictoa.com/albums/alexia-latina-teen-3710577.html";
    expect(albumPageUrl(album, 2)).toBe("https://www.pictoa.com/albums/alexia-latina-teen-3710577-p2.html");
    expect(canonicalAlbumUrl("https://www.pictoa.com/albums/alexia-latina-teen-3710577-p2.html")).toBe(album);
    expect(resolveUrl("/s/wood-sex/")).toBe("https://www.pictoa.com/s/wood-sex/");
  });
});

describe("listing pages (offline)", () => {
  test("the home page lists 72 albums with their covers", () => {
    const items = albumsFromRoot(parse(loadFixture("home.html")));
    expect(items.length).toBe(72);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(72);
    const first = items[0];
    expect(first.media_type).toBe("gallery");
    expect(first.title).toBe("Tits, boobs, breasts slapping GIFs");
    expect(first.external_id).toBe("pictoa-3710271");
    expect(first.source_url).toBe("https://www.pictoa.com/albums/tits-boobs-breasts-slapping-gifs-3710271.html");
    expect(first.poster_path).toBe(
      "https://t1.pictoa.com/media/galleries/293/763/2937635fdb75d43b44f/37102715fdb75d77181b.gif"
    );
  });

  test("album cards carry their photo count", () => {
    const items = albumsFromRoot(parse(loadFixture("search.html")));
    const zoe = items.find((item) => item.title.startsWith("Amatuer brunette Zoe Wood"));
    expect(zoe?.count).toBe(15);
  });

  test("a search page lists its albums too", () => {
    const items = albumsFromRoot(parse(loadFixture("search.html")));
    expect(items.length).toBe(72);
    expect(items[0].external_id).toBe("pictoa-3623103");
  });
});

describe("album page (offline)", () => {
  const firstPage = parse(loadFixture("album-page1.html"));
  const secondPage = parse(loadFixture("album-page2.html"));

  test("the heading is the album title", () => {
    expect(titleOf(firstPage, "3710577")).toBe("Alexia latina teen");
    expect(titleOf(parse("<html></html>"), "42")).toBe("Album 42");
  });

  test("pagination shows two pages", () => {
    expect(pageCount(firstPage)).toBe(2);
  });

  test("the first page yields full-size image URLs for its image cards", () => {
    const urls = imageUrlsOnPage(firstPage);
    expect(urls.length).toBe(84);
    expect(urls[0]).toBe("https://s2.pictoa.com/media/galleries/256/483/2564835fdb772a50f62/37105775fdb772cbbd87.jpg");
    for (const url of urls) expect(url).toStartWith("https://s2.pictoa.com/media/galleries/");
  });

  test("the second page continues the album", () => {
    expect(imageUrlsOnPage(secondPage).length).toBe(64);
  });
});

// ---------------------------------------------------------------------------
// Live integration against pictoa.com.

describe("gallery:list (live)", () => {
  test(
    "returns a bounded set of album stubs",
    async () => {
      const { items } = await galleryList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) {
        expect(item.media_type).toBe("gallery");
        expect(item.external_id).toStartWith("pictoa-");
        expect(item.source_url).toStartWith("https://www.pictoa.com/albums/");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "offset pages do not repeat the first page",
    async () => {
      const first = await galleryList({ limit: 5 });
      const second = await galleryList({ limit: 5, offset: 80 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(second.items.length).toBe(5);
      expect(second.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test(
    "query returns the search's albums",
    async () => {
      const { items } = await galleryList({ query: "wood sex", limit: 10 });
      expect(items.length).toBe(10);
      for (const item of items) expect(item.external_id).toStartWith("pictoa-");
    },
    2 * NET_TIMEOUT
  );
});

describe("gallery:find (live)", () => {
  test(
    "scrapes a listed album with all of its pages",
    async () => {
      const { items } = await galleryList({ limit: 1 });
      const [result] = await scrapeGallery(items[0].source_url);
      expect(result.type).toBe("gallery");
      if (result.type !== "gallery") throw new Error("expected a gallery");
      expect(result.gallery.external_id).toBe(items[0].external_id);
      expect(result.gallery.title.trim()).not.toBe("");
      expect(result.gallery.images?.length).toBeGreaterThan(0);
    },
    2 * NET_TIMEOUT
  );
});
