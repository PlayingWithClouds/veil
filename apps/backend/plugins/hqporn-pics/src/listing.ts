import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";
import { BASE_URL } from "./http.ts";
import { galleriesFromRoot } from "./cards.ts";

const DEFAULT_PAGE_SIZE = 20;
// Each category page is one fixed list (the site has no paging), so a browse
// walks the categories in turn, CATEGORY_SPAN galleries from each.
const CATEGORY_SPAN = 100;
const BROWSE_CATEGORIES = [
  "hardcore",
  "amateurs",
  "anal",
  "asians",
  "babes",
  "bbw",
  "bigtits",
  "blondes",
  "brunettes",
  "grannies",
  "groupsex",
  "hairy",
  "latinas",
  "lesbians",
  "matures",
  "redheads",
];

/**
 * Gallery stubs from the category lists, paged by offset. The site's search is
 * not served (its URL returns the front page), so a query picks the categories
 * whose name it matches instead.
 */
export async function galleryList(args: ListArgs): Promise<ItemsResult> {
  const categories = categoriesFor(args.query);
  let limit = DEFAULT_PAGE_SIZE;
  if (args.limit !== undefined && args.limit > 0) limit = args.limit;
  let offset = 0;
  if (args.offset !== undefined && args.offset > 0) offset = args.offset;

  const categoryIndex = Math.floor(offset / CATEGORY_SPAN);
  if (categoryIndex >= categories.length) return { items: [] };
  const innerOffset = offset % CATEGORY_SPAN;
  const categoryItems = await fetchCategoryGalleries(categories[categoryIndex]);
  return { items: categoryItems.slice(innerOffset, innerOffset + Math.min(limit, CATEGORY_SPAN)) };
}

/** The categories a query selects: those whose name contains the query, or all of them without a query. */
export function categoriesFor(query: string | undefined): string[] {
  if (query === undefined) return BROWSE_CATEGORIES;
  const needle = query.toLowerCase().replace(/\s+/g, "");
  if (needle === "") return BROWSE_CATEGORIES;
  return BROWSE_CATEGORIES.filter((category) => category.includes(needle));
}

/** The gallery stubs on one category page. */
async function fetchCategoryGalleries(category: string): Promise<DiscoveredItem[]> {
  const html = await fetchHtml(`${BASE_URL}/${category}/`);
  return galleriesFromRoot(parse(html)).slice(0, CATEGORY_SPAN);
}
