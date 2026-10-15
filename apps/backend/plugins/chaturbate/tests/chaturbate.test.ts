import { test, expect } from "bun:test";
import { buildRoomListQuery, toCamRoom } from "../src/rooms.ts";
import { toStreamResult } from "../src/stream.ts";

test("toCamRoom maps a public room", () => {
  expect(
    toCamRoom({
      username: "someone",
      gender: "c",
      tags: ["18", "anal"],
      num_users: 120,
      img: "https://thumb.live.mmcdn.com/riw/someone.jpg?1",
      room_subject: "hello",
      current_show: "public",
    }),
  ).toEqual({
    username: "someone",
    display_name: "someone",
    source_url: "https://chaturbate.com/someone/",
    thumbnail_url: "https://thumb.live.mmcdn.com/riw/someone.jpg?1",
    viewers: 120,
    tags: ["18", "anal"],
    subject: "hello",
    gender: "couple",
    online: true,
  });
});

test("toCamRoom marks private and password rooms as not watchable", () => {
  expect(toCamRoom({ username: "a", current_show: "private" }).online).toBe(false);
  expect(toCamRoom({ username: "a", current_show: "public", has_password: true }).online).toBe(false);
});

test("buildRoomListQuery caps the page size and maps filters", () => {
  const params = new URLSearchParams(
    buildRoomListQuery({ limit: 500, offset: 40, query: "anna", tag: "#Squirt", gender: "trans" }),
  );
  expect(params.get("limit")).toBe("100");
  expect(params.get("offset")).toBe("40");
  expect(params.get("keywords")).toBe("anna");
  expect(params.get("hashtags")).toBe("squirt");
  expect(params.get("genders")).toBe("t");
});

test("toStreamResult returns the HLS url with a room referer", () => {
  const result = toStreamResult("someone", { success: true, url: "https://edge.example/llhls.m3u8?token=x", room_status: "public" });
  expect(result.url).toBe("https://edge.example/llhls.m3u8?token=x");
  expect(result.mime_type).toBe("application/x-mpegURL");
  expect(result.headers?.Referer).toBe("https://chaturbate.com/someone/");
});

test("toStreamResult explains why a room cannot be watched", () => {
  expect(() => toStreamResult("someone", { success: true, url: "x", room_status: "private" })).toThrow("private");
  expect(() => toStreamResult("someone", { success: false, room_status: "offline" })).toThrow("offline");
});
