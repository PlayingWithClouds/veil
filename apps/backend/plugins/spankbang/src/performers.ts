import { parse, HTMLElement } from "node-html-parser";
import { foldName, isSamePerformer } from "@playingwithclouds/veil-sdk";
import type {
  Image,
  Performer,
  PerformerIdentifyArgs,
  PerformerIdentifyResult,
  ScrapeResult,
} from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities, fetchPage, resolveUrl } from "./http.ts";

/** Loads a page's HTML; injected so the identify decision can be tested without the network. */
export type PageFetcher = (url: string) => Promise<string>;

// Profiles are /<id>/pornstar/<slug>/. The id picks the person; the slug is
// cosmetic (any slug under a valid id serves the same page).
const PROFILE_PATH = /\/([0-9a-z]+)\/pornstar\/([^/?#"']+)/i;

// How many suggested profiles identify will open before giving up.
const MAX_CANDIDATES = 3;

// How many of the library's names (name first, then aliases) identify searches by.
const MAX_SEARCH_TERMS = 3;

/** Whether the URL is a spankbang pornstar profile. */
export function isPerformerUrl(url: string): boolean {
  return PROFILE_PATH.test(url);
}

/** The stable per-profile id (e.g. "1l") of a profile URL, or "" for other URLs. */
export function performerIdFromUrl(url: string): string {
  const match = url.match(PROFILE_PATH);
  if (!match) return "";
  return match[1];
}

/** Handles `performer:find`: reads one profile page into a Performer. */
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const performer = parsePerformerPage(await fetchPage(target), target);
  if (!performer) return [];
  return [{ type: "performer", performer }];
}

/**
 * Parses a profile page. The stats (age, origin, ethnicity, hair, height,
 * weight) are `Label: <em>value</em>` spans; the page gives an age but no
 * birthdate and no gender, so neither is filled. Returns null without a name.
 */
export function parsePerformerPage(html: string, url: string): Performer | null {
  const root = parse(html);
  const name = decodeEntities(root.querySelector('[data-testid="profile-name"]')?.text ?? "").trim();
  if (name === "") return null;

  const sourceUrl = canonicalProfileUrl(root, url);
  const stats = extractStats(root);
  const performer: Performer = {
    type: "performer",
    external_id: `spankbang-pornstar-${performerIdFromUrl(sourceUrl)}`,
    source_url: sourceUrl,
    name,
  };
  fillStats(performer, stats);

  const bio = extractBio(root);
  if (bio !== "") performer.details = bio;

  const avatar = extractAvatar(root);
  if (avatar !== "") {
    performer.image_path = avatar;
    performer.images = [{ type: "profile", file_path: avatar } as Image];
  }
  return performer;
}

/** The page's canonical link when it is a profile URL, else the URL asked for. */
function canonicalProfileUrl(root: HTMLElement, requested: string): string {
  const canonical = root.querySelector('link[rel="canonical"]')?.getAttribute("href") ?? "";
  if (isPerformerUrl(canonical)) return resolveUrl(canonical);
  return requested;
}

/** The `Label: <em>value</em>` spans, keyed by lowercased label; the first of a label wins. */
function extractStats(root: HTMLElement): Map<string, string> {
  const stats = new Map<string, string>();
  for (const span of root.querySelectorAll("span")) {
    const value = span.querySelector("em")?.text.trim() ?? "";
    if (value === "") continue;
    const label = span.text.replace(value, "").replace(/:\s*$/, "").trim().toLowerCase();
    if (label === "" || stats.has(label)) continue;
    stats.set(label, decodeEntities(value));
  }
  return stats;
}

/** Copies the stats the Performer model has a field for. */
function fillStats(performer: Performer, stats: Map<string, string>): void {
  const country = stats.get("from");
  if (country) performer.country = country;
  const ethnicity = stats.get("ethnicity");
  if (ethnicity) performer.ethnicity = ethnicity;
  const hairColor = stats.get("hair color");
  if (hairColor) performer.hair_color = hairColor;
  const eyeColor = stats.get("eye color");
  if (eyeColor) performer.eye_color = eyeColor;
  const heightCm = measure(stats.get("height"), "cm");
  if (heightCm) performer.height_cm = heightCm;
  const weightKg = measure(stats.get("weight"), "kg");
  if (weightKg) performer.weight_kg = weightKg;
}

/** "164 cm (5 ft 4 in)" with unit "cm" → 164; undefined when the unit is absent. */
function measure(value: string | undefined, unit: string): number | undefined {
  if (!value) return undefined;
  const match = value.match(new RegExp(`(\\d+(?:\\.\\d+)?)\\s*${unit}`, "i"));
  if (!match) return undefined;
  return Math.round(parseFloat(match[1]));
}

/** The biography paragraph, with its line breaks kept between paragraphs. */
function extractBio(root: HTMLElement): string {
  const paragraph = root.querySelector("p[aria-expanded]");
  if (!paragraph) return "";
  return decodeEntities(paragraph.text)
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "")
    .join("\n");
}

