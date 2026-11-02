import { isSamePerformer } from "@playingwithclouds/veil-sdk";
import type { Performer, PerformerIdentifyArgs, PerformerIdentifyResult } from "@playingwithclouds/veil-sdk";
import {
  fetchProfilePage,
  hasPornstarPicture,
  isProfileUrl,
  parseProfilePage,
  profileSlugFromUrl,
  profileUrl,
  slugFromName,
} from "./performer.ts";

/** Loads a profile page's HTML; undefined when the profile does not exist. */
export type PageFetcher = (url: string) => Promise<string | undefined>;

/** Most alias slugs tried after the name, so one lookup stays a few requests. */
const MAX_ALIAS_SLUGS = 4;

/** A profile slug to read and whether the library already knows it as this person's page. */
export interface Candidate {
  slug: string;
  known: boolean;
}

/**
 * Profile slugs to try, best first: profiles of this site the library already
 * knows, then the slug of the name, then of each alias.
 */
export function candidates(args: PerformerIdentifyArgs): Candidate[] {
  const found: Candidate[] = [];
  for (const url of args.urls ?? []) {
    if (isProfileUrl(url)) found.push({ slug: profileSlugFromUrl(url), known: true });
  }
  const names = [args.name, ...(args.aliases ?? []).slice(0, MAX_ALIAS_SLUGS)];
  for (const name of names) {
    found.push({ slug: slugFromName(name), known: false });
  }
  const seen = new Set<string>();
  return found.filter((candidate) => {
    if (candidate.slug === "" || seen.has(candidate.slug)) return false;
    seen.add(candidate.slug);
    return true;
  });
}

/**
 * Whether a profile found by name slug can be trusted to be a performer. Members
 * and performers share one page layout; only a performer photo tells them apart,
 * unless the library already holds the page as this person's.
 */
function isTrusted(performer: Performer, candidate: Candidate): boolean {
  return candidate.known || hasPornstarPicture(performer);
}

/** The profile at one candidate slug when it exists and is the person the args describe. */
async function profileAt(
  candidate: Candidate,
  args: PerformerIdentifyArgs,
  fetchPage: PageFetcher
): Promise<Performer | undefined> {
  const url = profileUrl(candidate.slug);
  const html = await fetchPage(url);
  if (html === undefined) return undefined;
  const performer = parseProfilePage(html, url);
  if (!performer) return undefined;
  if (!isTrusted(performer, candidate)) return undefined;
  if (!isSamePerformer(performer, args)) return undefined;
  return performer;
}

/**
 * performer:identify: tnaflix profiles are addressed by name slug (or a known
 * URL), so read the candidates directly and keep the first whose name or alias
 * agrees with the args. Returns {} when none does.
 */
export async function identifyPerformer(
  args: PerformerIdentifyArgs,
  fetchPage: PageFetcher = fetchProfilePage
): Promise<PerformerIdentifyResult> {
  for (const candidate of candidates(args)) {
    const performer = await profileAt(candidate, args, fetchPage);
    if (performer) return { performer };
  }
  return {};
}
