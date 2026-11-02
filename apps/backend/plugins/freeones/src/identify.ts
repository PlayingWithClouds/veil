import { parse } from "node-html-parser";
import { foldName, isSamePerformer } from "@playingwithclouds/veil-sdk";
import type { Performer, PerformerIdentifyArgs, PerformerIdentifyResult } from "@playingwithclouds/veil-sdk";
import { performerFromBio } from "./bio.ts";
import { fetchPage, BASE_URL, slugFromUrl } from "./http.ts";
import { listingUrl, performerItemsFromPage } from "./listing.ts";

const MAX_CANDIDATES = 4;
const SEARCH_PAGE_SIZE = 10;

/** The slug a name gets on the site: lower case words joined by hyphens ("Sara Jay" → sara-jay). */
export function slugFromName(name: string): string {
  return foldName(name).replace(/ /g, "-");
}

/** Slugs of freeones profile URLs the library already knows. */
function knownSlugs(args: PerformerIdentifyArgs): string[] {
  const slugs: string[] = [];
  for (const url of args.urls ?? []) {
    if (!/^https?:\/\/(?:www\.)?freeones\.com\//.test(url)) continue;
    const slug = slugFromUrl(url);
    if (slug !== "") slugs.push(slug);
  }
  return slugs;
}

/** Slugs of the search results whose card name is one the library knows the performer by. */
async function searchedSlugs(args: PerformerIdentifyArgs): Promise<string[]> {
  const wanted = new Set([args.name, ...(args.aliases ?? [])].map(foldName));
  const html = await fetchPage(listingUrl(args.name, 1, SEARCH_PAGE_SIZE));
  const slugs: string[] = [];
  for (const item of performerItemsFromPage(parse(html))) {
    if (!wanted.has(foldName(item.title ?? ""))) continue;
    slugs.push(slugFromUrl(item.source_url ?? ""));
  }
  return slugs;
}

/** The profile of a slug as performer:find reads it, or undefined when the page is missing or has no bio. */
async function profileOf(slug: string): Promise<Performer | undefined> {
  try {
    return performerFromBio(await fetchPage(`${BASE_URL}/${slug}/bio`), slug);
  } catch {
    return undefined;
  }
}

/** The first profile among the slugs that the identity rule confirms. */
async function firstConfirmed(slugs: string[], args: PerformerIdentifyArgs): Promise<Performer | undefined> {
  for (const slug of slugs.slice(0, MAX_CANDIDATES)) {
    const performer = await profileOf(slug);
    if (performer !== undefined && isSamePerformer(performer, args)) return performer;
  }
  return undefined;
}

/** Known profile URLs first, then the profile at the name's slug, then the site's own performer search. */
export async function identifyPerformer(args: PerformerIdentifyArgs): Promise<PerformerIdentifyResult> {
  const direct = await firstConfirmed([...knownSlugs(args), slugFromName(args.name)], args);
  if (direct !== undefined) return { performer: direct };
  const searched = await firstConfirmed(await searchedSlugs(args), args);
  if (searched !== undefined) return { performer: searched };
  return {};
}
