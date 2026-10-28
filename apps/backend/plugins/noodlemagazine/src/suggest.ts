import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";
import { decodeEntities } from "./http.ts";
import { fetchFromMirrors } from "./mirrors.ts";

const DEFAULT_LIMIT = 8;

/** The site's search-box completions for a partly typed query, from the first mirror that answers. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const page = await fetchFromMirrors(`/api/autocomplete?q=${encodeURIComponent(args.query)}`, isAutocompleteReply);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromBody(page.body).slice(0, limit) };
}

/** Whether a body is an autocomplete reply, not a challenge or error page. */
function isAutocompleteReply(body: string): boolean {
  try {
    const parsed = JSON.parse(jsonPayload(body)) as { results?: unknown };
    return Array.isArray(parsed.results);
  } catch {
    return false;
  }
}

/**
 * Reads the autocomplete reply, {"results": [...]}. Behind FlareSolverr the JSON
 * arrives inside a <pre> block, so the text is taken from there when present.
 */
export function suggestionsFromBody(body: string): SiteSuggestion[] {
  const json = jsonPayload(body);
  const parsed = JSON.parse(json) as { results?: unknown };
  if (!Array.isArray(parsed.results)) return [];
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const entry of parsed.results) {
    if (typeof entry !== "string") continue;
    const text = entry.trim();
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    suggestions.push({ text, kind: "query" });
  }
  return suggestions;
}

/** The JSON text of a reply: the body itself, or the content of its <pre> block. */
function jsonPayload(body: string): string {
  const match = body.match(/<pre[^>]*>([\s\S]*?)<\/pre>/);
  if (!match) return body.trim();
  return decodeEntities(match[1]).trim();
}
