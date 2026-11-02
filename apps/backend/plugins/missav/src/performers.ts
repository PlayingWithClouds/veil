import { parse, HTMLElement } from "node-html-parser";
import { foldName, isSamePerformer } from "@playingwithclouds/veil-sdk";
import type {
  Image,
  Performer,
  PerformerIdentifyArgs,
  PerformerIdentifyResult,
  ScrapeResult,
} from "@playingwithclouds/veil-sdk";
import { BASE_URL, canonicalUrl, decodeEntities, fetchPage, resolveUrl } from "./http.ts";

/** Loads a page's HTML; injected so the identify decision can be tested without the network. */
export type PageFetcher = (url: string) => Promise<string>;

// Actress pages are /<lang>/actresses/<name>, optionally behind a /dm<n> mirror
// prefix. The name may be spelled in Latin letters or in Japanese; both serve
// the same page.
const ACTRESS_PATH = /\/actresses\/([^/?#]+)/i;

// How many of the library's names (name first, then aliases) identify tries as page names.
const MAX_NAMES = 4;

/** Whether the URL is an actress profile (and not the /actresses index, ranking or saved lists). */
export function isActressUrl(url: string): boolean {
  const slug = slugFromUrl(url);
  return slug !== "" && slug !== "ranking";
}

/** The decoded name segment of an actress URL ("Yua%20Mikami" → "Yua Mikami"), or "". */
export function slugFromUrl(url: string): string {
  const match = url.match(ACTRESS_PATH);
  if (!match) return "";
  try {
    return decodeURIComponent(match[1]);
  } catch {
    return match[1];
  }
}

/** Handles `performer:find`: reads one actress page into a Performer. */
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  const performer = parsePerformerPage(await fetchPage(target), target);
  if (!performer) return [];
  return [{ type: "performer", performer }];
}

/**
 * Parses an actress page. The hero block shows the name, a portrait and up to
 * two lines: `159cm / 33F - 23 - 35` and `1993-08-16 (age 33)`; each part is
 * read when present. The page carries no country or gender, so neither is
 * filled. Returns null for names the site has no actress for, which it answers
 * with an empty 200 page.
 */
export function parsePerformerPage(html: string, url: string): Performer | null {
  const root = parse(html);
  const name = decodeEntities(root.querySelector(".hero-pattern h4")?.text ?? "").trim();
  if (name === "") return null;

  const performer: Performer = {
    type: "performer",
    external_id: `missav-actress-${name}`,
    source_url: actressUrl(name),
    name,
  };
  const aliases = extractAliases(root, name);
  if (aliases.length > 0) performer.aliases = aliases;

  fillStats(performer, heroLines(root).join(" "));

  const portrait = root.querySelector(".hero-pattern img")?.getAttribute("src") ?? "";
  if (portrait !== "") {
    performer.image_path = resolveUrl(portrait);
    performer.images = [{ type: "profile", file_path: performer.image_path } as Image];
  }
  return performer;
}

/**
 * The page URL credits link for an actress, the identity shared with
 * ScenePerformer.source_url. Built from the page's own name so a Japanese
 * spelling asked for still yields the Latin-letter URL.
 */
function actressUrl(name: string): string {
  return `${BASE_URL}/en/actresses/${encodeURIComponent(name)}`;
}

/** The other spellings of the name the page itself gives: its og:title and canonical URL (Japanese for Japanese actresses). */
function extractAliases(root: HTMLElement, name: string): string[] {
  const aliases: string[] = [];
  const title = root.querySelector('meta[property="og:title"]')?.getAttribute("content") ?? "";
  const titled = decodeEntities(title).match(/^Watch (.+)'s AV Online$/);
  const canonical = root.querySelector('link[rel="canonical"]')?.getAttribute("href") ?? "";
  for (const candidate of [titled?.[1] ?? "", slugFromUrl(canonical)]) {
    const spelling = candidate.trim();
    if (spelling === "" || foldName(spelling) === foldName(name)) continue;
    if (aliases.some((alias) => foldName(alias) === foldName(spelling))) continue;
    aliases.push(spelling);
  }
  return aliases;
}

/** The whitespace-collapsed text of each stats line under the name. */
function heroLines(root: HTMLElement): string[] {
  return root
    .querySelectorAll(".hero-pattern p")
    .map((line) => line.text.replace(/\s+/g, " ").trim())
    .filter((line) => line !== "");
}

/** Reads height, measurements and birthdate out of the stats text, whichever are shown. */
function fillStats(performer: Performer, text: string): void {
  const height = text.match(/(\d{2,3})\s*cm/i);
  if (height) performer.height_cm = parseInt(height[1], 10);
  const measurements = text.match(/(\d{2,3}[A-Z]{0,2})\s*-\s*(\d{2,3})\s*-\s*(\d{2,3})/);
  if (measurements) performer.measurements = `${measurements[1]}-${measurements[2]}-${measurements[3]}`;
  const birthdate = text.match(/(\d{4}-\d{2}-\d{2})/);
  if (birthdate) performer.birthdate = birthdate[1];
}

/**
 * Handles `performer:identify`: opens the actress page for each name the
 * library knows (or a known missav profile URL) and returns the first that
 * isSamePerformer accepts.
 */
export async function performerIdentify(
  args: PerformerIdentifyArgs,
  fetchHtml: PageFetcher = fetchPage,
): Promise<PerformerIdentifyResult> {
  for (const url of candidateUrls(args)) {
    const performer = parsePerformerPage(await fetchHtml(url), url);
    if (performer && isSamePerformer(performer, args)) return { performer };
  }
  return {};
}

/** Pages worth opening: profiles the library already links on this site, then one per known name. */
export function candidateUrls(args: PerformerIdentifyArgs): string[] {
  const urls: string[] = [];
  for (const url of args.urls ?? []) {
    if (isMissavActressUrl(url)) urls.push(canonicalUrl(url));
  }
  for (const name of [args.name, ...(args.aliases ?? [])].slice(0, MAX_NAMES)) {
    const url = actressUrl(name.trim());
    if (name.trim() !== "" && !urls.includes(url)) urls.push(url);
  }
  return urls;
}

/** Whether a URL is an actress profile on missav. */
function isMissavActressUrl(url: string): boolean {
  return /^https?:\/\/(www\.)?missav\.[a-z]+\//i.test(url) && isActressUrl(url);
}
