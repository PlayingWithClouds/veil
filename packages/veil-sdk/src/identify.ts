// The one rule every performer:identify plugin answers by: a site profile is
// the performer the library asked about only when a name matches exactly and
// the birthdate, when both sides know it, agrees. Strict on purpose: a wrong
// profile would mix two people's details.
import type { Performer, PerformerIdentifyArgs } from "./types.ts";

/** What of a site profile the rule reads. */
type ProfileNames = Pick<Performer, "name" | "aliases" | "birthdate">;

/** A name folded for comparison: lower case, accents and punctuation dropped, spaces collapsed. */
export function foldName(name: string): string {
  return name
    .normalize("NFKD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, " ")
    .trim();
}

/** Every folded name the args know the performer by. */
function wantedNames(args: PerformerIdentifyArgs): Set<string> {
  const names = new Set<string>();
  for (const name of [args.name, ...(args.aliases ?? [])]) {
    const folded = foldName(name);
    if (folded !== "") names.add(folded);
  }
  return names;
}

/** Whether a name or alias of the profile matches a name the args give. */
function sharesName(profile: ProfileNames, wanted: Set<string>): boolean {
  return [profile.name, ...(profile.aliases ?? [])].some((name) => wanted.has(foldName(name)));
}

/** Whether both sides know a birthdate and they differ. Countries are spelled too many ways to compare. */
function birthdatesDiffer(args: PerformerIdentifyArgs, profile: ProfileNames): boolean {
  const known = args.birthdate?.slice(0, 10);
  const found = profile.birthdate?.slice(0, 10);
  if (!known || !found) return false;
  return known !== found;
}

/**
 * Whether a site profile is the performer the args describe: a name or alias
 * in common and no different birthdate. Plugins call this before returning a
 * profile from performer:identify.
 */
export function isSamePerformer(profile: ProfileNames, args: PerformerIdentifyArgs): boolean {
  return sharesName(profile, wantedNames(args)) && !birthdatesDiffer(args, profile);
}
