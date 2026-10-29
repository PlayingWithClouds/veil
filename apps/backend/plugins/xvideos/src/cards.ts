import { parse, type HTMLElement } from "node-html-parser";
import { relativeDate } from "@playingwithclouds/veil-sdk";
import type { DiscoveredItem, StudioRef } from "@playingwithclouds/veil-sdk";
import {
  canonicalUrl,
  countFromText,
  decodeEntities,
  ratingFromPercent,
  secondsFromLabel,
  type Site,
} from "./http.ts";

/** The stable id of a video: its encoded id ("hofbtbke2c2") on its site. */
export function externalId(site: Site, encodedId: string): string {
  return `${site}-${encodedId}`;
}

/** The uploader a card or page links: a channel ("/bubblebooties"), pornstar or user profile. */
export function studioFromLink(name: string | undefined, path: string | undefined, site: Site): StudioRef | undefined {
  const trimmed = name ? decodeEntities(name).trim() : "";
  if (trimmed === "") return undefined;
  const studio: StudioRef = { name: trimmed };
  if (path && path.startsWith("/")) {
    studio.source_url = canonicalUrl(path, site);
    studio.external_id = `${site}-uploader-${path.replace(/^\/+/, "").replace(/\//g, "-")}`;
  }
  return studio;
}

/**
 * Reads the video cards (".thumb-block") of a listing page: search, newest,
 * tag pages. Each card carries title, page, thumbnail, hover clip, coarse
 * runtime, views and the uploader.
 */
export function parseCardsHtml(html: string, site: Site): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  for (const block of parse(html).querySelectorAll(".thumb-block")) {
    const item = cardItem(block, site);
    if (item) items.push(item);
  }
  return items;
}

/** One listing card as an item; undefined for ads and blocks without a video. */
function cardItem(block: HTMLElement, site: Site): DiscoveredItem | undefined {
  const encodedId = block.getAttribute("data-eid");
  const link = block.querySelector(".thumb-under a") || block.querySelector(".thumb a");
  const path = link?.getAttribute("href");
  if (!encodedId || !path) return undefined;

  const title = decodeEntities(link?.getAttribute("title") || link?.text || "").trim();
  const image = block.querySelector(".thumb img");
  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: canonicalUrl(path, site),
    external_id: externalId(site, encodedId),
  };
  const poster = image?.getAttribute("data-src");
  if (poster) item.poster_path = poster;
  const preview = image?.getAttribute("data-pvv");
  if (preview) item.preview_video = preview;
  const duration = secondsFromLabel(block.querySelector(".metadata .duration")?.text || block.querySelector(".metadata")?.text);
  if (duration > 0) item.duration_seconds = duration;
  const views = countFromText(viewsText(block));
  if (views !== undefined) item.view_count = views;
  const uploader = block.querySelector(".metadata a");
  const studio = studioFromLink(uploader?.querySelector(".name")?.text, uploader?.getAttribute("href"), site);
  if (studio) item.studio = studio;
  return item;
}

/** The views figure of a card's metadata line ("931.8k Views"). */
function viewsText(block: HTMLElement): string | undefined {
  const text = block.querySelector(".metadata")?.text || "";
  return text.match(/([\d.,]+\s*[kKmMbB]?)\s*Views/)?.[1];
}

/** One video as the sites' JSON describes it: related videos, profile and channel listings. */
export interface JsonVideo {
  eid?: string;
  /** page path */
  u?: string;
  /** full title, else the shortened one */
  tf?: string;
  t?: string;
  /** thumbnail */
  i?: string;
  /** hover clip */
  ipu?: string;
  /** "12 min" */
  d?: string;
  /** "99%" */
  r?: string;
  /** "4.3M" views */
  n?: string;
  /** uploader name and page */
  pn?: string;
  pu?: string;
  /** "12 years ago" */
  ut?: string | null;
}

/** The videos of a JSON listing: an array or an object keyed by position, as the endpoints vary. */
export function jsonVideoList(videos: unknown): JsonVideo[] {
  if (Array.isArray(videos)) return videos as JsonVideo[];
  if (videos && typeof videos === "object") return Object.values(videos) as JsonVideo[];
  return [];
}

/**
 * The page path of a JSON video. Channel listings give a click-counting
 * address ("/prof-video-click/upload/<channel>/<id>/<slug>") instead, so the
 * page is rebuilt from id and slug as each site writes it.
 */
function videoPath(path: string, encodedId: string, site: Site): string {
  if (path.startsWith("/video")) return path;
  const slug = path.split("/").filter((segment) => segment !== "").pop() || "";
  if (site === "xnxx") return `/video-${encodedId}/${slug}`;
  return `/video.${encodedId}/${slug}`;
}

/** A JSON video as an item; undefined without id or page. */
export function jsonItem(video: JsonVideo, site: Site): DiscoveredItem | undefined {
  if (!video.eid || !video.u) return undefined;
  const item: DiscoveredItem = {
    title: decodeEntities(video.tf || video.t || "").trim(),
    media_type: "scene",
    source_url: canonicalUrl(videoPath(video.u, video.eid, site), site),
    external_id: externalId(site, video.eid),
  };
  if (video.i) item.poster_path = video.i;
  if (video.ipu) item.preview_video = video.ipu;
  const duration = secondsFromLabel(video.d);
  if (duration > 0) item.duration_seconds = duration;
  const rating = ratingFromPercent(video.r);
  if (rating !== undefined) item.rating = rating;
  const views = countFromText(video.n);
  if (views !== undefined) item.view_count = views;
  const studio = studioFromLink(video.pn, video.pu, site);
  if (studio) item.studio = studio;
  if (video.ut) {
    const date = relativeDate(video.ut);
    if (date) item.date = date;
  }
  return item;
}
