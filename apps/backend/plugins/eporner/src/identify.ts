import { parse } from "node-html-parser";
import { foldName, isSamePerformer } from "@playingwithclouds/veil-sdk";
import type { Performer, PerformerIdentifyArgs, PerformerIdentifyResult } from "@playingwithclouds/veil-sdk";
import { SITE, UA } from "./http.ts";
import { performerFromPage } from "./performers.ts";
import { fetchPornstarByUrl } from "./pornstar.ts";

const SUGGEST_TIMEOUT_MS = 8_000;
const MAX_CANDIDATES = 3;

/** One pornstar the search box offers: the profile URL (with the site's id suffix) and the name shown. */
export interface PornstarSuggestion {
  url: string;
  name: string;
}

/** Reads the pornstar rows of the search box fragment; keyword rows and repeated profiles are skipped. */
export function pornstarsFromSuggestHtml(html: string): PornstarSuggestion[] {
  const found: PornstarSuggestion[] = [];
  const seen = new Set<string>();
  for (const anchor of parse(html).querySelectorAll(".qs-results a[title]")) {
    const href = anchor.getAttribute("href") ?? "";
    if (!href.startsWith("/pornstar/") || seen.has(href)) continue;
    seen.add(href);
    found.push({ url: `${SITE}${href}`, name: (anchor.getAttribute("title") ?? "").trim() });
  }
  return found;
}

/** Profile URLs of this site already known to the library. */
function knownProfileUrls(args: PerformerIdentifyArgs): string[] {
  const urls: string[] = [];
  for (const url of args.urls ?? []) {
    if (!/^https?:\/\/(?:www\.)?eporner\.com\/pornstar\/[^/?#]+/.test(url)) continue;
    urls.push(url.split(/[?#]/)[0]);
  }
  return urls;
}

/** Asks the site's search box for pornstars matching a name. */
async function suggestedPornstars(name: string): Promise<PornstarSuggestion[]> {
  const response = await fetch(`${SITE}/suggest/${encodeURIComponent(name)}/all/?layout=sections`, {
    headers: { "User-Agent": UA },
    signal: AbortSignal.timeout(SUGGEST_TIMEOUT_MS),
  });
  if (!response.ok) return [];
  return pornstarsFromSuggestHtml(await response.text());
}

/**
 * The profile at a URL filled as performer:find fills it, or undefined when the
 * page is gone or shows no name (performer:find would invent one from the slug).
 */
async function profileAt(url: string): Promise<Performer | undefined> {
  const partial = await fetchPornstarByUrl(url);
  if (partial === null || !partial.name) return undefined;
  return performerFromPage(url, partial);
}


/** Box suggestions whose shown name is one the library knows the performer by. */
async function suggestedUrls(args: PerformerIdentifyArgs): Promise<string[]> {
  const wanted = new Set([args.name, ...(args.aliases ?? [])].map(foldName));
  const suggestions = await suggestedPornstars(args.name);
  return suggestions.filter((suggestion) => wanted.has(foldName(suggestion.name))).map((suggestion) => suggestion.url);
}

/** The first profile among the URLs that the identity rule confirms. */
async function firstConfirmed(urls: string[], args: PerformerIdentifyArgs): Promise<Performer | undefined> {
  for (const url of urls.slice(0, MAX_CANDIDATES)) {
    const performer = await profileAt(url);
    if (performer !== undefined && isSamePerformer(performer, args)) return performer;
  }
  return undefined;
}

/**
 * Finds this site's profile of the performer: a known profile URL first, then
 * the search box's pornstar rows. Returns it only when the rule confirms the person.
 */
export async function identifyPerformer(args: PerformerIdentifyArgs): Promise<PerformerIdentifyResult> {
  const known = await firstConfirmed(knownProfileUrls(args), args);
  if (known !== undefined) return { performer: known };
  const suggested = await firstConfirmed(await suggestedUrls(args), args);
  if (suggested !== undefined) return { performer: suggested };
  return {};
}
