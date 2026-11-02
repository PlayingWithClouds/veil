import {
  foldName,
  isSamePerformer,
  type Performer,
  type PerformerIdentifyArgs,
  type PerformerIdentifyResult,
} from "@playingwithclouds/veil-sdk";
import { pageHeaders } from "./http.ts";
import { canonicalProfileUrl, isProfileUrl, parseProfile, profileUrlForSlug } from "./profile.ts";
import { ALL_SITES, siteForUrl, type Site } from "./sites.ts";

/** Fetches a profile page; undefined when the site has none at that URL. */
export type ProfileFetcher = (url: string) => Promise<string | undefined>;

/** How many of the library's aliases are tried as profile slugs after the name. */
const MAX_ALIAS_SLUGS = 2;

/** Fetches a profile page, answering undefined on 404 and throwing on any other failure. */
export async function fetchProfileHtml(url: string): Promise<string | undefined> {
  const response = await fetch(url, {
    headers: pageHeaders(),
    redirect: "follow",
    signal: AbortSignal.timeout(20_000),
  });
  if (response.status === 404) return undefined;
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  return response.text();
}

/**
 * performer:identify. Looks for the performer's profile on the Aylo sites by
 * the URL shape each site uses for a name (and on the sites already linked in
 * args.urls first), and returns it only when isSamePerformer agrees. Pornhub
 * profiles carry no birthdate, so when the library knows one, the other sites
 * are tried for a profile that confirms it before settling for a name match.
 */
export async function identifyPerformer(
  args: PerformerIdentifyArgs,
  fetchHtml: ProfileFetcher = fetchProfileHtml,
): Promise<PerformerIdentifyResult> {
  let unconfirmed: Performer | undefined;
  let contradicted = false;
  for (const site of sitesToTry(args)) {
    const found = await identifyOnSite(site, args, fetchHtml);
    if (found.contradicted) contradicted = true;
    if (!found.performer) continue;
    if (!args.birthdate || found.performer.birthdate) return { performer: found.performer };
    if (!unconfirmed) unconfirmed = found.performer;
  }
  // Another site's profile of that name has a different birthdate, so a bare name match elsewhere proves nothing.
  if (unconfirmed && !contradicted) return { performer: unconfirmed };
  return {};
}

/** Sites in the order to ask: those the library already links a profile of, then Pornhub first. */
function sitesToTry(args: PerformerIdentifyArgs): Site[] {
  const sites: Site[] = [];
  for (const url of args.urls || []) {
    const site = siteForUrl(url);
    if (site && isProfileUrl(url) && !sites.includes(site)) sites.push(site);
  }
  for (const site of ALL_SITES) {
    if (!sites.includes(site)) sites.push(site);
  }
  return sites;
}

/** What one site said: the matching profile, and whether a profile of the right name had another birthdate. */
interface SiteAnswer {
  performer?: Performer;
  contradicted: boolean;
}

/** The first profile on one site that is the described performer. */
async function identifyOnSite(
  site: Site,
  args: PerformerIdentifyArgs,
  fetchHtml: ProfileFetcher,
): Promise<SiteAnswer> {
  let contradicted = false;
  for (const url of candidateUrls(site, args)) {
    const html = await fetchHtml(url);
    if (!html) continue;
    const performer = parseProfile(site, url, html);
    if (!performer) continue;
    if (isSamePerformer(performer, args)) return { performer, contradicted };
    if (isSamePerformer(performer, { ...args, birthdate: undefined })) contradicted = true;
  }
  return { contradicted };
}

/** Profile URLs to try on a site: known ones first, then the name's and a few aliases' slugs. */
export function candidateUrls(site: Site, args: PerformerIdentifyArgs): string[] {
  const urls: string[] = [];
  for (const known of args.urls || []) {
    if (siteForUrl(known) === site && isProfileUrl(known)) urls.push(canonicalProfileUrl(site, known));
  }
  const names = [args.name, ...(args.aliases || []).slice(0, MAX_ALIAS_SLUGS)];
  for (const name of names) {
    const slug = foldName(name).replace(/ /g, "-");
    if (slug !== "") urls.push(profileUrlForSlug(site, slug));
  }
  return urls.filter((url, index) => urls.indexOf(url) === index);
}
