import { parse, type HTMLElement } from "node-html-parser";
import {
  BROWSER_HEADERS,
  isSamePerformer,
  type Performer,
  type PerformerIdentifyArgs,
  type PerformerIdentifyResult,
  type ScrapeResult,
} from "@playingwithclouds/veil-sdk";
import { canonicalUrl, decodeEntities, fetchJson, ORIGINS, siteOf } from "./http.ts";

const PROFILE_PATH = /^\/(pornstars|models|profiles)\/([^/?#]+)/;
const PERFORMER_PATH = /^\/(pornstars|models)\/([^/?#]+)/;
const MAX_NAMES_LOOKED_UP = 3;

/** How identification reaches the site; tests replace it with saved answers. */
export interface ProfileFetcher {
  /** The site's search suggestions for a name, as the JSON document it answers with. */
  suggest: (name: string) => Promise<any>;
  /** A profile page's html and the URL it ended at after redirects. */
  page: (url: string) => Promise<{ html: string; url: string }>;
}

/** The live site: suggestions from its search box, pages as a browser fetches them. */
const liveFetcher: ProfileFetcher = {
  suggest: (name) => fetchJson(`${ORIGINS.xvideos}/search-suggest/${encodeURIComponent(name)}`, ORIGINS.xvideos),
  page: fetchProfilePage,
};

/** Fetches a profile page in English and says where redirects ended up. */
async function fetchProfilePage(url: string): Promise<{ html: string; url: string }> {
  const response = await fetch(url, {
    headers: { ...BROWSER_HEADERS, "Accept-Language": "en-US,en;q=0.9" },
    signal: AbortSignal.timeout(20_000),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  return { html: await response.text(), url: response.url };
}

/** Whether the URL is a profile page of xvideos (pornstar, model or user/channel profile). */
export function isProfileUrl(url: string): boolean {
  if (siteOf(url) !== "xvideos") return false;
  return PROFILE_PATH.test(new URL(url).pathname);
}

/** Serves performer:find: reads one xvideos profile page into a performer. */
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const target = canonicalUrl(url);
  const page = await liveFetcher.page(target);
  const performer = parseProfile(page.html, target);
  if (!performer) throw new Error(`xvideos: no profile on ${target}`);
  return [{ type: "performer", performer }];
}

/** Parses a profile page into a performer; undefined when the page has no profile heading. target is the canonical profile URL. */
export function parseProfile(html: string, target: string): Performer | undefined {
  const root = parse(html);
  const name = decodeEntities(root.querySelector("#profile-title h2 strong")?.text ?? "").trim();
  if (name === "") return undefined;
  const path = new URL(target).pathname;

  const performer: Performer = {
    type: "performer",
    external_id: `xvideos-${path.replace(/^\/+/, "").replace(/\//g, "-")}`,
    source_url: target,
    name,
  };
  const gender = genderOf(infoValue(root, "pinfo-sex"));
  if (gender) performer.gender = gender;
  const country = infoValue(root, "pinfo-country");
  if (country !== "") performer.country = country;
  const image = imageOf(root);
  if (image) {
    performer.image_path = image;
    performer.images = [{ type: "profile", file_path: image }];
  }
  return performer;
}

/** The text of one "About me" row (id like pinfo-country), empty when the row is missing. */
function infoValue(root: HTMLElement, rowId: string): string {
  return decodeEntities(root.querySelector(`#${rowId} span`)?.text ?? "").trim();
}

/** The site's gender wording as the library's; undefined for wording it doesn't know. */
function genderOf(label: string): string | undefined {
  const genders: Record<string, string> = {
    woman: "female",
    man: "male",
    "trans woman": "transgender_female",
    "trans man": "transgender_male",
  };
  return genders[label.toLowerCase()];
}

/** The profile picture; the player-less placeholder the page falls back to is not one. */
function imageOf(root: HTMLElement): string | undefined {
  const source = root.querySelector("#profile-title .profile-pic img")?.getAttribute("src");
  if (!source || !source.startsWith("http")) return undefined;
  return decodeEntities(source);
}

/** The /pornstars/<slug> profile URLs a suggestions answer lists for a name, only for entries named exactly as asked. */
export function pornstarUrlsFromSuggestions(answer: any, args: PerformerIdentifyArgs): string[] {
  const entries = answer?.data?.pornstar;
  if (!Array.isArray(entries)) return [];
  const urls: string[] = [];
  for (const entry of entries) {
    if (typeof entry?.N !== "string" || typeof entry?.F !== "string") continue;
    if (!isSamePerformer({ name: entry.N }, args)) continue;
    urls.push(`${ORIGINS.xvideos}/pornstars/${entry.F}`);
  }
  return urls;
}

/** The pornstar or model profile URLs among the library's known pages, which name this site's slug directly. */
function knownProfileUrls(args: PerformerIdentifyArgs): string[] {
  const urls: string[] = [];
  for (const url of args.urls ?? []) {
    if (siteOf(url) !== "xvideos") continue;
    if (!PERFORMER_PATH.test(new URL(url).pathname)) continue;
    urls.push(canonicalUrl(url));
  }
  return urls;
}

/** The URLs of this site's candidate profiles: known ones first, then what the search suggestions name. */
async function candidateUrls(args: PerformerIdentifyArgs, fetcher: ProfileFetcher): Promise<string[]> {
  const urls = knownProfileUrls(args);
  if (urls.length > 0) return urls;
  for (const name of [args.name, ...(args.aliases ?? [])].slice(0, MAX_NAMES_LOOKED_UP)) {
    const answer = await fetcher.suggest(name);
    for (const url of pornstarUrlsFromSuggestions(answer, args)) {
      if (!urls.includes(url)) urls.push(url);
    }
  }
  return urls;
}

/**
 * Serves performer:identify: finds the pornstar profile of the performer in the
 * site's own search suggestions (or a known profile URL) and returns it only
 * when isSamePerformer agrees. A slug that redirects away is a user account,
 * not the pornstar, and is refused.
 */
export async function identifyPerformer(
  args: PerformerIdentifyArgs,
  fetcher: ProfileFetcher = liveFetcher,
): Promise<PerformerIdentifyResult> {
  for (const url of await candidateUrls(args, fetcher)) {
    const page = await fetcher.page(url);
    if (new URL(page.url).pathname !== new URL(url).pathname) continue;
    const performer = parseProfile(page.html, url);
    if (performer && isSamePerformer(performer, args)) return { performer };
  }
  return {};
}
