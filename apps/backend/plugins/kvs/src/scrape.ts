import { parse, type HTMLElement } from "node-html-parser";
import type {
  DiscoveredItem,
  Download,
  ScenePerformer,
  Scene,
  ScrapeResult,
  Studio,
  StudioRef,
} from "@playingwithclouds/veil-sdk";
import { siteForUrl, type SiteConfig } from "./sites.ts";
import { fetchPage } from "./http.ts";
import { parseVideoCards, imageSource } from "./listing.ts";
import { extractSources, parseFlashvars, type MediaSource } from "./media.ts";
import { jsonLdList, jsonLdNames, jsonLdString, metaContent, metaContents, videoObject, type VideoObject } from "./page.ts";
import { cleanText, entityName, isoDate, parseDurationText, parseIsoDuration } from "./text.ts";
import {
  absoluteUrl,
  canonicalUrl,
  entityExternalId,
  entitySlug,
  entityUrl,
  videoExternalId,
  videoKey,
} from "./urls.ts";

const MAX_RELATED = 40;

/** What a video page's credit block links: models, tags and its channel. */
interface Credits {
  performers: ScenePerformer[];
  tags: string[];
  /** the most specific channel (e.g. a paysite) */
  channel?: Studio;
}

/** Everything read off one video page. */
interface VideoPage {
  site: SiteConfig;
  key: string;
  pageUrl: string;
  html: string;
  root: HTMLElement;
  flashvars: Record<string, string>;
  videoObject: VideoObject;
  /** the credit blocks (every match of the site's infoSelector) */
  info: HTMLElement[];
}

/**
 * Scrapes a KVS video page into one Scene, followed by its channel as a
 * Studio when the page credits one.
 */
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const page = await loadVideoPage(url);
  const credits = readCredits(page);
  const title = videoTitle(page);
  if (!title) throw new Error(`kvs: no title on ${url}`);

  const scene: Scene = {
    type: "scene",
    external_id: videoExternalId(page.site, page.key),
    source_url: page.pageUrl,
    title,
    studio: siteStudio(page.site),
    performers: credits.performers,
    tags: credits.tags,
    downloads: downloads(page, extractSources(page.site, page.root, page.html)),
    related: relatedItems(page),
  };
  setOptionalFields(scene, page);

  if (!credits.channel) {
    const named = namedStudio(page);
    if (named) scene.studio = named;
    return [{ type: "scene", scene }];
  }
  scene.studio = {
    name: credits.channel.name,
    external_id: credits.channel.external_id,
    source_url: credits.channel.source_url,
  };
  return [
    { type: "scene", scene },
    { type: "studio", studio: credits.channel },
  ];
}

/** Fetches and parses a video page; throws for URLs that are no supported KVS video. */
export async function loadVideoPage(url: string): Promise<VideoPage> {
  const site = siteForUrl(url);
  if (!site) throw new Error(`kvs: unsupported site ${url}`);
  const key = videoKey(site, url);
  if (!key) throw new Error(`kvs: not a video page: ${url}`);

  const pageUrl = canonicalUrl(site, url);
  const html = await fetchPage(site, pageUrl);
  const root = parse(html);
  return {
    site,
    key,
    pageUrl,
    root,
    html,
    flashvars: parseFlashvars(html),
    videoObject: videoObject(root),
    info: root.querySelectorAll(site.infoSelector),
  };
}

/** Description, date, duration and poster, each from the first source that has it. */
function setOptionalFields(scene: Scene, page: VideoPage): void {
  const details = videoDetails(page, scene.title);
  if (details) scene.details = details;
  const date = videoDate(page);
  if (date) scene.date = date;
  const duration = videoDuration(page);
  if (duration > 0) scene.duration_seconds = duration;
  const poster = videoPoster(page);
  if (poster) {
    scene.poster_path = poster;
    scene.images = [{ type: "poster", file_path: poster }];
  }
}

/**
 * The title from OpenGraph, JSON-LD, kt_player's video_title, the heading or
 * the document title, without a trailing " | Site" — the first that isn't
 * cut short with "...", else the first found.
 */
export function videoTitle(page: VideoPage): string {
  const candidates = [
    metaContent(page.root, "og:title"),
    jsonLdString(page.videoObject.name),
    cleanText(page.flashvars.video_title ?? ""),
    cleanText(page.root.querySelector("h1")?.text ?? ""),
    cleanText(page.root.querySelector("title")?.text ?? ""),
  ]
    .map((candidate) => withoutSiteSuffix(candidate, page.site))
    .filter((candidate) => candidate);
  const complete = candidates.find((candidate) => !/(\.\.\.|…)$/.test(candidate));
  if (complete) return complete;
  return candidates[0] ?? "";
}

/**
 * `title` without a trailing " - PornGO.com" / " | Any Porn" naming the site
 * or an SEO tail like " watch online or download".
 */
export function withoutSiteSuffix(title: string, site: SiteConfig): string {
  return withoutSiteName(title, site).replace(/\s+watch online(?: or download)?$/i, "");
}

