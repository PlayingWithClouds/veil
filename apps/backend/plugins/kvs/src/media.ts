import type { HTMLElement } from "node-html-parser";
import type { SiteConfig } from "./sites.ts";
import { decodeVideoUrl } from "./decode.ts";
import { absoluteUrl } from "./urls.ts";
import { cleanText, decodeEntities } from "./text.ts";

// A KVS video page names its files in up to three places, depending on the
// theme: kt_player's flashvars (video_url, video_alt_url, video_alt_url2…,
// each with a *_text quality label, possibly obfuscated), <source> tags of an
// HTML5 player, or plain /get_file/ links in scripts and download buttons.

/** One playable file of a video. */
export interface MediaSource {
  url: string;
  /** e.g. "720p"; "" when the page doesn't say */
  quality: string;
  /** vertical resolution for ranking; 0 when unknown */
  height: number;
}

/**
 * KVS serves media from /get_file/ (or /media/ on some themes): a server
 * group, then a 32+ character hex hash — or, with path encryption on, one
 * opaque base64url token standing for the whole path.
 */
const MEDIA_PATH = /\/(?:get_file|media)\/\d+\/[A-Za-z0-9_-]{32,}(?:\.mp4)?(?:\/|\?|$)/;

/** Quoted media URLs anywhere in the page source. */
const MEDIA_URL_IN_SOURCE = /https?:\/\/[^\s"'<>\\]+?\/(?:get_file|media)\/\d+\/[0-9a-f]{32,}\/[^\s"'<>\\]*?\.mp4\/?[^\s"'<>\\]*/gi;

/** Script objects keyed by resolution: `{ 360 : 'https://…/get_file/…', 720 : '…' }`. */
const HEIGHT_KEYED_URL = /\b(\d{3,4})\s*:\s*['"](https?:\/\/[^'"\s]+?\/(?:get_file|media)\/\d+\/[0-9a-f]{32,}\/[^'"\s]+)['"]/gi;

/** Files that are previews, not the video. */
const NOT_THE_VIDEO = /preview|vthumb|trailer|_tr\.mp4|screens/i;

/** Quality labels naming no resolution (ok.xxx's "Auto" duplicates another source). */
const AUTO_LABEL = /^auto$/i;

/** Whether `url` is a KVS media URL rather than a page. */
export function isMediaUrl(url: string): boolean {
  return MEDIA_PATH.test(url);
}

/**
 * The flashvars object of kt_player as name → value, empty when the page has
 * none. Some sites rename the variable (`var t86b401a0ea = {…}`), so the
 * object is found by its `video_id`/`license_code` entries instead.
 */
export function parseFlashvars(html: string): Record<string, string> {
  const flashvars: Record<string, string> = {};
  const body = flashvarsBody(html);
  for (const pair of body.matchAll(/([A-Za-z_][\w]*)\s*:\s*'((?:[^'\\]|\\.)*)'/g)) {
    flashvars[pair[1]] = pair[2].replace(/\\(.)/g, "$1");
  }
  return flashvars;
}

