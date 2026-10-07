import {
  fetchJson,
  hostFromUrl,
  baseUrl,
  videoIdFromUrl,
  videoApiPath,
  relatedApiPath,
} from "./http.ts";
import { videoUrl, parseDurationSeconds, videoToDiscoveredItem, hostScopedId, type ApiVideo } from "./listing.ts";
import { fetchSources } from "./streams.ts";
import type {
  ScrapeResult,
  Scene,
  ScenePerformer,
  StudioRef,
  Studio,
  Download,
  Image,
  DiscoveredItem,
} from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// A named entity dict as returned by the API: { "<id>": { title, dir }, ... }.
type EntityMap = Record<string, { title?: string; dir?: string }> | null | undefined;

// The video's channel (paysite/brand). The API sends an empty array when the
// video has none.
interface ApiChannel {
  id?: string;
  dir?: string;
  title?: string;
  cf3?: string;
}

// The uploading member account.
interface ApiUser {
  id?: string;
  username?: string;
  avatar?: string;
}

export interface ApiVideoDetail {
  video_id: string;
  title: string;
  description?: string;
  dir?: string;
  duration?: string;
  post_date?: string;
  thumb?: string | null;
  thumbsrc?: string | null;
  pv?: string | null;
  categories?: EntityMap;
  tags?: EntityMap;
  models?: EntityMap;
  models_suggested?: string[] | null;
  channel?: ApiChannel | unknown[] | null;
  source_title?: string;
  source_dir?: string;
  csg_title?: string;
  csg_dir?: string;
  user?: ApiUser | null;
}

// The content source groups member uploads are filed under ("UsersUpload" on
// txxx, "UserUpload" on hclips); they are not real sources and have no page.
const USER_UPLOAD_SOURCE_GROUPS = new Set(["usersupload", "userupload"]);

// Every TxxxNetwork item is a single adult video, modelled as one Scene. Detail
// comes from the site's JSON API rather than the SPA-rendered HTML page.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const host = hostFromUrl(url);
  const id = videoIdFromUrl(url);
  if (!id) throw new Error(`txxx-network: no video id in ${url}`);

  const response = await fetchJson<{ video?: ApiVideoDetail }>(host, videoApiPath(id));
  const detail = response.video;
  if (!detail) throw new Error(`txxx-network: no video record for ${url}`);

  const poster = withScheme(detail.thumb);
  const [downloads, related] = await Promise.all([
    extractDownloads(host, id, url),
    fetchRelated(host, id),
  ]);

  const scene: Scene = {
    type: "scene",
    external_id: `txxx-${id}`,
    source_url: videoUrl(host, id, detail.dir),
    title: (detail.title ?? "").trim(),
    details: cleanDescription(detail.description, detail.title),
    date: normalizeDate(detail.post_date),
    duration_seconds: parseDurationSeconds(detail.duration),
    poster_path: poster,
    preview_video: withScheme(detail.pv),
    images: poster ? [{ type: "poster", file_path: poster } as Image] : [],
    performers: extractPerformers(host, detail),
    tags: [...entityNames(detail.categories), ...entityNames(detail.tags)],
    downloads,
    related,
  };

  const credit = creditedStudio(host, detail);
  if (!credit) {
    scene.studio = siteStudio(host);
    return [{ type: "scene", scene }];
  }
  scene.studio = { name: credit.name, external_id: credit.external_id };
  if (credit.source_url) {
    scene.studio.source_url = credit.source_url;
  }
  return [
    { type: "scene", scene },
    { type: "studio", studio: credit },
  ];
}

/**
 * The entity that published the video, parented to the network site: its
 * channel, else its content source, else the uploading member. Null when the
 * API names none of them, leaving the site itself as the studio.
 */
export function creditedStudio(host: string, detail: ApiVideoDetail): Studio | null {
  const parent = siteStudio(host).name;

  const channel = channelOf(detail);
  if (channel) {
    return {
      type: "studio",
      external_id: hostScopedId(host, "channel", channel.id as string),
      source_url: channelUrl(host, channel.dir as string),
      name: (channel.title as string).trim(),
      image_path: withScheme(channel.cf3),
      parent,
    };
  }

  const source = contentSourceStudio(host, detail);
  if (source) {
    source.parent = parent;
    return source;
  }

  const user = detail.user;
  if (!user || !user.id || !(user.username ?? "").trim()) return null;
  return {
    type: "studio",
    external_id: hostScopedId(host, "member", user.id),
    source_url: `${baseUrl(host)}/members/${user.id}/`,
    name: (user.username as string).trim(),
    image_path: withScheme(user.avatar),
    parent,
  };
}

/** The video's channel when it has a usable one (the API sends [] for none). */
function channelOf(detail: ApiVideoDetail): ApiChannel | null {
  const channel = detail.channel;
  if (!channel || Array.isArray(channel)) return null;
  if (!channel.id || !channel.dir || !(channel.title ?? "").trim()) return null;
  return channel;
}

/**
 * The video's content source: `source_title` (whose page is the channel page
 * of `source_dir`), else the source group `csg_title`, which has no page and
 * is identified by name. Member uploads' placeholder group is ignored.
 */
