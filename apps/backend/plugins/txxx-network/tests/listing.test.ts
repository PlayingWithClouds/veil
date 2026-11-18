import { test, expect } from "bun:test";
import { listingStudio, listingUploader, type ApiVideo } from "../src/listing.ts";

const video: ApiVideo = { video_id: "1", title: "Video", dir: "video" };

test("listingStudio credits the content source", () => {
  expect(listingStudio({ ...video, content_source_name: "Brazzers", user_id: "7", display_name: "Uploader" })).toEqual({
    name: "Brazzers",
  });
});

test("listingStudio names nobody for a member's upload", () => {
  expect(listingStudio({ ...video, content_source_name: "", user_id: "6041293", display_name: "Angelina Barbosa" })).toBeUndefined();
  expect(listingStudio(video)).toBeUndefined();
});

test("listingUploader names the member with the video page's identity", () => {
  expect(listingUploader("txxx.com", { ...video, user_id: "6041293", display_name: "Angelina Barbosa" })).toEqual({
    name: "Angelina Barbosa",
    external_id: "txxx-member-6041293@txxx.com",
    url: "https://txxx.com/members/6041293/",
  });
  expect(listingUploader("txxx.com", video)).toBeUndefined();
});
