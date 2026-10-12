import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";
import { UA } from "./http.ts";
import { enabledSites } from "./sites.ts";

const REDTUBE_SUGGEST_URL = "https://www.redtube.com/video/search_autocomplete";
const DEFAULT_LIMIT = 8;
const SUGGEST_TIMEOUT_MS = 5_000;

// One completion of RedTube's search box. It carries no type: performers and
// plain searches are mixed.
interface RedtubeEntry {
  label?: string;
}

/**
 * RedTube's search-box completions. Only RedTube is asked: Pornhub's endpoint
 * needs a token and session cookie from a page visit, which a plugin's fetch
 * cannot keep between calls, and YouPorn and Tube8 have no endpoint a plain
 * request reaches. Nothing is returned while RedTube is switched off in AYLO_SITES.
 */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  if (!enabledSites().some((site) => site.key === "redtube")) return { suggestions: [] };
  const response = await fetch(`${REDTUBE_SUGGEST_URL}?q=${encodeURIComponent(args.query)}`, {
    headers: {
      "User-Agent": UA,
      Accept: "application/json",
      "X-Requested-With": "XMLHttpRequest",
      Referer: "https://www.redtube.com/",
    },
    signal: AbortSignal.timeout(SUGGEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`redtube autocomplete ${response.status}`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromResponse((await response.json()) as RedtubeEntry[]).slice(0, limit) };
}

/** Turns RedTube's entries into plain-search suggestions, without blanks or repeats. */
export function suggestionsFromResponse(entries: RedtubeEntry[]): SiteSuggestion[] {
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const entry of entries) {
    const text = (entry.label ?? "").trim();
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    suggestions.push({ text, kind: "query" });
  }
  return suggestions;
}
