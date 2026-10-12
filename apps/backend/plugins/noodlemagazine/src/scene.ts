import { parse, type HTMLElement } from "node-html-parser";
import type { DiscoveredItem, Scene, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { parseVideoCards } from "./cards.ts";
import { attributeOf, decodeEntities, fetchPage, normalizeText, videoIdFromUrl, watchUrl } from "./http.ts";

/** Fetches a watch page and returns the scene as the scrape stream's root record. */
export async function scrapeScene(url: string): Promise<ScrapeResult[]> {
  const videoId = videoIdFromUrl(url);
  if (videoId === "") throw new Error(`noodlemagazine: no video in ${url}`);
  const root = parse(await fetchPage(watchUrl(videoId)));
  return [{ type: "scene", scene: sceneFromWatchPage(root, videoId) }];
}

/** The content of a meta tag, entities decoded; empty when the tag is missing. */
export function metaContent(root: HTMLElement, property: string): string {
  const meta = root.querySelector(`meta[property="${property}"]`);
  if (meta === null) return "";
  return decodeEntities(attributeOf(meta, "content"));
}

/**
 * The scene record of a watch page. The stream is not part of it: its links
 * expire, so the resolve capability fetches a fresh one when the video plays.
 */
export function sceneFromWatchPage(root: HTMLElement, videoId: string): Scene {
  const scene: Scene = {
    type: "scene",
    external_id: `noodlemagazine-${videoId}`,
    source_url: watchUrl(videoId),
    title: normalizeText(metaContent(root, "og:title")),
    tags: tagsFrom(metaContent(root, "video:tag")),
    related: relatedFrom(root, videoId),
  };
  const date = metaContent(root, "ya:ovs:upload_date");
  if (/^\d{4}-\d{2}-\d{2}$/.test(date)) scene.date = date;
  const duration = Number(metaContent(root, "video:duration"));
  if (duration > 0) scene.duration_seconds = duration;
  const views = Number(metaContent(root, "ya:ovs:views_total"));
  if (views > 0) scene.view_count = views;
  const poster = metaContent(root, "og:image");
  if (poster !== "") scene.poster_path = poster;
  return scene;
}

/** The comma-separated tag list of the video:tag meta tag. */
export function tagsFrom(content: string): string[] {
  return content
    .split(",")
    .map((tag) => normalizeText(tag))
    .filter((tag) => tag !== "");
}

/** The page's related videos, without the video itself. */
function relatedFrom(root: HTMLElement, videoId: string): DiscoveredItem[] {
  return parseVideoCards(root).filter((item) => videoIdFromUrl(item.source_url) !== videoId);
}
