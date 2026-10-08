import { afterEach, beforeEach, expect, test } from "bun:test";
import { Context } from "@neoworks/extension-system";
import type { RouteHandler, SourceProvider } from "@atlas/contracts/server";
import { createVeilProvider } from "../src/provider";
import VeilPlugin from "../src/server";
import { createVeilApi } from "../src/veilApi";

interface MockState {
  ensureGalleryCalls: string[];
  ensureSceneCalls: string[];
  streamCalls: string[];
}

const GALLERY_IMAGES: Record<string, { filePath: string; position: number }[]> = {
  stub: [],
  full: [
    { filePath: "https://cdn.example/b.jpg", position: 2 },
    { filePath: "https://cdn.example/a.jpg", position: 1 },
  ],
};

const SCRAPED_IMAGES = [
  { filePath: "https://cdn.example/s1.jpg", position: 1 },
  { filePath: "https://cdn.example/s2.jpg", position: 2 },
  { filePath: "https://cdn.example/s3.jpg", position: 3 },
];

let state: MockState;
let server: ReturnType<typeof Bun.serve>;
let baseUrl: string;

function answer(query: string, variables: Record<string, any>): Record<string, unknown> {
  if (query.includes("ensureGalleryImages")) {
    state.ensureGalleryCalls.push(JSON.stringify(variables));
    return { ensureGalleryImages: { title: "Scraped", images: SCRAPED_IMAGES } };
  }
  if (query.includes("ensureSceneStreams")) {
    state.ensureSceneCalls.push(query);
    return {
      ensureSceneStreams: [
        { url: "https://embed.example/download", kind: "download", format: "mkv", mimeType: null },
        { url: "https://embed.example/play", kind: "stream", format: "mp4", mimeType: "video/mp4" },
      ],
    };
  }
  if (query.includes("stream(")) {
    state.streamCalls.push(JSON.stringify(variables));
    return {
      stream: { url: `https://cdn.example/video-${state.streamCalls.length}.mp4`, headers: [{ name: "Referer", value: "https://embed.example/" }] },
    };
  }
  if (query.includes("scene(")) {
    return { scene: { title: "A scene", durationSeconds: 125 } };
  }
  if (query.includes("galleries(")) {
    return { galleries: [{ id: "stub", title: "Stub", coverPath: "https://cdn.example/c.jpg", imageCount: 0 }, { id: "full", title: "Full", coverPath: null, imageCount: 2 }] };
  }
  if (query.includes("scenes(")) {
    return { scenes: [{ id: "7", title: "Seven", posterPath: "https://cdn.example/p.jpg", durationSeconds: 60 }] };
  }
  if (query.includes("gallery(")) {
    const galleryId = Object.values(variables).includes("full") ? "full" : "stub";
    return { gallery: { title: galleryId, images: GALLERY_IMAGES[galleryId] } };
  }
  throw new Error(`unexpected query: ${query}`);
}

beforeEach(() => {
  state = { ensureGalleryCalls: [], ensureSceneCalls: [], streamCalls: [] };
  server = Bun.serve({
    port: 0,
    async fetch(request) {
      const body = (await request.json()) as { query: string; variables?: Record<string, any> };
      try {
        return Response.json({ data: answer(body.query, body.variables ?? {}) });
      } catch (error) {
        return Response.json({ errors: [{ message: (error as Error).message }] });
      }
    },
  });
  baseUrl = `http://localhost:${server.port}`;
});

afterEach(() => {
  server.stop(true);
});

function makeProvider(now?: () => number): SourceProvider {
  return createVeilProvider(createVeilApi(baseUrl), now);
}

test("list galleries and scenes proxies thumbnails through veil", async () => {
  const provider = makeProvider();
  const galleries = await provider.list!("gallery", { search: "", limit: 10, offset: 0 });
  expect(galleries.items[0].thumbnail).toBe(`${baseUrl}/api/img?url=${encodeURIComponent("https://cdn.example/c.jpg")}`);
  expect(galleries.items[1].thumbnail).toBeUndefined();
  const scenes = await provider.list!("scene", { search: "x", limit: 10, offset: 0 });
  expect(scenes.items.map((item) => item.id)).toEqual(["7"]);
});

