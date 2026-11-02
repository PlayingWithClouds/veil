import {
  foldName,
  isSamePerformer,
  type Performer,
  type PerformerIdentifyArgs,
  type PerformerIdentifyResult,
} from "@playingwithclouds/veil-sdk";
import { fetchPageAt, type FetchedPage } from "./http.ts";
import { parseProfile, profileSlug } from "./profile.ts";
import { enabledSites, siteForUrl, type SiteConfig } from "./sites.ts";
import { canonicalUrl } from "./urls.ts";

/** Sites asked at once; the first to hold a matching profile ends the search. */
const SITES_PER_ROUND = 8;

/** A model page address on one site. */
export interface ProfileCandidate {
  site: SiteConfig;
  url: string;
}

/** Fetches a candidate page, undefined when the site has no such page. */
export type ProfileFetcher = (candidate: ProfileCandidate) => Promise<FetchedPage | undefined>;

/** The page at a candidate URL, undefined for a missing page or one the site redirected elsewhere. */
async function fetchCandidate(candidate: ProfileCandidate): Promise<FetchedPage | undefined> {
  let page: FetchedPage;
  try {
    page = await fetchPageAt(candidate.site, candidate.url);
  } catch {
    return undefined;
  }
  return page;
}

/**
 * performer:identify: this plugin's sites each list a model under a slug of
 * the name, so the profile URL is built from the name (and aliases) on every
 * enabled site, after any model page the args already know. The first
 * profile that is the same person is returned.
 */
export async function identifyPerformer(args: PerformerIdentifyArgs): Promise<PerformerIdentifyResult> {
  return identifyWith(args, enabledSites(), fetchCandidate);
}

/** The identification over `sites`, reading pages through `fetchProfile`. */
export async function identifyWith(
  args: PerformerIdentifyArgs,
  sites: SiteConfig[],
  fetchProfile: ProfileFetcher,
): Promise<PerformerIdentifyResult> {
  const known = knownProfiles(args);
  const knownResult = await firstMatch(args, known, fetchProfile);
  if (knownResult) return { performer: knownResult };

  const guesses = sites.map((site) => slugGuesses(args, site));
  for (let start = 0; start < guesses.length; start += SITES_PER_ROUND) {
    const round = guesses.slice(start, start + SITES_PER_ROUND);
    const found = await Promise.all(round.map((candidates) => firstMatch(args, candidates, fetchProfile)));
    const best = fullestProfile(found);
    if (best) return { performer: best };
  }
  return {};
}

/** The profile that says the most about the performer (many sites list a name and nothing else); the earliest wins a tie. */
export function fullestProfile(profiles: (Performer | undefined)[]): Performer | undefined {
  let best: Performer | undefined;
  for (const profile of profiles) {
    if (!profile) continue;
    if (!best || filledFields(profile) > filledFields(best)) best = profile;
  }
  return best;
}

/** How many fields of a profile carry a value. */
function filledFields(profile: Performer): number {
  return Object.values(profile).filter((value) => value !== undefined && value !== "").length;
}

/** The model pages of this plugin's sites among the URLs the library already knows. */
export function knownProfiles(args: PerformerIdentifyArgs): ProfileCandidate[] {
  const candidates: ProfileCandidate[] = [];
  for (const url of args.urls ?? []) {
    const site = siteForUrl(url);
    if (!site || profileSlug(site, url) === "") continue;
    candidates.push({ site, url });
  }
  return candidates;
}

/** The profile URLs a site would give the name and each alias, if it lists them. */
export function slugGuesses(args: PerformerIdentifyArgs, site: SiteConfig): ProfileCandidate[] {
  const prefix = modelPrefix(site);
  if (prefix === "") return [];
  const slugs: string[] = [];
  for (const name of [args.name, ...(args.aliases ?? [])]) {
    const slug = foldName(name).split(" ").join("-");
    if (slug !== "" && !slugs.includes(slug)) slugs.push(slug);
  }
  return slugs.map((slug) => ({ site, url: `${site.origin}${prefix}${encodeURIComponent(slug)}/` }));
}

/** The path prefix a site's model pages live under, "" when it has none. */
function modelPrefix(site: SiteConfig): string {
  if (site.modelIndexPath === undefined || site.modelPrefixes.length === 0) return "";
  if (site.modelPrefixes.includes(site.modelIndexPath)) return site.modelIndexPath;
  return site.modelPrefixes[0];
}

/** The first candidate, in order, whose page is a profile of the performer. */
async function firstMatch(
  args: PerformerIdentifyArgs,
  candidates: ProfileCandidate[],
  fetchProfile: ProfileFetcher,
): Promise<Performer | undefined> {
  for (const candidate of candidates) {
    const performer = await matchingProfile(args, candidate, fetchProfile);
    if (performer) return performer;
  }
  return undefined;
}

/** The candidate's profile when its page exists under the same slug and is the performer asked for. */
async function matchingProfile(
  args: PerformerIdentifyArgs,
  candidate: ProfileCandidate,
  fetchProfile: ProfileFetcher,
): Promise<Performer | undefined> {
  const page = await fetchProfile(candidate);
  if (!page) return undefined;
  // A site that sends unknown names to a listing or the home page has no profile.
  if (profileSlug(candidate.site, page.url) !== profileSlug(candidate.site, candidate.url)) return undefined;
  const performer = parseProfile(candidate.site, canonicalUrl(candidate.site, page.url), page.html);
  if (!isSamePerformer(performer, args)) return undefined;
  return performer;
}
