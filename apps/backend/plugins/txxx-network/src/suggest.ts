import { configuredHost, fetchJson } from "./http.ts";
import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";

const DEFAULT_LIMIT = 8;

// What the site's header search box gets from /api/suggester.php: plain
// searches as strings, and models, categories and channels as records.
interface SuggesterResponse {
  keywords?: string[];
  models?: { title?: string }[];
  categories?: { title?: string }[];
  channels?: { title?: string }[];
}

/** The configured site's search-box completions: searches first, then performers, categories, channels. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const response = await fetchJson<SuggesterResponse>(
    configuredHost(),
    `/api/suggester.php?c=${encodeURIComponent(args.query)}`,
  );
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromResponse(response).slice(0, limit) };
}

/** Flattens the suggester's groups into one list, best group first, without blanks or repeats. */
export function suggestionsFromResponse(response: SuggesterResponse): SiteSuggestion[] {
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  const add = (text: string | undefined, kind: SiteSuggestion["kind"]) => {
    const trimmed = (text ?? "").trim();
    const key = trimmed.toLowerCase();
    if (trimmed === "" || seen.has(key)) return;
    seen.add(key);
    suggestions.push({ text: trimmed, kind });
  };
  for (const keyword of response.keywords ?? []) add(keyword, "query");
  for (const model of response.models ?? []) add(model.title, "performer");
  for (const category of response.categories ?? []) add(category.title, "tag");
  for (const channel of response.channels ?? []) add(channel.title, "studio");
  return suggestions;
}
