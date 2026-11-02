import { normalizeText } from "./site.ts";

// Bracketed phrases longer than this are sentences about the audio, not tags.
const MAX_TAG_LENGTH = 40;

/** A title split into its clean wording and the bracketed tags it carried. */
export interface SplitTitle {
  title: string;
  tags: string[];
}

/** Lists the lowercased bracketed phrases of a title, once each. */
function bracketTags(rawTitle: string): string[] {
  const tags: string[] = [];
  for (const match of rawTitle.matchAll(/\[([^\]]*)\]/g)) {
    const tag = normalizeText(match[1]).toLowerCase();
    if (tag === "" || tag.length > MAX_TAG_LENGTH) continue;
    if (tags.includes(tag)) continue;
    tags.push(tag);
  }
  return tags;
}

/**
 * Separates "[F4M] [ASMR] Name [Whispers]" into the title "Name" and its tags.
 * A title made only of brackets stays whole, so the audio is never nameless.
 */
export function splitTitle(rawTitle: string): SplitTitle {
  const tags = bracketTags(rawTitle);
  const title = normalizeText(rawTitle.replace(/\[[^\]]*\]/g, " "));
  if (title === "") {
    return { title: normalizeText(rawTitle), tags };
  }
  return { title, tags };
}
