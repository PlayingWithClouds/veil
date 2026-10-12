import { parse } from "node-html-parser";
import type { SiteSuggestion, SuggestArgs, SuggestResult } from "@playingwithclouds/veil-sdk";
import { SITE, UA } from "./http.ts";

const DEFAULT_LIMIT = 8;
const SUGGEST_TIMEOUT_MS = 5_000;

/** The site's search-box completions: keyword searches and pornstars, as the box lists them. */
export async function suggest(args: SuggestArgs): Promise<SuggestResult> {
  // The box asks for every kind of content ("all") and gets an HTML fragment back.
  const response = await fetch(`${SITE}/suggest/${encodeURIComponent(args.query)}/all/?layout=sections`, {
    headers: { "User-Agent": UA },
    signal: AbortSignal.timeout(SUGGEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`eporner suggest ${response.status}`);
  let limit = DEFAULT_LIMIT;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  return { suggestions: suggestionsFromHtml(await response.text()).slice(0, limit) };
}

/**
 * Reads the suggestion rows: each link's title is the suggestion, and a link
 * into /pornstar/ is a performer. Blanks and repeats are skipped.
 */
export function suggestionsFromHtml(html: string): SiteSuggestion[] {
  const suggestions: SiteSuggestion[] = [];
  const seen = new Set<string>();
  for (const anchor of parse(html).querySelectorAll(".qs-results a[title]")) {
    const text = (anchor.getAttribute("title") ?? "").trim();
    const key = text.toLowerCase();
    if (text === "" || seen.has(key)) continue;
    seen.add(key);
    const href = anchor.getAttribute("href") ?? "";
    const kind: SiteSuggestion["kind"] = href.startsWith("/pornstar/") ? "performer" : "query";
    suggestions.push({ text, kind });
  }
  return suggestions;
}