/** The profile picture as an absolute URL. */
function extractAvatar(root: HTMLElement): string {
  const src = root.querySelector('img[data-testid="profile-detail-image"]')?.getAttribute("src") ?? "";
  if (src === "") return "";
  return resolveUrl(src);
}

/**
 * Handles `performer:identify`: opens this site's profile of the performer the
 * library describes and returns it only when isSamePerformer agrees.
 */
export async function performerIdentify(
  args: PerformerIdentifyArgs,
  fetchHtml: PageFetcher = fetchPage,
): Promise<PerformerIdentifyResult> {
  const urls = await candidateUrls(args, fetchHtml);
  for (const url of urls.slice(0, MAX_CANDIDATES)) {
    const performer = parsePerformerPage(await fetchHtml(url), url);
    if (performer && isSamePerformer(performer, args)) return { performer };
  }
  return {};
}

/** Profile URLs worth opening: ones the library already knows on this site, else the search suggestions. */
async function candidateUrls(args: PerformerIdentifyArgs, fetchHtml: PageFetcher): Promise<string[]> {
  const known = (args.urls ?? []).filter((url) => /^https?:\/\/([a-z]+\.)?spankbang\.com\//i.test(url) && isPerformerUrl(url));
  if (known.length > 0) return known.map(resolveUrl);
  return suggestedProfileUrls(args, fetchHtml);
}

/**
 * Profiles the site's search suggestions offer for the performer's names,
 * keeping only those whose slug is one of the names: suggestions also carry
 * unrelated trending profiles.
 */
async function suggestedProfileUrls(args: PerformerIdentifyArgs, fetchHtml: PageFetcher): Promise<string[]> {
  const names = [args.name, ...(args.aliases ?? [])].slice(0, MAX_SEARCH_TERMS);
  const wanted = new Set(names.map(foldName));
  const urls: string[] = [];
  for (const name of names) {
    const suggestions = await fetchHtml(suggestUrl(name));
    for (const url of profileLinks(suggestions)) {
      if (!wanted.has(foldName(slugName(url))) || urls.includes(url)) continue;
      urls.push(url);
    }
  }
  return urls;
}

/** The search-bar suggestion endpoint's URL for a term. */
function suggestUrl(term: string): string {
  const query = new URLSearchParams({ term, ver: "6", orientation: "Straight" });
  return `${BASE_URL}/api/search/search_bar_suggest?${query.toString()}`;
}

/**
 * Every profile link in a suggestion response. Read from the raw text: the
 * suggestions sit inside Alpine `<template>` tags that HTML parsers skip.
 */
export function profileLinks(html: string): string[] {
  const links: string[] = [];
  for (const match of html.matchAll(/href="(\/[0-9a-z]+\/pornstar\/[^"/]+\/?)"/gi)) {
    const url = resolveUrl(match[1]);
    if (!links.includes(url)) links.push(url);
  }
  return links;
}

/** The name a profile URL's slug spells ("riley+reid" → "riley reid"). */
function slugName(url: string): string {
  const match = url.match(PROFILE_PATH);
  if (!match) return "";
  return decodeURIComponent(match[2].replace(/\+/g, " "));
}
