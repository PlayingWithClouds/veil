import type { HTMLElement } from "node-html-parser";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import {
  externalId,
  parseClockSeconds,
  parseSiteDate,
  posterUrl,
  recordingStampFromUrl,
  resolveUrl,
  videoIdFromUrl,
} from "./http.ts";

/**
 * The video cards below `root`, in page order. Every card is an `li` whose
 * `a.moiclick1` links the video page; title, runtime, upload time and cam
 * site sit in sibling spans. The cards of one page never repeat a video.
 */
export function parseVideoCards(root: HTMLElement): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const anchor of root.querySelectorAll("a.moiclick1")) {
    const item = parseCard(anchor);
    if (!item || seen.has(item.external_id)) continue;
    seen.add(item.external_id);
    items.push(item);
  }
  return items;
}

/**
 * One card as a scene stub, or undefined when it has no usable title. The
 * site's search answers a name it does not know with a placeholder card that
 * has no title and no model; those are skipped here.
 */
function parseCard(anchor: HTMLElement): DiscoveredItem | undefined {
  const href = anchor.getAttribute("href") || "";
  const videoId = videoIdFromUrl(href);
  const title = (anchor.getAttribute("title") || "").trim();
  if (!videoId || !title) return undefined;

  const item: DiscoveredItem = {
    title,
    media_type: "scene",
    source_url: resolveUrl(href),
    external_id: externalId(videoId),
    poster_path: posterUrl(videoId),
  };
  const card = anchor.closest("li");
  if (!card) return item;

  const duration = parseClockSeconds(card.querySelector("span.dur")?.text || "");
  if (duration > 0) item.duration_seconds = duration;
  const date = parseSiteDate(card.querySelector("span.add")?.text || "");
  if (date) item.date = date;
  const siteName = (anchor.getAttribute("sitename") || "").trim();
  if (siteName) item.studio = { name: siteName };
  return item;
}

/**
 * The title a video page gets: the listing's `<model> <site> webcam rip
 * <date>_<time>`, rebuilt from the page's credits and the URL's recording
 * time, because the page itself only has the generic heading
 * "<model> <category> <site> CamRips".
 */
export function sceneTitle(modelName: string, siteName: string, videoUrl: string): string {
  const stamp = recordingStampFromUrl(videoUrl);
  if (!modelName || !siteName || !stamp) return "";
  return `${modelName} ${siteName} webcam rip ${stamp}`;
}
