import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";
import { BASE_URL, cleanTitle } from "./http.ts";

// One entry from pornpics' search/list JSON endpoint.
interface SearchRow {
  g_url: string;
  t_url?: string;
  t_url_460?: string;
  desc?: string;
  gid?: string;
}

// Queries pornpics' JSON search endpoint. It powers both keyword search and
// browse (via broad terms), paginating cleanly by offset.
export async function searchGalleries(
  query: string,
  limit: number,
  offset: number
): Promise<DiscoveredItem[]> {
  const url =
    `${BASE_URL}/search/srch.php?q=${encodeURIComponent(query)}` +
    `&lang=en&limit=${limit}&offset=${offset}`;
  const response = await fetch(url, {
    headers: { ...BROWSER_HEADERS, "X-Requested-With": "XMLHttpRequest" },
  });
  if (!response.ok) {
    throw new Error(`pornpics search ${response.status} for ${url}`);
  }
  const rows = (await response.json()) as SearchRow[];
  if (!Array.isArray(rows)) return [];

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();
  for (const row of rows) {
    if (!row.g_url || !row.gid || seen.has(row.gid)) continue;
    seen.add(row.gid);
    items.push({
      title: cleanTitle(row.desc ?? ""),
      media_type: "gallery",
      source_url: row.g_url,
      external_id: `pornpics-${row.gid}`,
      poster_path: row.t_url_460 ?? row.t_url,
    });
  }
  return items;
}
