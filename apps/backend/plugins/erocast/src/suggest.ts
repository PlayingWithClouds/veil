import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities, fetchText, normalizeText } from "./http.ts";

const DEFAULT_LIMIT = 8;

/** The site's search-box completions for a partly typed query: matching creators, then track titles. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const params = new URLSearchParams({ q: args.query });
  const reply = await fetchText(`${BASE_URL}/api/search/suggest?${params.toString()}`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromReply(reply).slice(0, limit) };
}

/** The suggestions of a reply, {"users": [{name}], "songs": [{title}], …}, each text once. */
export function suggestionsFromReply(reply: string): SiteSuggestion[] {
  const parsed = JSON.parse(reply) as { users?: { name?: string }[]; songs?: { title?: string }[] };
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  const add = (text: string | undefined, kind: SiteSuggestion["kind"]): void => {
    if (text === undefined) return;
    const cleaned = normalizeText(decodeEntities(text));
    const key = cleaned.toLowerCase();
    if (cleaned === "" || seen.has(key)) return;
    seen.add(key);
    suggestions.push({ text: cleaned, kind });
  };
  if (parsed.users) {
    for (const user of parsed.users) add(user.name, "performer");
  }
  if (parsed.songs) {
    for (const song of parsed.songs) add(song.title, "query");
  }
  return suggestions;
}
