import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";
import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";

// The suggest service the site's search box calls runs on its own port.
const SUGGEST_URL = "https://www.tnaflix.com:8001/suggest";
const DEFAULT_LIMIT = 8;
const SUGGEST_TIMEOUT_MS = 5_000;

// One entry of the service's answer; link_type says what the link opens.
interface SuggestEntry {
  query?: string;
  link_type?: string;
}

/** The site's search-box completions: searches, pornstars and channels, most popular first. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const response = await fetch(`${SUGGEST_URL}?prefix=${encodeURIComponent(args.query)}`, {
    headers: { ...BROWSER_HEADERS, Accept: "application/json", Referer: "https://www.tnaflix.com/" },
    signal: AbortSignal.timeout(SUGGEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`tnaflix suggest ${response.status}`);
  const body = (await response.json()) as { suggestions?: SuggestEntry[] };
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromResponse(body.suggestions ?? []).slice(0, limit) };
}

/** Maps the service's link types onto suggestion kinds; anything else is a plain search. */
function kindOf(linkType: string | undefined): SiteSuggestion["kind"] {
  if (linkType === "pornstar") return "performer";
  if (linkType === "channel") return "studio";
  return "query";
}

/** Turns the service's entries into suggestions, without blanks or repeats. */
export function suggestionsFromResponse(entries: SuggestEntry[]): SiteSuggestion[] {
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const entry of entries) {
    const text = (entry.query ?? "").trim();
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    suggestions.push({ text, kind: kindOf(entry.link_type) });
  }
  return suggestions;
}