/** The source text of the object literal holding the player's video_id/license_code, or "". */
function flashvarsBody(html: string): string {
  let marker = html.search(/\blicense_code\s*:\s*'/);
  if (marker === -1) marker = html.search(/\bvideo_url\s*:\s*'/);
  if (marker === -1) return "";

  // The object is assigned ("= {"); values may hold braces ("{time}.jpg"), so
  // the nearest assignment whose object spans the marker is the one.
  const opens = [...html.slice(0, marker).matchAll(/=\s*\{/g)].map((match) => (match.index ?? 0) + match[0].length - 1);
  for (let index = opens.length - 1; index >= 0; index--) {
    const close = matchingBrace(html, opens[index]);
    if (close > marker) return html.slice(opens[index] + 1, close);
  }
  return "";
}

/** The index of the "}" closing the "{" at `open`, skipping quoted strings; -1 if unclosed. */
function matchingBrace(source: string, open: number): number {
  let depth = 0;
  let quote = "";
  for (let index = open; index < source.length; index++) {
    const character = source.charAt(index);
    if (quote) {
      if (character === "\\") index++;
      else if (character === quote) quote = "";
      continue;
    }
    if (character === "'" || character === '"') quote = character;
    else if (character === "{") depth++;
    else if (character === "}") {
      depth--;
      if (depth === 0) return index;
    }
  }
  return -1;
}

/**
 * Every distinct playable file of the video on `root`/`html`, best quality
 * first. Unlabelled files are dropped when labelled ones exist, as they
 * duplicate one of them.
 */
export function extractSources(site: SiteConfig, root: HTMLElement, html: string): MediaSource[] {
  const sources: MediaSource[] = [];
  const seen = new Set<string>();
  const add = (url: string, label: string) => {
    const source = mediaSource(site, url, label);
    if (!source) return;
    const identity = source.url.split("?")[0];
    if (seen.has(identity)) return;
    seen.add(identity);
    sources.push(source);
  };

  for (const [url, label] of flashvarSources(parseFlashvars(html))) add(url, label);
  for (const element of root.querySelectorAll("source[src]")) {
    const label = element.getAttribute("title") ?? element.getAttribute("label") ?? "";
    if (AUTO_LABEL.test(label.trim())) continue;
    add(element.getAttribute("src") ?? "", label);
  }
  for (const match of html.matchAll(HEIGHT_KEYED_URL)) add(match[2], `${match[1]}p`);
  for (const match of html.matchAll(MEDIA_URL_IN_SOURCE)) add(match[0], "");

  return bestPerQuality(sources);
}

/**
 * One source per quality, best first. Sources of unknown resolution are
 * dropped when others name theirs, and unlabelled ones when others carry a
 * label, as they duplicate one of them.
 */
function bestPerQuality(sources: MediaSource[]): MediaSource[] {
  let usable = sources.filter((source) => source.height > 0);
  if (usable.length === 0) usable = sources.filter((source) => source.quality);
  if (usable.length === 0) usable = sources;
  const unique: MediaSource[] = [];
  for (const source of usable) {
    if (unique.some((known) => known.quality === source.quality)) continue;
    unique.push(source);
  }
  return unique.sort((first, second) => second.height - first.height);
}

/** The (url, label) pairs of kt_player's video_url / video_alt_url* flashvars, decoded. */
function flashvarSources(flashvars: Record<string, string>): [string, string][] {
  const pairs: [string, string][] = [];
  const license = flashvars.license_code ?? "";
  for (const name of Object.keys(flashvars)) {
    if (!/^video_(?:alt_)?url\d*$/.test(name)) continue;
    const value = flashvars[name];
    if (!value) continue;
    try {
      pairs.push([decodeVideoUrl(value, license), flashvars[`${name}_text`] ?? ""]);
    } catch {
      // An undecodable entry is skipped; other sources may still play.
    }
  }
  return pairs;
}

/** A media source from a raw URL and label, or undefined when it's no KVS video file. */
function mediaSource(site: SiteConfig, rawUrl: string, label: string): MediaSource | undefined {
  const url = absoluteUrl(site, decodeEntities(rawUrl.trim()));
  if (!url || !isMediaUrl(url)) return undefined;
  if (NOT_THE_VIDEO.test(fileName(url))) return undefined;

  const cleanLabel = cleanText(label);
  const height = heightOf(cleanLabel) || heightOf(fileName(url));
  // Resolutions are named uniformly ("HD" → "720p") so downloads read alike across sites.
  let quality = cleanLabel;
  if (height > 0) quality = `${height}p`;
  return { url, quality, height };
}

/** The file name segment of a media URL ("123_720p.mp4"). */
function fileName(url: string): string {
  const path = url.split("?")[0].replace(/\/$/, "");
  return path.slice(path.lastIndexOf("/") + 1);
}

/**
 * The vertical resolution a label or file name names: "720p", "1080",
 * "_480m.mp4", KVS's format names ("SD", "HD", "FHD", "hd.mp4"); 0 when none.
 */
function heightOf(text: string): number {
  const resolution = text.match(/(?:^|[^\d])(2160|1440|1080|720|540|480|360|240)(?:p|m|[^\d]|$)/i);
  if (resolution) return parseInt(resolution[1], 10);
  if (/^(?:4k|uhd)$/i.test(text)) return 2160;
  if (/^(?:fhd|full ?hd)$/i.test(text)) return 1080;
  if (/(?:^|\d|_)hd\.mp4$|^hd$/i.test(text)) return 720;
  if (/^sd$/i.test(text)) return 480;
  return 0;
}
