import { test, expect, describe } from "bun:test";

import {
  resolveUrl,
  galleryIdFromUrl,
  isGalleryUrl,
  fullSizeImage,
  cleanTitle,
  decodeEntities,
  BASE_URL,
} from "../src/http.ts";
import { galleryList } from "../src/galleries.ts";
import { tagList } from "../src/categories.ts";
import { scrape } from "../src/scrape.ts";

// ---------------------------------------------------------------------------
// Pure URL/text helpers (no network)

describe("http helpers", () => {
  test("resolveUrl handles relative, absolute and protocol-relative", () => {
    expect(resolveUrl("/milf/")).toBe(`${BASE_URL}/milf/`);
    expect(resolveUrl("milf/")).toBe(`${BASE_URL}/milf/`);
    expect(resolveUrl("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
    expect(resolveUrl("//cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });

  test("galleryIdFromUrl extracts the trailing numeric id", () => {
    expect(galleryIdFromUrl("/galleries/hot-set-12345678/")).toBe("12345678");
    expect(galleryIdFromUrl(`${BASE_URL}/galleries/a-b-1/`)).toBe("1");
    expect(galleryIdFromUrl("/milf/")).toBe("");
  });

  test("isGalleryUrl detects gallery detail pages", () => {
    expect(isGalleryUrl("/galleries/hot-set-12345678/")).toBe(true);
    expect(isGalleryUrl("/milf/")).toBe(false);
  });

  test("fullSizeImage rewrites thumbnail paths to /1280/", () => {
    expect(fullSizeImage("https://cdni.pornpics.com/460/1/2/3/x.jpg")).toBe(
      "https://cdni.pornpics.com/1280/1/2/3/x.jpg"
    );
    expect(fullSizeImage("https://cdni.pornpics.com/300/1/2/3/x.jpg")).toBe(
      "https://cdni.pornpics.com/1280/1/2/3/x.jpg"
    );
  });

  test("cleanTitle strips the trailing gallery id and decodes entities", () => {
    expect(cleanTitle("Hot &amp; Wild Set 12345678")).toBe("Hot & Wild Set");
    expect(cleanTitle("Plain Title")).toBe("Plain Title");
  });

  test("decodeEntities handles named and numeric entities", () => {
    expect(decodeEntities("a &quot;b&quot; &#39;c&#x27;")).toBe("a \"b\" 'c'");
    expect(decodeEntities("&#65;&#x42;")).toBe("AB");
  });
});

// ---------------------------------------------------------------------------
// Live integration — real HTTP against pornpics.com. Chained so they survive
// catalog changes (scrape a gallery that gallery:list actually returned).

const NET_TIMEOUT = 30_000;
const QUERY = "milf";

describe("gallery:list (live)", () => {
  test(
    "returns a bounded set of gallery stubs",
    async () => {
      const { items } = await galleryList({ limit: 6 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(6);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("/galleries/");
        expect(item.external_id).toStartWith("pornpics-");
        expect(item.media_type).toBe("gallery");
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
      expect(second.items.length).toBeGreaterThan(0);
      expect(second.items.every((item) => firstIds.has(item.external_id))).toBe(false);
    },
    2 * NET_TIMEOUT
  );
});

describe("gallery:list with query (live)", () => {
  test(
    "returns well-formed items for a common query",
    async () => {
      const { items } = await galleryList({ query: QUERY, limit: 10 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(10);
      for (const item of items) {
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toContain("/galleries/");
        expect(item.external_id).toStartWith("pornpics-");
        expect(item.media_type).toBe("gallery");
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("tag catalog (live)", () => {
  test(
    "tag:list returns category stubs",
    async () => {
      const { items } = await tagList({ limit: 15 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(15);
      for (const item of items) {
        expect(item.media_type).toBe("tag");
        expect(item.title.trim()).not.toBe("");
        expect(item.source_url).toStartWith(BASE_URL);
        expect(item.external_id).toStartWith("pornpics-category-");
      }
    },
    2 * NET_TIMEOUT
  );

  test(
    "tag:list with query filters by name",
    async () => {
      const { items } = await tagList({ query: QUERY });
      expect(items.length).toBeGreaterThan(0);
      for (const item of items) {
        expect(item.title.toLowerCase()).toContain(QUERY);
      }
    },
    2 * NET_TIMEOUT
  );
});

describe("scrape (live)", () => {
  test(
    "scrapes a gallery with full-size images",
    async () => {
      const { items } = await galleryList({ query: QUERY, limit: 10 });
      expect(items.length).toBeGreaterThan(0);

      const results = await scrape(items[0].source_url);
      const gallery = results.flatMap((r) => (r.type === "gallery" ? [r.gallery] : []))[0];

      expect(gallery).toBeDefined();
      expect(gallery!.title.trim()).not.toBe("");
      expect(gallery!.external_id).toStartWith("pornpics-");
      expect(gallery!.source_url).toContain("/galleries/");
      expect((gallery!.images ?? []).length).toBeGreaterThan(0);

      for (const image of gallery!.images ?? []) {
        expect(image.file_path).toContain("/1280/");
      }
      expect(gallery!.cover_path).toBe(gallery!.images![0].file_path);
    },
    4 * NET_TIMEOUT
  );
});
