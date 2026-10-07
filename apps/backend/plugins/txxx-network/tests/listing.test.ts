import { test, expect } from "bun:test";
import { listingStudio, type ApiVideo } from "../src/listing.ts";

const video: ApiVideo = { video_id: "1", title: "Video", dir: "video" };

test("listingStudio credits the content source first", () => {
  expect(listingStudio("txxx.com", { ...video, content_source_name: "Brazzers", user_id: "7", display_name: "Uploader" })).toEqual({
    name: "Brazzers",
  });
});

test("listingStudio falls back to the member with the video page's identity", () => {
  expect(listingStudio("txxx.com", { ...video, content_source_name: "", user_id: "6041293", display_name: "Angelina Barbosa" })).toEqual({
    name: "Angelina Barbosa",
    external_id: "txxx-member-6041293@txxx.com",
    source_url: "https://txxx.com/members/6041293/",
  });
});

test("listingStudio names nobody when the entry has no publisher", () => {
  expect(listingStudio("txxx.com", video)).toBeUndefined();
});
