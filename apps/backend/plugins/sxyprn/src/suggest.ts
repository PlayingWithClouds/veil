import { BROWSER_HEADERS, type SiteSuggestion, type SuggestArgs, type SuggestResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities, normalizeText } from "./http.ts";
import { keywordFromQuery } from "./listing.ts";

const DEFAULT_LIMIT = 8;

/** The site's search-box completions for a partly typed query, as its own live search asks for them. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  const reply = await fetch(`${BASE_URL}/php/livesearch.php`, {
    method: "POST",
    headers: {
      ...BROWSER_HEADERS,
      "Content-Type": "application/x-www-form-urlencoded",
      "X-Requested-With": "XMLHttpRequest",
    },
    body: new URLSearchParams({ uid: "", aid: "all", key: keywordFromQuery(args.query), c: "ls7" }).toString(),
  });
  if (!reply.ok) throw new Error(`sxyprn: HTTP ${reply.status} for the search suggestions`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromReply(await reply.text()).slice(0, limit) };
}

/**
 * The keyword suggestions of a live-search reply, {"searches": [{"title", "link"}],
 * "blogs": [...]}. Only the keyword matches are suggestions; the posts are not.
 */
export function suggestionsFromReply(body: string): SiteSuggestion[] {
  const parsed = JSON.parse(body) as { searches?: { title?: unknown }[] };
  if (!Array.isArray(parsed.searches)) return [];
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const entry of parsed.searches) {
    if (typeof entry.title !== "string") continue;
    const text = normalizeText(decodeEntities(entry.title.replace(/<[^>]+>/g, " ")));
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    suggestions.push({ text, kind: "query" });
  }
  return suggestions;
}
