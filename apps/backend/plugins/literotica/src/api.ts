import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";

export const SITE_URL = "https://www.literotica.com";
const API_URL = "https://literotica.com/api/3";

/** Results the search API returns per page. */
export const SEARCH_PAGE_SIZE = 50;

/** A story's tag as the API lists it. */
export interface ApiTag {
  tag: string | number;
}

/** One work in a series, as a story's own `series.items` lists it. */
export interface ApiSeriesItem {
  id: number;
  title: string;
  url: string;
}

/** The series a work belongs to; `url` is empty for older series. */
export interface ApiSeriesLink {
  meta: { id: number; title: string; url: string };
  items: ApiSeriesItem[];
}

/** One work (story, audio, poem …) as search and series listings return it. */
export interface ApiWork {
  id: number;
  type: string;
  title: string;
  url: string;
  authorname: string;
  description: string;
  date_approve: string;
  category: number;
  category_info: { pageUrl: string };
  rate_all?: number;
  view_count?: number;
  words_count?: number;
  tags?: ApiTag[];
  /** an empty array (or missing) for a standalone work */
  series?: ApiSeriesLink | unknown[] | null;
}

/** Answer of /stories/<slug>: the work plus one page of its text. */
export interface ApiStoryPage {
  meta: { pages_count: number };
  submission: ApiWork;
  pageText: string;
}

/** Series record of /series/<slug or id>. */
export interface ApiSeries {
  id: number;
  url: string | null;
  title: string;
  state: string;
  description: string | null;
  introduction?: string | null;
  user?: { username: string };
}

/** Answer of /search/stories. */
export interface ApiSearchPage {
  data: ApiWork[];
  meta: { pageSize: number; total: number };
}

/** Fetches one JSON document from the site's own API. */
async function fetchJson<T>(path: string): Promise<T> {
  const response = await fetch(`${API_URL}/${path}`, { headers: { ...BROWSER_HEADERS, Accept: "application/json" } });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status} fetching ${path}`);
  }
  return (await response.json()) as T;
}

/** One page (1-based) of a work's text and its metadata. */
export function fetchStoryPage(slug: string, page: number): Promise<ApiStoryPage> {
  if (page === 1) {
    return fetchJson<ApiStoryPage>(`stories/${encodeURIComponent(slug)}`);
  }
  const params = encodeURIComponent(JSON.stringify({ contentPage: page }));
  return fetchJson<ApiStoryPage>(`stories/${encodeURIComponent(slug)}?params=${params}`);
}

/** The series record for a slug or numeric id. */
export async function fetchSeries(slugOrId: string | number): Promise<ApiSeries> {
  const answer = await fetchJson<{ data: ApiSeries }>(`series/${encodeURIComponent(String(slugOrId))}`);
  return answer.data;
}

/** The works of a series, in reading order. */
export function fetchSeriesWorks(slugOrId: string | number): Promise<ApiWork[]> {
  return fetchJson<ApiWork[]>(`series/${encodeURIComponent(String(slugOrId))}/works`);
}

/** Filters of the search API; absent fields don't filter. */
export interface SearchParams {
  q?: string;
  sort?: "date";
  categories?: number[];
  author?: string;
  page: number;
}

/** One page (1-based) of the search API. */
export function fetchSearchPage(params: SearchParams): Promise<ApiSearchPage> {
  return fetchJson<ApiSearchPage>(`search/stories?params=${encodeURIComponent(JSON.stringify(params))}`);
}

/** The reader URL of a work. */
export function workUrl(slug: string): string {
  return `${SITE_URL}/s/${slug}`;
}

/** The reader URL of a series; older series without a slug are addressed by id. */
export function seriesUrl(series: { id: number; url: string | null }): string {
  if (series.url === null || series.url === "") {
    return `${SITE_URL}/series/se/${series.id}`;
  }
  return `${SITE_URL}/series/se/${series.url}`;
}

/** An author's profile URL. */
export function authorUrl(username: string): string {
  return `${SITE_URL}/authors/${encodeURIComponent(username)}`;
}

/** What a Literotica URL points at. */
export type SiteTarget =
  | { kind: "work"; slug: string }
  | { kind: "series"; slugOrId: string }
  | { kind: "category"; slug: string }
  | { kind: "author"; username: string }
  | { kind: "newest" }
  | { kind: "unknown" };

/** Classifies a Literotica page URL by its path. */
export function targetOf(url: string): SiteTarget {
  const segments = new URL(url).pathname.split("/").filter((segment) => segment !== "");
  const [first, second, third] = segments;
  if (first === "s" && second !== undefined) {
    return { kind: "work", slug: decodeURIComponent(second) };
  }
  if (first === "series" && second === "se" && third !== undefined) {
    return { kind: "series", slugOrId: decodeURIComponent(third) };
  }
  if (first === "c" && second !== undefined) {
    return { kind: "category", slug: decodeURIComponent(second) };
  }
  if (first === "authors" && second !== undefined) {
    return { kind: "author", username: decodeURIComponent(second) };
  }
  if (first === "new" && second === "stories") {
    return { kind: "newest" };
  }
  return { kind: "unknown" };
}

/** Whether a work is a story proper; audio and poems share the API. */
export function isStory(work: ApiWork): boolean {
  return work.type === "story";
}

/** The series a work is linked to, or undefined for a standalone work. */
export function seriesOf(work: ApiWork): ApiSeriesLink | undefined {
  const series = work.series;
  if (series === undefined || series === null || Array.isArray(series)) {
    return undefined;
  }
  return series as ApiSeriesLink;
}
