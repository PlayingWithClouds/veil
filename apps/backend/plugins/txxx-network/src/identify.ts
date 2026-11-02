import {
  foldName,
  isSamePerformer,
  type Performer,
  type PerformerIdentifyArgs,
  type PerformerIdentifyResult,
} from "@playingwithclouds/veil-sdk";
import { NETWORK_DOMAINS, configuredHost, hostFromUrl } from "./http.ts";
import { fetchLanguage, fetchModel, modelDir, parseModel, type ApiModel, type SiteLanguage } from "./profile.ts";

/** Network sites asked at once; the first with a matching model ends the search. */
const HOSTS_PER_ROUND = 6;

/** Looks a model up on a site by slug, undefined when the site has none. */
export type ModelLookup = (host: string, dir: string) => Promise<ApiModel | undefined>;

/** A model slug to ask a network site for. */
export interface ModelGuess {
  host: string;
  dir: string;
}

/** The model record at a site, undefined when it has none or the site fails. */
async function lookupModel(host: string, dir: string): Promise<ApiModel | undefined> {
  try {
    return await fetchModel(host, dir);
  } catch {
    return undefined;
  }
}

/**
 * performer:identify: every network site keeps its own model records under a
 * slug of the name. Known /models/ pages of the network come first, then the
 * name and aliases as slugs on the browse host and the other network sites.
 */
export async function identifyPerformer(args: PerformerIdentifyArgs): Promise<PerformerIdentifyResult> {
  const found = await identifyWith(args, lookupModel);
  if (!found) return {};
  const language = await fetchLanguage(found.host);
  return { performer: parseModel(found.host, found.model, language) };
}

/** The first network site whose model record is the performer the args describe. */
export async function identifyWith(
  args: PerformerIdentifyArgs,
  lookup: ModelLookup,
): Promise<{ host: string; model: ApiModel } | undefined> {
  const known = knownGuesses(args);
  const knownMatch = await firstMatch(args, known, lookup);
  if (knownMatch) return knownMatch;

  const guesses = hostsToTry().map((host) => nameGuesses(args, host));
  for (let start = 0; start < guesses.length; start += HOSTS_PER_ROUND) {
    const round = guesses.slice(start, start + HOSTS_PER_ROUND);
    const found = await Promise.all(round.map((hostGuesses) => firstMatch(args, hostGuesses, lookup)));
    const match = found.find((entry) => entry !== undefined);
    if (match) return match;
  }
  return undefined;
}

/** The browse host first, then the other network sites. */
function hostsToTry(): string[] {
  const hosts = [configuredHost()];
  for (const host of NETWORK_DOMAINS) {
    if (!hosts.includes(host)) hosts.push(host);
  }
  return hosts;
}

/** The model pages of the network among the URLs the library knows. */
export function knownGuesses(args: PerformerIdentifyArgs): ModelGuess[] {
  const guesses: ModelGuess[] = [];
  for (const url of args.urls ?? []) {
    const host = hostFromUrl(url);
    const dir = modelDir(url);
    if (dir === "" || !NETWORK_DOMAINS.includes(host)) continue;
    guesses.push({ host, dir });
  }
  return guesses;
}

/** The slugs of the name and each alias on one site. */
export function nameGuesses(args: PerformerIdentifyArgs, host: string): ModelGuess[] {
  const guesses: ModelGuess[] = [];
  for (const name of [args.name, ...(args.aliases ?? [])]) {
    const dir = foldName(name).split(" ").join("-");
    if (dir === "" || guesses.some((guess) => guess.dir === dir)) continue;
    guesses.push({ host, dir });
  }
  return guesses;
}

/** The first guess, in order, whose record is the performer. */
async function firstMatch(
  args: PerformerIdentifyArgs,
  guesses: ModelGuess[],
  lookup: ModelLookup,
): Promise<{ host: string; model: ApiModel } | undefined> {
  for (const guess of guesses) {
    const model = await lookup(guess.host, guess.dir);
    if (model && isSameModel(guess.host, model, args)) return { host: guess.host, model };
  }
  return undefined;
}

/** Whether the record is the performer: same name or alias, no other birthdate, no other gender. */
export function isSameModel(host: string, model: ApiModel, args: PerformerIdentifyArgs): boolean {
  const profile = parseModel(host, model, {} as SiteLanguage);
  if (!isSamePerformer(profile, args)) return false;
  return !gendersDiffer(profile, args);
}

/** Whether both sides state a gender and the two differ. */
function gendersDiffer(profile: Performer, args: PerformerIdentifyArgs): boolean {
  if (!profile.gender || !args.gender) return false;
  return profile.gender !== args.gender.toLowerCase();
}