/** `title` without a trailing " - <site name>". */
function withoutSiteName(title: string, site: SiteConfig): string {
  const match = title.match(/^(.*\S)\s+[-|/–—]\s+([^-|/–—]+)$/);
  if (!match) return title;
  const suffix = squash(match[2]);
  if (suffix.length < 3) return title;
  const siteNames = [squash(site.name), squash(new URL(site.origin).hostname.replace(/^www\./, ""))];
  if (siteNames.some((name) => suffix === name || suffix.startsWith(name) || name.startsWith(suffix))) return match[1];
  return title;
}

/** Lower-cased letters and digits only, for loose name comparison. */
function squash(text: string): string {
  return text.toLowerCase().replace(/[^a-z0-9]/g, "");
}

/** The description, unless it merely repeats the title or is the site's SEO boilerplate. */
function videoDetails(page: VideoPage, title: string): string {
  const candidates = [metaContent(page.root, "og:description"), jsonLdString(page.videoObject.description)];
  for (const candidate of candidates) {
    if (candidate && candidate !== title && !isBoilerplate(candidate, page.site)) return candidate;
  }
  return "";
}

/** Template descriptions: "Watch <title> for free…", "Porn video: …", or ones naming the site. */
function isBoilerplate(description: string, site: SiteConfig): boolean {
  if (/^(?:watch\b|porn video\b|free porn\b|welcome to\b)/i.test(description)) return true;
  return squash(description).includes(squash(site.name));
}

/** The upload date as YYYY-MM-DD, or undefined. */
function videoDate(page: VideoPage): string | undefined {
  const candidates = [
    jsonLdString(page.videoObject.uploadDate),
    metaContent(page.root, "video:release_date"),
    metaContent(page.root, "uploadDate"),
  ];
  for (const candidate of candidates) {
    const date = isoDate(candidate);
    if (date) return date;
  }
  return undefined;
}

/**
 * The runtime in seconds from JSON-LD, video meta tags, kt_player's
 * flashvars, the site's duration element or the credit block's text; 0 when
 * unknown.
 */
function videoDuration(page: VideoPage): number {
  const isoDuration = parseIsoDuration(jsonLdString(page.videoObject.duration));
  if (isoDuration > 0) return isoDuration;
  const secondsCandidates = [
    metaContent(page.root, "og:video:duration"),
    metaContent(page.root, "video:duration"),
    page.flashvars.duration ?? "",
  ];
  for (const candidate of secondsCandidates) {
    const seconds = parseInt(candidate, 10);
    if (seconds > 0) return seconds;
  }
  const itemprop = parseIsoDuration(metaContent(page.root, "duration"));
  if (itemprop > 0) return itemprop;
  if (page.site.durationSelector) {
    const badge = parseDurationText(page.root.querySelector(page.site.durationSelector)?.text ?? "");
    if (badge > 0) return badge;
  }
  for (const block of page.info) {
    const seconds = parseDurationText(block.text);
    if (seconds > 0) return seconds;
  }
  return 0;
}

/** The poster from OpenGraph, JSON-LD, kt_player's preview_url or the player's poster. */
function videoPoster(page: VideoPage): string {
  const candidates = [
    metaContent(page.root, "og:image"),
    jsonLdString(page.videoObject.thumbnailUrl),
    page.flashvars.preview_url ?? "",
    page.root.querySelector("video[poster]")?.getAttribute("poster") ?? "",
  ];
  for (const candidate of candidates) {
    if (candidate) return absoluteUrl(page.site, candidate);
  }
  return "";
}

/**
 * Models, tags and channel linked from the page's credit blocks; models and
 * tags fall back to the page's metadata when no block links any.
 */
function readCredits(page: VideoPage): Credits {
  const anchors = creditAnchors(page);
  const credits: Credits = {
    performers: linkedPerformers(page.site, anchors),
    tags: linkedTags(page.site, anchors),
    channel: linkedChannel(page.site, anchors),
  };
  if (credits.performers.length === 0) credits.performers = fallbackPerformers(page);
  if (credits.tags.length === 0) credits.tags = fallbackTags(page);
  return credits;
}

/** The links in the credit blocks, each once even when blocks nest. */
function creditAnchors(page: VideoPage): HTMLElement[] {
  const anchors = new Set<HTMLElement>();
  for (const block of page.info) {
    for (const anchor of block.querySelectorAll("a[href]")) anchors.add(anchor);
  }
  return [...anchors];
}

/** The model links, de-duplicated by page. */
function linkedPerformers(site: SiteConfig, anchors: HTMLElement[]): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const prefix of site.modelPrefixes) {
    for (const anchor of anchors) {
      const slug = entitySlug(site, anchor.getAttribute("href") ?? "", [prefix]);
      const name = entityName(anchor);
      if (!slug || !name || seen.has(slug)) continue;
      seen.add(slug);
      performers.push({
        name,
        order: performers.length,
        external_id: entityExternalId(site, "model", slug),
        source_url: entityUrl(site, prefix, slug),
      });
    }
  }
  return performers;
}