test("stub gallery triggers ensureGalleryImages, full gallery does not", async () => {
  const provider = makeProvider();
  const scraped = await provider.resolve("gallery", { id: "stub" });
  expect(state.ensureGalleryCalls).toHaveLength(1);
  expect(scraped.items.map((item) => item.ref)).toEqual([
    "veil:image:https://cdn.example/s1.jpg",
    "veil:image:https://cdn.example/s2.jpg",
    "veil:image:https://cdn.example/s3.jpg",
  ]);
  const full = await provider.resolve("gallery", { id: "full" });
  expect(state.ensureGalleryCalls).toHaveLength(1);
  expect(full.items.map((item) => item.ref)).toEqual([
    "veil:image:https://cdn.example/a.jpg",
    "veil:image:https://cdn.example/b.jpg",
  ]);
  expect(full.items.every((item) => item.mediaKind === "image")).toBe(true);
});

test("scene resolve picks the best stream and returns a stable ref plus duration", async () => {
  const provider = makeProvider();
  const resolved = await provider.resolve("scene", { id: "7" });
  expect(resolved.items).toEqual([{ ref: "veil:scene:7", mediaKind: "video" }]);
  expect(resolved.sessionMeta).toEqual({ duration: 125 });
  expect(resolved.label).toBe("A scene");
  expect(state.streamCalls[0]).toContain("https://embed.example/play");
  const location = await provider.locate("veil:scene:7");
  expect(location).toEqual({
    kind: "url",
    url: "https://cdn.example/video-1.mp4",
    headers: { Referer: "https://embed.example/" },
  });
});

test("locate caches the scene stream for ten minutes, then re-resolves", async () => {
  let clock = 1_000;
  const provider = makeProvider(() => clock);
  await provider.locate("veil:scene:9");
  await provider.locate("veil:scene:9");
  expect(state.streamCalls).toHaveLength(1);
  clock += 10 * 60 * 1000 + 1;
  const refreshed = await provider.locate("veil:scene:9");
  expect(state.streamCalls).toHaveLength(2);
  expect(refreshed).toMatchObject({ url: "https://cdn.example/video-2.mp4" });
});

test("locate image ref returns the veil proxy url", async () => {
  const provider = makeProvider();
  const location = await provider.locate("veil:image:https://cdn.example/a.jpg?x=1&y=2");
  expect(location).toEqual({
    kind: "url",
    url: `${baseUrl}/api/img?url=${encodeURIComponent("https://cdn.example/a.jpg?x=1&y=2")}`,
  });
  await expect(provider.locate("other:ref")).rejects.toMatchObject({ status: 400 });
});

test("random sample returns at most the requested count", async () => {
  const provider = makeProvider();
  const sample = await provider.resolve("random", { count: 3 });
  expect(sample.items).toHaveLength(3);
  expect(new Set(sample.items.map((item) => item.ref)).size).toBe(3);
  await expect(provider.resolve("random", { count: 0 })).rejects.toMatchObject({ status: 400 });
});

test("plugin registers the provider and config route, and dispose removes them", async () => {
  const providers = new Map<string, SourceProvider>();
  const routes = new Map<string, RouteHandler>();
  const root = new Context();
  root.provide("sources", {
    register: (provider: SourceProvider) => {
      providers.set(provider.id, provider);
      return () => providers.delete(provider.id);
    },
  } as never);
  root.provide("http", {
    route: (method: string, pattern: string, handler: RouteHandler) => {
      routes.set(`${method} ${pattern}`, handler);
      return () => routes.delete(`${method} ${pattern}`);
    },
  } as never);

  const fiber = root.plugin(VeilPlugin as never, { url: `${baseUrl}/` } as never);
  await fiber;
  expect(providers.has("veil")).toBe(true);
  const configRoute = routes.get("GET /api/plugins/veil/config");
  expect(await configRoute!(new Request("http://x"), {})).toEqual({ url: baseUrl });

  await fiber.dispose();
  expect(providers.has("veil")).toBe(false);
  expect(routes.size).toBe(0);
});
