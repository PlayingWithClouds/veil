import { BROWSER_HEADERS, fetchHtml } from "@playingwithclouds/veil-sdk";

export const BASE_URL = "https://www.showcamrips.com";

/** Site root with the language prefix every listing lives under. */
export const HOME_URL = `${BASE_URL}/en/`;

/** Both the play page and the video host answer only to a showcamrips referer. */
export const PLAYBACK_HEADERS = {
  Referer: `${BASE_URL}/`,
  "User-Agent": BROWSER_HEADERS["User-Agent"],
};

/** An absolute URL for a site-relative or protocol-relative link. */
export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  if (path.startsWith("/")) return BASE_URL + path;
  return `${BASE_URL}/${path}`;
}

/** Whether `url` points at showcamrips.com (with or without www). */
export function isShowcamripsUrl(url: string): boolean {
  try {
    return new URL(resolveUrl(url)).hostname.replace(/^www\./, "") === "showcamrips.com";
  } catch {
    return false;
  }
}

/**
 * The id of a video page, `/show-cam-sex-movies/2210361-cosmictandem-….html`,
 * or "" for any other URL. Ids grow with every upload and never change.
 */
export function videoIdFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/show-cam-sex-movies\/(\d+)-/);
  if (!match) return "";
  return match[1];
}

/** Whether `url` is a video page. */
export function isVideoUrl(url: string): boolean {
  return videoIdFromUrl(url) !== "";
}

/**
 * The id carried by a play-page URL (`play.php?idd=…`, `loading_video.php?idd=…`)
 * or by a video page URL; "" when there is none.
 */
export function playbackIdFromUrl(url: string): string {
  const match = resolveUrl(url).match(/[?&]idd=(\d+)/);
  if (match) return match[1];
  return videoIdFromUrl(url);
}

/** The play page that holds a video's player. It needs only the id. */
export function playUrl(videoId: string): string {
  return `${BASE_URL}/play.php?idd=${videoId}`;
}

/**
 * The thumbnail of a video. Listings inline it as a base64 image, but the
 * image host serves every video's still at a fixed path.
 */
export function posterUrl(videoId: string): string {
  return `https://www.showpicmoi.com/img/${videoId}.jpg`;
}

/** The external id listings and video pages share. */
export function externalId(videoId: string): string {
  return `showcamrips-${videoId}`;
}

/** The kind of entity a `/<kind>/en/<name>/` page lists. */
export type ListingKind = "model" | "site" | "cat";

/** A model, cam site or category page parsed from its URL. */
export interface ListingPage {
  kind: ListingKind;
  name: string;
}

/** The model/site/category a listing URL names, or undefined for other URLs. */
export function listingPageFromUrl(url: string): ListingPage | undefined {
  const match = resolveUrl(url).match(/^https:\/\/www\.showcamrips\.com\/(model|site|cat)\/en\/([^/?#]+)/);
  if (!match) return undefined;
  return { kind: match[1] as ListingKind, name: decodeURIComponent(match[2]) };
}

/** The first page of a model, cam site or category listing. */
export function listingPageUrl(page: ListingPage): string {
  return `${BASE_URL}/${page.kind}/en/${encodeURIComponent(page.name)}/`;
}

/**
 * The URL of page `page` (1-based) of a listing whose first page is `firstPageUrl`.
 * Later pages append `<n>-pg/`.
 */
export function pageUrl(firstPageUrl: string, page: number): string {
  if (page === 1) return firstPageUrl;
  return `${firstPageUrl}${page}-pg/`;
}

/** Seconds in "00:13:13" or "13:13"; 0 when unparseable. */
export function parseClockSeconds(text: string): number {
  const match = text.trim().match(/^(?:(\d+):)?(\d+):(\d{2})$/);
  if (!match) return 0;
  const hours = parseInt(match[1] || "0", 10);
  return hours * 3600 + parseInt(match[2], 10) * 60 + parseInt(match[3], 10);
}

/** "2026-10-28" from "2026.10.28_21.05.24" or "2026.10.28"; "" when there is no date. */
export function parseSiteDate(text: string): string {
  const match = text.match(/(\d{4})\.(\d{2})\.(\d{2})/);
  if (!match) return "";
  return `${match[1]}-${match[2]}-${match[3]}`;
}

/**
 * The recording time from a video URL's slug (`…-rip-20261028-210524.html`) as
 * the site writes it in titles, "2026.10.28_21.05.24"; "" without one.
 */
export function recordingStampFromUrl(url: string): string {
  const match = url.match(/-(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})\.html/);
  if (!match) return "";
  return `${match[1]}.${match[2]}.${match[3]}_${match[4]}.${match[5]}.${match[6]}`;
}

const EMPTY_PAGE_ATTEMPTS = 3;

/**
 * Fetches a page's HTML. The site now and then answers 200 with an empty
 * body (search above all), so an empty answer is asked for again; "" is
 * returned when every attempt came back empty.
 */
export async function fetchPage(url: string): Promise<string> {
  for (let attempt = 1; attempt <= EMPTY_PAGE_ATTEMPTS; attempt++) {
    const html = await fetchHtml(url);
    if (html.trim()) return html;
  }
  return "";
}