/** The tag/category link names, de-duplicated case-insensitively. */
function linkedTags(site: SiteConfig, anchors: HTMLElement[]): string[] {
  const names: string[] = [];
  for (const anchor of anchors) {
    if (!entitySlug(site, anchor.getAttribute("href") ?? "", site.tagPrefixes)) continue;
    names.push(entityName(anchor));
  }
  return uniqueNames(names);
}

/**
 * The most specific channel linked (prefixes are ordered most specific
 * first), parented to the next broader one or else to the site.
 */
function linkedChannel(site: SiteConfig, anchors: HTMLElement[]): Studio | undefined {
  const channels: Studio[] = [];
  for (const prefix of site.channelPrefixes) {
    const channel = channelUnder(site, anchors, prefix);
    if (channel && !channels.some((known) => known.name === channel.name)) channels.push(channel);
  }
  if (channels.length === 0) return undefined;
  const studio = channels[0];
  studio.parent = site.name;
  if (channels.length > 1) studio.parent = channels[1].name;
  return studio;
}

/** The first channel linked under `prefix`, or undefined. */
function channelUnder(site: SiteConfig, anchors: HTMLElement[], prefix: string): Studio | undefined {
  for (const anchor of anchors) {
    const slug = entitySlug(site, anchor.getAttribute("href") ?? "", [prefix]);
    const name = entityName(anchor);
    if (!slug || !name) continue;
    const studio: Studio = {
      type: "studio",
      external_id: entityExternalId(site, "channel", slug),
      source_url: entityUrl(site, prefix, slug),
      name,
    };
    const logo = anchor.querySelector("img");
    if (logo) {
      const source = imageSource(site, logo);
      if (source) studio.image_path = source;
    }
    return studio;
  }
  return undefined;
}

/** Performers named by JSON-LD actors when the credit block links none. */
function fallbackPerformers(page: VideoPage): ScenePerformer[] {
  return uniqueNames(jsonLdNames(page.videoObject.actor)).map((name, order) => ({ name, order }));
}

/** Tags from JSON-LD, kt_player's flashvars or video:tag meta when the credit block links none. */
function fallbackTags(page: VideoPage): string[] {
  const fromJsonLd = [...jsonLdList(page.videoObject.genre), ...jsonLdList(page.videoObject.keywords)];
  if (fromJsonLd.length > 0) return uniqueNames(fromJsonLd);
  const fromFlashvars = jsonLdList([page.flashvars.video_categories ?? "", page.flashvars.video_tags ?? ""]);
  if (fromFlashvars.length > 0) return uniqueNames(fromFlashvars);
  return uniqueNames(metaContents(page.root, "video:tag"));
}

/** A studio named only by JSON-LD (production company, or a non-person author). */
function namedStudio(page: VideoPage): StudioRef | undefined {
  const names = [
    ...jsonLdNames(page.videoObject.productionCompany),
    ...jsonLdNames(page.videoObject.author, "Person"),
  ];
  if (names.length === 0) return undefined;
  return { name: names[0] };
}

/** The site itself, credited when a video names no channel. */
export function siteStudio(site: SiteConfig): StudioRef {
  return { name: site.name, source_url: site.origin };
}

/**
 * The scene's downloads. Each points back at the page with the quality as
 * fragment ("…/video/1/#720p"), because /get_file/ links are signed per
 * request; resolve re-reads the page and follows that quality's link. Sites
 * serving one adaptive HLS stream get a single download.
 */
function downloads(page: VideoPage, sources: MediaSource[]): Download[] {
  if (sources.length === 0) return [];
  if (page.site.adaptiveStreams) {
    return [{ label: "auto", url: page.pageUrl, format: "hls" }];
  }
  if (sources.length === 1) {
    const download: Download = { label: "mp4", url: page.pageUrl, format: "mp4" };
    if (sources[0].quality) {
      download.label = sources[0].quality;
      download.quality = sources[0].quality;
    }
    return [download];
  }
  return sources.map((source) => ({
    label: source.quality,
    url: `${page.pageUrl}#${encodeURIComponent(source.quality)}`,
    quality: source.quality,
    format: "mp4",
  }));
}

/**
 * The page's related videos, parsed like a listing from the first related
 * block that links videos, without the video itself.
 */
function relatedItems(page: VideoPage): DiscoveredItem[] {
  const selfId = videoExternalId(page.site, page.key);
  for (const container of page.root.querySelectorAll(page.site.relatedSelector)) {
    const items = parseVideoCards(page.site, container).filter((item) => item.external_id !== selfId);
    if (items.length > 0) return items.slice(0, MAX_RELATED);
  }
  return [];
}

/** Names with blanks and case-insensitive duplicates removed, first spelling kept. */
function uniqueNames(names: string[]): string[] {
  const unique: string[] = [];
  const seen = new Set<string>();
  for (const name of names) {
    const normalized = name.toLowerCase();
    if (!name || seen.has(normalized)) continue;
    seen.add(normalized);
    unique.push(name);
  }
  return unique;
}
