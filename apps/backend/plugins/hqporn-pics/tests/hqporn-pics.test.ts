import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parse } from "node-html-parser";
import { galleriesFromRoot } from "../src/cards.ts";
import { gallerySlugFromUrl, decodeEntities, resolveUrl } from "../src/http.ts";
import { galleryFromPage, scrapeGallery } from "../src/gallery.ts";
import { categoriesFor, galleryList } from "../src/listing.ts";

const NET_TIMEOUT = 30_000;

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

describe("helpers", () => {
  test("gallerySlugFromUrl reads the slug of a gallery URL only", () => {
    expect(gallerySlugFromUrl("https://hqporn.pics/galls/hardx146")).toBe("hardx146");
    expect(gallerySlugFromUrl("/galls/GFrevenge_117")).toBe("GFrevenge_117");
    expect(gallerySlugFromUrl("https://hqporn.pics/hardcore/")).toBe("");
  });

  test("resolveUrl and decodeEntities", () => {
    expect(resolveUrl("/model/Mia_Malkova")).toBe("https://hqporn.pics/model/Mia_Malkova");
    expect(decodeEntities("Tom &amp; Jerry")).toBe("Tom & Jerry");
  });
});

describe("category page (offline)", () => {
  const items = galleriesFromRoot(parse(loadFixture("category.html")));

  test("galleriesFromRoot yields each gallery once, as a gallery stub", () => {
    expect(items.length).toBeGreaterThan(100);
    expect(new Set(items.map((item) => item.external_id)).size).toBe(items.length);
    for (const item of items) {
      expect(item.media_type).toBe("gallery");
      expect(item.external_id).toStartWith("hqporn-");
      expect(item.source_url).toStartWith("https://hqporn.pics/galls/");
      expect(item.title).not.toBe("");
    }
  });

  test("a card takes its title from the anchor or thumbnail and its poster from the thumbnail", () => {
    const roxi = items.find((item) => item.external_id === "hqporn-UKsluts72");
    expect(roxi).toBeDefined();
    expect(roxi?.title).toBe("Roxi");
    expect(roxi?.source_url).toBe("https://hqporn.pics/galls/UKsluts72");
    expect(roxi?.poster_path).toBe("https://images.hqporn.pics/st/thumbs/413/Vs4qxmi49s.jpg");
  });
});

describe("query to category", () => {
  test("a query selects the categories whose name it matches", () => {
    expect(categoriesFor("Blonde")).toEqual(["blondes"]);
    expect(categoriesFor("no such category")).toEqual([]);
    expect(categoriesFor(undefined).length).toBeGreaterThan(10);
  });
});

describe("gallery page (offline)", () => {
  const results = galleryFromPage("https://hqporn.pics/galls/hardx146", loadFixture("gallery.html"));
  const [result] = results;

  test("the gallery is keyed by its slug and titled by the model", () => {
    expect(results.length).toBe(1);
    if (result.type !== "gallery") throw new Error("expected a gallery");
    expect(result.gallery.external_id).toBe("hqporn-hardx146");
    expect(result.gallery.source_url).toBe("https://hqporn.pics/galls/hardx146");
    expect(result.gallery.title).toBe("Mia Malkova");
  });

  test("the model and the site become performer and studio", () => {
    if (result.type !== "gallery") throw new Error("expected a gallery");
    expect(result.gallery.performers).toEqual([
      { name: "Mia Malkova", source_url: "https://hqporn.pics/model/Mia_Malkova" },
    ]);
    expect(result.gallery.studio).toEqual({
      name: "Hard X",
      external_id: "hqporn-site-Hard_X",
      source_url: "https://hqporn.pics/site/Hard_X",
    });
  });

  test("the images are the full-size fancybox links in page order", () => {
    if (result.type !== "gallery") throw new Error("expected a gallery");
    const images = result.gallery.images ?? [];
    expect(images.length).toBe(15);
    expect(images[0]).toEqual({
      type: "gallery",
      file_path: "https://images.hqporn.pics/pictures/hardx146/01.jpg",
      position: 0,
    });
    expect(images[14].position).toBe(14);
  });
});

// ---------------------------------------------------------------------------
// Live integration against hqporn.pics.

describe("gallery:list (live)", () => {
  test(
    "returns a bounded set of gallery stubs",
    async () => {
      const { items } = await galleryList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) {
        expect(item.media_type).toBe("gallery");
        expect(item.external_id).toStartWith("hqporn-");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "offset pages do not repeat the first page",
    async () => {
      const first = await galleryList({ limit: 5 });
      const second = await galleryList({ limit: 5, offset: 5 });
      const firstIds = new Set(first.items.map((item) => item.external_id));
      expect(second.items.length).toBe(5);
      expect(second.items.some((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );

  test("a query browses the category it names", async () => {
    const { items } = await galleryList({ query: "blonde", limit: 5 });
    expect(items.length).toBe(5);
    for (const item of items) expect(item.source_url).toStartWith("https://hqporn.pics/galls/");
  });
});

describe("gallery:find (live)", () => {
  test(
    "scrapes a listed gallery with its images",
    async () => {
      const { items } = await galleryList({ limit: 1 });
      const [result] = await scrapeGallery(items[0].source_url);
      expect(result.type).toBe("gallery");
      if (result.type !== "gallery") throw new Error("expected a gallery");
      expect(result.gallery.external_id).toBe(items[0].external_id);
      expect(result.gallery.images?.length).toBeGreaterThan(0);
    },
    2 * NET_TIMEOUT
  );
});
