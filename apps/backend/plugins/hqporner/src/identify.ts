import { parse } from "node-html-parser";
import { fetchHtml, foldName, isSamePerformer } from "@playingwithclouds/veil-sdk";
import type { Performer, PerformerIdentifyArgs, PerformerIdentifyResult } from "@playingwithclouds/veil-sdk";
import { performerList } from "./catalog.ts";
import { BASE_URL, actressSlugFromUrl } from "./http.ts";
import { isActressPage, performerFromActressPage } from "./scrape.ts";

const MAX_CANDIDATES = 4;

/** The slug a name gets on the site: lower case words joined by hyphens ("Mia Malkova" → mia-malkova). */
export function slugFromName(name: string): string {
  return foldName(name).replace(/ /g, "-");
}

/** Slugs of hqporner actress URLs the library already knows. */
function knownSlugs(args: PerformerIdentifyArgs): string[] {
  const slugs: string[] = [];
  for (const url of args.urls ?? []) {
    if (!/^https?:\/\/(?:www\.)?hqporner\.com\/actress\//.test(url)) continue;
    slugs.push(actressSlugFromUrl(url));
  }
  return slugs;
}

/** Slugs of the site's actress catalog entries named like the performer. */
async function catalogSlugs(args: PerformerIdentifyArgs): Promise<string[]> {
  const wanted = new Set([args.name, ...(args.aliases ?? [])].map(foldName));
  const { items } = await performerList({ query: args.name });
  const named = items.filter((item) => wanted.has(foldName(item.title ?? "")));
  return named.map((item) => actressSlugFromUrl(item.source_url ?? ""));
}

/** The actress page of a slug as performer:find reads it, or undefined when the site has no such actress. */
async function profileOf(slug: string): Promise<Performer | undefined> {
  if (slug === "") return undefined;
  try {
    const root = parse(await fetchHtml(`${BASE_URL}/actress/${slug}`));
    if (!isActressPage(root)) return undefined;
    return performerFromActressPage(slug, root);
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

/**
 * Known profile URLs first, then the actress page at the name's slug, then the
 * site's actress catalog. The profile has no birthdate, so the name alone decides.
 */
export async function identifyPerformer(args: PerformerIdentifyArgs): Promise<PerformerIdentifyResult> {
  const direct = await firstConfirmed([...knownSlugs(args), slugFromName(args.name)], args);
  if (direct !== undefined) return { performer: direct };
  const listed = await firstConfirmed(await catalogSlugs(args), args);
  if (listed !== undefined) return { performer: listed };
  return {};
}
