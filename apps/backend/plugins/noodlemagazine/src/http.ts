import type { HTMLElement } from "node-html-parser";
import { CANONICAL_ORIGIN, fetchFromMirrors, isSitePage, mirrorPath, type MirrorPage } from "./mirrors.ts";

/** Fetches a site path (with its query) from the first mirror that serves it. */
export function fetchPage(path: string): Promise<MirrorPage> {
  return fetchFromMirrors(path);
}

/**
 * Fetches a video's watch page from the first mirror that shows the video:
 * its metadata, and with needPlayer its player playlist too (a copy may have
 * lost a video another still plays).
 */
export function fetchWatchPage(videoId: string, needPlayer: boolean): Promise<MirrorPage> {
  return fetchFromMirrors(watchPath(videoId), (body) => {
    if (!isSitePage(body) || !body.includes('property="og:title"')) return false;
    return !needPlayer || body.includes("window.playlist");
  });
}

/** The canonical watch page of a video id such as -152603790_456239053. */
export function watchUrl(videoId: string): string {
  return `${CANONICAL_ORIGIN}${watchPath(videoId)}`;
}

/** The path of a video's watch page, the same on every mirror. */
export function watchPath(videoId: string): string {
  return `/watch/${videoId}`;
}

/** The plugin's external id of a video, whichever mirror listed it. */
export function externalIdOf(videoId: string): string {
  return `noodlemagazine-${videoId}`;
}

/**
 * The video id in a watch link (absolute on any mirror, or relative), or an
 * empty string when the link is no watch page.
 */
export function videoIdFromUrl(url: string): string {
  let path = url;
  if (/^https?:\/\//.test(url)) {
    const onMirror = mirrorPath(url);
    if (onMirror === undefined) return "";
    path = onMirror;
  }
  const match = path.match(/^\/watch\/(-?\d+_\d+)\/?(?:[?#]|$)/);
  if (!match) return "";
  return match[1];
}

/** An element's attribute value, or an empty string when the attribute is missing. */
export function attributeOf(element: HTMLElement, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}

/** Collapses whitespace runs and trims. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

const NAMED_ENTITIES: Record<string, string> = {
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Decodes the HTML entities a page writes into attribute values and text. */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => {
    if (name.startsWith("#x") || name.startsWith("#X")) return String.fromCodePoint(parseInt(name.slice(2), 16));
    if (name.startsWith("#")) return String.fromCodePoint(parseInt(name.slice(1), 10));
    const decoded = NAMED_ENTITIES[name.toLowerCase()];
    if (decoded === undefined) return match;
    return decoded;
  }).replace(/&amp;/g, "&");
}
