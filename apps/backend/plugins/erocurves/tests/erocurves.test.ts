import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { decodeEntities, stripTags } from "../src/http.ts";
import { galleryFromPost, imageUrlsFromContent, studioFromExcerpt, stubFromPost, type WordPressPost } from "../src/posts.ts";
import { slugFromUrl, scrapeGallery } from "../src/galleries.ts";
import { galleryList } from "../src/listing.ts";

const NET_TIMEOUT = 30_000;

/** Loads a saved WordPress REST response from the fixtures folder. */
function loadPosts(fileName: string): WordPressPost[] {
  return JSON.parse(readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8")) as WordPressPost[];
}

describe("text helpers", () => {
  test("decodeEntities handles named and numeric entities", () => {
    expect(decodeEntities("Tom &amp; Jerry &#8217;s &#x27;x&#x27;")).toBe("Tom & Jerry ’s 'x'");
  });

  test("stripTags keeps the text between tags", () => {
    expect(stripTags("<p>Get More Of This Babe At <b>Met Art</b></p>\n")).toBe("Get More Of This Babe At Met Art");
  });

  test("slugFromUrl takes the last path segment", () => {
    expect(slugFromUrl("https://www.erocurves.com/vina-rae-in-exotic-taste/")).toBe("vina-rae-in-exotic-taste");
    expect(slugFromUrl("https://www.erocurves.com/")).toBe("");
  });
});

describe("post parsing (offline)", () => {
  const [embeddedPost] = loadPosts("posts-embed.json");

  test("imageUrlsFromContent returns the full-size gallery links once each", () => {
    const urls = imageUrlsFromContent(embeddedPost.content.rendered);
    expect(urls.length).toBeGreaterThan(0);
    expect(new Set(urls).size).toBe(urls.length);
    expect(urls[0]).toBe(
      "https://cdn.erocurves.com/wp-content/uploads/2026/10/12/MetArt_Exotic-Taste_Vina-Rae_medium_0003.jpg"
    );
    for (const url of urls) expect(url).not.toContain("-172x262");
  });

  test("studioFromExcerpt reads the studio named in the excerpt", () => {
    expect(studioFromExcerpt("<p>Get More Of This Babe At Met Art</p>\n")).toBe("Met Art");
    expect(studioFromExcerpt("<p>Nothing to see</p>")).toBeUndefined();
  });

  test("galleryFromPost maps title, studio, performers, tags and images", () => {
    const gallery = galleryFromPost(embeddedPost);
    expect(gallery.external_id).toBe("erocurves-1692060");
    expect(gallery.source_url).toBe("https://www.erocurves.com/vina-rae-in-exotic-taste/");
    expect(gallery.title).toBe("Vina Rae in Exotic Taste");
    expect(gallery.studio).toEqual({ name: "Met Art" });
    expect(gallery.performers).toEqual([{ name: "Vina Rae" }]);
    expect(gallery.tags).toEqual(["Asian", "Brunette"]);
    expect(gallery.cover_path).toBe(gallery.images?.[0].file_path);
    expect(gallery.images?.[1].position).toBe(1);
  });

  test("sourceUrlOf stays on the site even when the permalink points elsewhere", () => {
    const offsitePost = { ...embeddedPost, slug: "horny-petite-redhead", link: "https://www.brdteengal.com/x/" };
    expect(galleryFromPost(offsitePost).source_url).toBe("https://www.erocurves.com/horny-petite-redhead/");
  });

  test("stubFromPost builds a gallery stub with the first image as poster", () => {
    const [listPost] = loadPosts("posts-list.json");
    const stub = stubFromPost(listPost);
    expect(stub.media_type).toBe("gallery");
    expect(stub.external_id).toBe(`erocurves-${listPost.id}`);
    expect(stub.source_url).toBe(`https://www.erocurves.com/${listPost.slug}/`);
    expect(stub.poster_path).toStartWith("https://cdn.erocurves.com/");
    expect(stub.count).toBe(imageUrlsFromContent(listPost.content.rendered).length);
  });
});

// ---------------------------------------------------------------------------
// Live integration against erocurves.com.

describe("gallery:list (live)", () => {
  test(
    "returns a bounded set of gallery stubs",
    async () => {
      const { items } = await galleryList({ limit: 5 });
      expect(items.length).toBe(5);
      for (const item of items) {
        expect(item.media_type).toBe("gallery");
        expect(item.external_id).toStartWith("erocurves-");
        expect(item.source_url).toStartWith("https://www.erocurves.com/");
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

  test(
    "query runs the site search",
    async () => {
      const { items } = await galleryList({ query: "asian", limit: 5 });
      expect(items.length).toBeGreaterThan(0);
      expect(items.length).toBeLessThanOrEqual(5);
    },
    2 * NET_TIMEOUT
  );
});

describe("gallery:find (live)", () => {
  test(
    "scrapes a listed gallery with images",
    async () => {
      const { items } = await galleryList({ limit: 1 });
      const [results] = await scrapeGallery(items[0].source_url);
      expect(results.type).toBe("gallery");
      if (results.type !== "gallery") throw new Error("expected a gallery");
      expect(results.gallery.external_id).toBe(items[0].external_id);
      expect(results.gallery.title.trim()).not.toBe("");
      expect(results.gallery.images?.length).toBeGreaterThan(0);
    },
    2 * NET_TIMEOUT
  );
});