function contentSourceStudio(host: string, detail: ApiVideoDetail): Studio | null {
  const sourceTitle = (detail.source_title ?? "").trim();
  const sourceDir = (detail.source_dir ?? "").trim();
  if (sourceTitle && sourceDir) {
    return {
      type: "studio",
      external_id: hostScopedId(host, "source", sourceDir),
      source_url: channelUrl(host, sourceDir),
      name: sourceTitle,
    };
  }

  const groupTitle = (detail.csg_title ?? "").trim();
  const groupDir = (detail.csg_dir ?? "").trim();
  if (!groupTitle || !groupDir || USER_UPLOAD_SOURCE_GROUPS.has(groupDir.toLowerCase())) return null;
  return {
    type: "studio",
    external_id: hostScopedId(host, "source-group", groupDir),
    source_url: "",
    name: groupTitle,
  };
}

/** A channel's page on the site. */
function channelUrl(host: string, dir: string): string {
  return `${baseUrl(host)}/channel/${dir}/`;
}

// The SPA loads the "related videos" block from a separate endpoint, so this is
// one extra request. A failure only drops the related list, not the scene.
async function fetchRelated(host: string, id: string): Promise<DiscoveredItem[]> {
  try {
    const response = await fetchJson<{ videos?: ApiVideo[] }>(host, relatedApiPath(id, MAX_RELATED + 1));
    return relatedItems(host, id, response.videos);
  } catch {
    return [];
  }
}

export function relatedItems(
  host: string,
  sceneId: string,
  videos: ApiVideo[] | undefined
): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>([`txxx-${sceneId}`]);

  for (const video of videos ?? []) {
    const item = videoToDiscoveredItem(host, video);
    if (!item || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
    if (items.length >= MAX_RELATED) break;
  }

  return items;
}

/**
 * The network site itself: the parent of every channel/uploader studio, and
 * the scene's studio when the API credits no publisher, so a whole site's
 * catalog can carry shared tags and be browsed as one source.
 */
function siteStudio(host: string): StudioRef {
  const label = host.replace(/\.(com|net|org)$/, "");
  return {
    name: label.charAt(0).toUpperCase() + label.slice(1),
    source_url: baseUrl(host),
  };
}

// Every available quality is offered so the UI can present them all for
// download; each is also directly streamable. Falls back to the page URL as a
// resolvable handle if the stream API is momentarily unavailable.
async function extractDownloads(host: string, id: string, pageUrl: string): Promise<Download[]> {
  try {
    const sources = await fetchSources(host, id, pageUrl);
    if (sources.length > 0) {
      return sources.map((source) => ({
        label: source.quality,
        url: source.url,
        quality: source.quality,
        format: source.format,
      }));
    }
  } catch {
    // Fall through to the page-URL handle below.
  }
  return [{ label: "txxx", url: pageUrl }];
}

/**
 * The credited models, each with their /models/ page. Most uploads leave
 * `models` empty and only carry the site's name-only `models_suggested`, which
 * is used as the fallback.
 */
export function extractPerformers(host: string, detail: ApiVideoDetail): ScenePerformer[] {
  const credited = modelPerformers(host, detail.models);
  if (credited.length > 0) return credited;
  return suggestedPerformers(detail.models_suggested);
}

/** Performers from the `models` dict, de-duplicated by name. */
function modelPerformers(host: string, models: EntityMap): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const model of Object.values(models ?? {})) {
    const name = (model?.title ?? "").trim();
    if (!name || seen.has(name)) continue;
    seen.add(name);
    const performer: ScenePerformer = { name, order: performers.length };
    if (model.dir) {
      performer.source_url = `${baseUrl(host)}/models/${model.dir}/`;
    }
    performers.push(performer);
  }
  return performers;
}

/** Performers from the name-only `models_suggested` list, de-duplicated. */
function suggestedPerformers(suggested: string[] | null | undefined): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const entry of suggested ?? []) {
    const name = entry.trim();
    if (!name || seen.has(name)) continue;
    seen.add(name);
    performers.push({ name, order: performers.length });
  }
  return performers;
}

function entityNames(entities: EntityMap): string[] {
  if (!entities) return [];
  const names: string[] = [];
  for (const entry of Object.values(entities)) {
    const name = (entry?.title ?? "").trim();
    if (name) names.push(name);
  }
  return names;
}

function withScheme(url: string | null | undefined): string | undefined {
  if (!url) return undefined;
  if (url.startsWith("http")) return url;
  if (url.startsWith("//")) return "https:" + url;
  return "https://" + url;
}

function normalizeDate(postDate: string | undefined): string | undefined {
  if (!postDate) return undefined;
  const match = postDate.match(/^(\d{4}-\d{2}-\d{2})/);
  return match ? match[1] : undefined;
}

// The API repeats the title as the description for many videos; drop that noise.
function cleanDescription(
  description: string | undefined,
  title: string | undefined
): string | undefined {
  const text = (description ?? "").trim();
  if (!text) return undefined;
  if (text === (title ?? "").trim()) return undefined;
  return text;
}
