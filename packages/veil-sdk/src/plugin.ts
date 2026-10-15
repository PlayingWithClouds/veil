import { writeSync } from "fs";
import type {
  PluginMeta,
  ScrapeResult,
  ListArgs,
  ItemsResult,
  EnrichArgs,
  EnrichResult,
  ResolveResult,
  SuggestArgs,
  SuggestResult,
  CamListArgs,
  CamListResult,
  CamStreamArgs,
  FaceIdentifyArgs,
  FaceIdentifyResult,
} from "./types.ts";

/** Handler for an "<entity>:list" capability. */
export type ListHandler = (args: ListArgs) => Promise<ItemsResult>;

export interface PluginHandlers {
  meta: PluginMeta;
  // find serves every "<entity>:find" capability: fetch one full record by URL
  // and return the NDJSON stream items — root entity first, then related
  // standalone entities (performers, studio, tags, collection). The plugin
  // routes internally by URL shape; meta.capabilities declares which entity
  // roots it can produce.
  find?: (url: string) => Promise<ScrapeResult[]>;
  // Entity-scoped listings ("<entity>:list"). A set args.query means keyword
  // search; otherwise the site's paginated catalog.
  sceneList?: ListHandler;
  galleryList?: ListHandler;
  performerList?: ListHandler;
  studioList?: ListHandler;
  tagList?: ListHandler;
  enrich?: (args: EnrichArgs) => Promise<EnrichResult>;
  /** Serves the "stream:resolve" capability. */
  resolve?: (url: string) => Promise<ResolveResult>;
  /** Serves the "search:suggest" capability: the site's own search-box completions. */
  suggest?: (args: SuggestArgs) => Promise<SuggestResult>;
  /** Serves the "cam:list" capability: the rooms online right now. */
  camList?: (args: CamListArgs) => Promise<CamListResult>;
  /** Serves the "cam:stream" capability: a room's live HLS master playlist. */
  camStream?: (args: CamStreamArgs) => Promise<ResolveResult>;
  /** Serves the "face:identify" capability: performers who look like the face in a picture. */
  identifyFace?: (args: FaceIdentifyArgs) => Promise<FaceIdentifyResult>;
}

interface Invocation {
  capability: string;
  args?: Record<string, unknown>;
}

function respond<T>(result: T): never {
  writeSync(1, JSON.stringify({ result }) + "\n");
  process.exit(0);
}

function fail(message: string): never {
  writeSync(1, JSON.stringify({ error: message }) + "\n");
  process.exit(0);
}

function parseListArgs(args: Record<string, unknown> | undefined): ListArgs {
  const parsed: ListArgs = {};
  if (typeof args?.query === "string" && args.query !== "") parsed.query = args.query;
  if (typeof args?.url === "string" && args.url !== "") parsed.url = args.url;
  if (typeof args?.limit === "number") parsed.limit = args.limit;
  if (typeof args?.offset === "number") parsed.offset = args.offset;
  return parsed;
}

/** Reads the cam:list arguments, dropping anything of the wrong type. */
function parseCamListArgs(args: Record<string, unknown> | undefined): CamListArgs {
  const parsed: CamListArgs = {};
  if (typeof args?.query === "string" && args.query !== "") parsed.query = args.query;
  if (typeof args?.tag === "string" && args.tag !== "") parsed.tag = args.tag;
  if (typeof args?.gender === "string" && args.gender !== "") parsed.gender = args.gender;
  if (typeof args?.limit === "number") parsed.limit = args.limit;
  if (typeof args?.offset === "number") parsed.offset = args.offset;
  return parsed;
}

function listHandlers(handlers: PluginHandlers): Record<string, ListHandler | undefined> {
  return {
    "scene:list": handlers.sceneList,
    "gallery:list": handlers.galleryList,
    "performer:list": handlers.performerList,
    "studio:list": handlers.studioList,
    "tag:list": handlers.tagList,
  };
}

const FIND_CAPABILITIES = new Set([
  "scene:find",
  "gallery:find",
  "performer:find",
  "studio:find",
  "tag:find",
]);

// Emit one ScrapeResult per NDJSON line (no {"result":...} wrapper).
// Go reads lines until EOF.
async function runFind(handlers: PluginHandlers, input: Invocation): Promise<never> {
  if (!handlers.find) fail(`${input.capability} not implemented`);
  const url = input.args?.url;
  if (!url) fail(`args.url required for ${input.capability}`);

  const items = await handlers.find(String(url));
  for (const item of items) {
    writeSync(1, JSON.stringify(item) + "\n");
  }
  process.exit(0);
}

async function runList(handlers: PluginHandlers, input: Invocation): Promise<never> {
  const handler = listHandlers(handlers)[input.capability];
  if (!handler) fail(`${input.capability} not implemented`);
  respond<ItemsResult>(await handler(parseListArgs(input.args)));
}

export async function runPlugin(handlers: PluginHandlers): Promise<never> {
  const raw = await Bun.stdin.text();

  let input: Invocation;
  try {
    input = JSON.parse(raw);
  } catch {
    fail("invalid JSON on stdin");
  }

  try {
    if (input.capability === "meta") respond<PluginMeta>(handlers.meta);
    if (input.capability === "cam:list") {
      if (!handlers.camList) fail("cam:list not implemented");
      respond<CamListResult>(await handlers.camList(parseCamListArgs(input.args)));
    }

    if (input.capability === "cam:stream") {
      if (!handlers.camStream) fail("cam:stream not implemented");
      const username = input.args?.username;
      if (typeof username !== "string" || username === "") fail("args.username required for cam:stream");
      respond<ResolveResult>(await handlers.camStream({ username }));
    }

    if (FIND_CAPABILITIES.has(input.capability)) return await runFind(handlers, input);
    if (input.capability.endsWith(":list")) return await runList(handlers, input);

    if (input.capability === "enrich") {
      if (!handlers.enrich) fail("enrich not implemented");
      respond<EnrichResult>(await handlers.enrich(input.args as unknown as EnrichArgs));
    }

    if (input.capability === "stream:resolve") {
      if (!handlers.resolve) fail("stream:resolve not implemented");
      const url = input.args?.url;
      if (!url) fail("args.url required for stream:resolve");
      respond<ResolveResult>(await handlers.resolve(String(url)));
    }

    if (input.capability === "search:suggest") {
      if (!handlers.suggest) fail("search:suggest not implemented");
      const query = input.args?.query;
      if (typeof query !== "string" || query === "") fail("args.query required for search:suggest");
      let limit: number | undefined;
      if (typeof input.args?.limit === "number") limit = input.args.limit;
      respond<SuggestResult>(await handlers.suggest({ query, limit }));
    }

    if (input.capability === "face:identify") {
      if (!handlers.identifyFace) fail("face:identify not implemented");
      const imageBase64 = input.args?.image_base64;
      if (typeof imageBase64 !== "string" || imageBase64 === "") fail("args.image_base64 required for face:identify");
      let mimeType = "image/jpeg";
      if (typeof input.args?.mime_type === "string" && input.args.mime_type !== "") mimeType = input.args.mime_type;
      respond<FaceIdentifyResult>(await handlers.identifyFace({ image_base64: imageBase64, mime_type: mimeType }));
    }

    fail(`unsupported capability: ${input.capability}`);
  } catch (err) {
    fail(err instanceof Error ? err.message : String(err));
  }
}
