import { test, expect } from "bun:test";
import { extractMarkers, parseInitials } from "../src/initials.ts";

test("the player's action bar becomes point markers", async () => {
  const html = await Bun.file(`${import.meta.dir}/fixtures/video-acts.html`).text();
  expect(extractMarkers(parseInitials(html))).toEqual([
    { tag: "Cum in Mouth", seconds: 130 },
    { tag: "Doggy Style", seconds: 210 },
  ]);
});

test("pages without acts give no markers", () => {
  expect(extractMarkers(undefined)).toEqual([]);
  expect(extractMarkers({ xplayerPluginSettings: { actionsData: { actions: [] } } })).toEqual([]);
  expect(extractMarkers({ xplayerPluginSettings: { actionsData: { actions: [{ label: " ", time: 5 }, { label: "Anal" }] } } })).toEqual([]);
});
