import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";
import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL } from "./http.ts";

const DEFAULT_LIMIT = 8;
const SUGGEST_TIMEOUT_MS = 5_000;

// One entry of the site's /autocomplete.php: a search, with a link when it is
// a pornstar page.
interface AutocompleteEntry {
  query?: string;
  link?: string;
}

/** The site's search-box completions; the ones that are a pornstar page count as performers. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const url = `${BASE_URL}/autocomplete.php?term=${encodeURIComponent(args.query)}&lang=en&orientation=straight`;
  const response = await fetch(url, {
    headers: BROWSER_HEADERS,
    signal: AbortSignal.timeout(SUGGEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`pornpics autocomplete ${response.status}`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromResponse((await response.json()) as AutocompleteEntry[]).slice(0, limit) };
}

/** Turns the autocomplete entries into suggestions, without blanks or repeats. */
export function suggestionsFromResponse(entries: AutocompleteEntry[]): SiteSuggestion[] {
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const entry of entries) {
    const text = (entry.query ?? "").trim();
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    const kind: SiteSuggestion["kind"] = entry.link ? "performer" : "query";
    suggestions.push({ text, kind });
  }
  return suggestions;
}
