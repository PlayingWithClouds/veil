import { test, expect, afterEach } from "bun:test";
import { sceneList, modelVideosPath } from "../src/scenes.ts";

const rileyResponse = await Bun.file(`${import.meta.dir}/fixtures/model-riley-reid.json`).text();
const pageOne = await Bun.file(`${import.meta.dir}/fixtures/model-videos-riley-reid-1.json`).text();
const pageTwo = await Bun.file(`${import.meta.dir}/fixtures/model-videos-riley-reid-2.json`).text();
const PROFILE = "https://txxx.com/models/riley-reid/";
const originalFetch = globalThis.fetch;
const requested: string[] = [];

// Serves the model record, two pages of videos and, like the API, an error body past the end.
function mockApi(): void {
  requested.length = 0;
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = String(input);
    requested.push(url);
    if (url.includes("/api/json/model/")) return new Response(rileyResponse);
    if (url.endsWith(modelVideosPath("41056", 1))) return new Response(pageOne);
    if (url.endsWith(modelVideosPath("41056", 2))) return new Response(pageTwo);
    return new Response(JSON.stringify({ error: 1, code: "not_found" }));
  }) as typeof fetch;
}

afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("model videos path filters the videos2 API by model id", () => {
  expect(modelVideosPath("41056", 3)).toBe("/api/json/videos2/14400/str/latest-updates/60/model.41056.3.all...jsond");
});

test("a model url lists that model's videos with the normal external ids", async () => {
  mockApi();
  const result = await sceneList({ url: PROFILE, limit: 3 });
  const pageIds = (JSON.parse(pageOne).videos as { video_id: string }[]).map((video) => `txxx-${video.video_id}`);
  expect(result.items.map((item) => item.external_id)).toEqual(pageIds);
  expect(result.items[0].source_url).toStartWith("https://txxx.com/videos/");
  expect(requested.some((url) => url.endsWith(modelVideosPath("41056", 2)))).toBe(false);
});

test("offset continues on the next page and stops past the end", async () => {
  mockApi();
  const second = await sceneList({ url: PROFILE, offset: 3, limit: 2 });
  const pageTwoIds = (JSON.parse(pageTwo).videos as { video_id: string }[]).map((video) => `txxx-${video.video_id}`);
  expect(second.items.map((item) => item.external_id)).toEqual(pageTwoIds.slice(0, 2));

  const beyond = await sceneList({ url: PROFILE, offset: 6, limit: 5 });
  expect(beyond.items).toEqual([]);
});

test("without a limit every page is listed", async () => {
  mockApi();
  const result = await sceneList({ url: PROFILE });
  expect(result.items).toHaveLength(6);
});
