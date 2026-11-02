import { isSamePerformer } from "@playingwithclouds/veil-sdk";
import type { Performer, PerformerIdentifyArgs, PerformerIdentifyResult } from "@playingwithclouds/veil-sdk";
import {
  fetchPornstarPage,
  isPornstarUrl,
  parsePornstarPage,
  pornstarSlugFromUrl,
  pornstarUrl,
  slugFromName,
} from "./performer.ts";

/** Loads a pornstar page's HTML; undefined when the profile does not exist. */
export type PageFetcher = (url: string) => Promise<string | undefined>;

/** Most alias slugs tried after the name, so one lookup stays a few requests. */
const MAX_ALIAS_SLUGS = 4;

/**
 * Profile slugs to try, best first: a profile of this site the library already
 * knows, then the slug of the name, then of each alias.
 */
export function candidateSlugs(args: PerformerIdentifyArgs): string[] {
  const slugs: string[] = [];
  for (const url of args.urls ?? []) {
    if (isPornstarUrl(url)) slugs.push(pornstarSlugFromUrl(url));
  }
  slugs.push(slugFromName(args.name));
  for (const alias of (args.aliases ?? []).slice(0, MAX_ALIAS_SLUGS)) {
    slugs.push(slugFromName(alias));
  }
  return [...new Set(slugs.filter((slug) => slug !== ""))];
}

/** The profile at one slug when it exists and is the person the args describe. */
async function profileAt(
  slug: string,
  args: PerformerIdentifyArgs,
  fetchPage: PageFetcher
): Promise<Performer | undefined> {
  const url = pornstarUrl(slug);
  const html = await fetchPage(url);
  if (html === undefined) return undefined;
  const performer = parsePornstarPage(html, url);
  if (!performer) return undefined;
  if (!isSamePerformer(performer, args)) return undefined;
  return performer;
}

/**
 * performer:identify: xhamster pornstar pages are addressed by name slug, so
 * read the slug candidates directly and keep the first page whose name or alias
 * and birthdate agree with the args. Returns {} when none does.
 */
export async function identifyPerformer(
  args: PerformerIdentifyArgs,
  fetchPage: PageFetcher = fetchPornstarPage
): Promise<PerformerIdentifyResult> {
  for (const slug of candidateSlugs(args)) {
    const performer = await profileAt(slug, args, fetchPage);
    if (performer) return { performer };
  }
  return {};
}
