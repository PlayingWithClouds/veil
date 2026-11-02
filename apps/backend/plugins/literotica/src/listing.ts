import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import {
  fetchSearchPage,
  isStory,
  SEARCH_PAGE_SIZE,
  seriesOf,
  seriesUrl,
  targetOf,
  workUrl,
  type ApiWork,
  type SearchParams,
} from "./api.ts";
import { categoryIdOf } from "./categories.ts";
import { seriesExternalId, workExternalId } from "./story.ts";
import { isoDateOf } from "./text.ts";

/** Serves "story:list": the newest stories, a search, or a category or author page, as one entry per story or series. */
export async function storyList(args: ListArgs): Promise<ItemsResult> {
  const filters = filtersOf(args);
  const offset = args.offset ?? 0;
  const limit = args.limit ?? SEARCH_PAGE_SIZE;
  const stories = await fetchStories(filters, offset, limit);
  return { items: dedupe(stories.map(itemOf)) };
}

/** The search filters a list request asks for. */
function filtersOf(args: ListArgs): Omit<SearchParams, "page"> {
  const filters: Omit<SearchParams, "page"> = {};
  if (args.query !== undefined) {
    filters.q = args.query;
  } else {
    filters.sort = "date";
  }
  if (args.url !== undefined) {
    Object.assign(filters, pageFilters(args.url));
  }
  return filters;
}

/** The filters of a category, author or "new stories" page. */
function pageFilters(url: string): Omit<SearchParams, "page"> {
  const target = targetOf(url);
  if (target.kind === "newest") {
    return {};
  }
  if (target.kind === "author") {
    return { author: target.username };
  }
  if (target.kind !== "category") {
    throw new Error(`Cannot list stories of ${url}: expected a category, author or new-stories page`);
  }
  const categoryId = categoryIdOf(target.slug);
  if (categoryId === undefined) {
    throw new Error(`Unknown Literotica category: ${target.slug}`);
  }
  return { categories: [categoryId] };
}

/**
 * The stories of the search pages covering `offset` to `offset + limit`. Audio
 * shares the search results (and can fill a whole relevance page's top), so
 * the offset counts stories from the first page's start, exactly so only for
 * multiples of the page size.
 */
async function fetchStories(filters: Omit<SearchParams, "page">, offset: number, limit: number): Promise<ApiWork[]> {
  const firstPage = Math.floor(offset / SEARCH_PAGE_SIZE) + 1;
  const lastPage = Math.floor((offset + limit - 1) / SEARCH_PAGE_SIZE) + 1;
  const stories: ApiWork[] = [];
  for (let page = firstPage; page <= lastPage; page++) {
    const answer = await fetchSearchPage({ ...filters, page });
    stories.push(...answer.data.filter(isStory));
    if (answer.data.length < SEARCH_PAGE_SIZE) {
      break;
    }
  }
  const skipped = offset - (firstPage - 1) * SEARCH_PAGE_SIZE;
  return stories.slice(skipped, skipped + limit);
}

/** The listing entry of a work; a work in a series stands for the series. */
function itemOf(work: ApiWork): DiscoveredItem {
  const series = seriesOf(work);
  if (series === undefined) {
    return workItem(work);
  }
  const item = workItem(work);
  item.title = series.meta.title;
  item.source_url = seriesUrl(series.meta);
  item.external_id = seriesExternalId(series.meta.id);
  item.count = series.items.length;
  return item;
}

/** The listing entry of a work taken alone. */
function workItem(work: ApiWork): DiscoveredItem {
  const item: DiscoveredItem = {
    title: work.title,
    media_type: "story",
    source_url: workUrl(work.url),
    external_id: workExternalId(work.id),
    date: isoDateOf(work.date_approve),
    view_count: work.view_count,
  };
  if (work.rate_all !== undefined && work.rate_all > 0) {
    item.rating = Math.round(work.rate_all * 20) / 10;
  }
  return item;
}

/** The entries without later repeats of the same story, which a series' works would cause. */
function dedupe(items: DiscoveredItem[]): DiscoveredItem[] {
  const seen = new Set<string>();
  return items.filter((item) => {
    if (seen.has(item.external_id)) {
      return false;
    }
    seen.add(item.external_id);
    return true;
  });
}
