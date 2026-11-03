// Our own categories: a taxonomy over the sites' tags, seeded with a default
// list the user edits. Media is in a category when one of its tags is a
// match tag or folds to one of its terms, or it's in a subcategory.
import type {
	Category,
	CategoryInput,
	CategoryTile,
	FieldsSelection,
	MediaKind
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** A category as a tile shows it, with how much of one kind is in it. */
const TILE = {
	count: true,
	preview: { uri: true },
	category: { id: true, name: true, position: true, parent: { id: true } }
} as const;

/** A category's page and editor. */
const CATEGORY_DETAIL = {
	id: true,
	name: true,
	slug: true,
	matchTerms: true,
	matchTags: { id: true, name: true },
	cover: { uri: true },
	parent: { id: true, name: true },
	children: { id: true, name: true },
	videoCount: true,
	galleryCount: true
} as const;

/** A category to pick as a parent. */
const PARENT_OPTION = { id: true, name: true, parent: { id: true } } as const;

export type CategoryTileRow = FieldsSelection<CategoryTile, typeof TILE>;
export type CategoryDetail = FieldsSelection<Category, typeof CATEGORY_DETAIL>;
export type ParentOption = FieldsSelection<Category, typeof PARENT_OPTION>;

/** A group of tiles: a top-level category and the categories inside it. */
export interface TileGroup {
	parent: CategoryTileRow;
	children: CategoryTileRow[];
}

/** Every category with how much media of a kind it holds and a picture, in their order. */
export async function fetchCategoryTiles(kind: MediaKind): Promise<CategoryTileRow[]> {
	const data = await gqlClient.query({ categoryTiles: { __args: { kind }, ...TILE } });
	return data.categoryTiles;
}

/** Sorts tiles into top-level groups, in their order; children keep theirs. */
export function groupTiles(tiles: CategoryTileRow[]): TileGroup[] {
	const groups = tiles
		.filter((tile) => tile.category.parent === null)
		.map((parent) => ({ parent, children: [] as CategoryTileRow[] }));
	const byId = new Map(groups.map((group) => [group.parent.category.id, group]));
	for (const tile of tiles) {
		const parentId = tile.category.parent?.id;
		if (!parentId) continue;
		byId.get(parentId)?.children.push(tile);
	}
	return groups;
}

/** One category's page. */
export async function fetchCategory(id: string): Promise<CategoryDetail | null> {
	const data = await gqlClient.query({ category: { __args: { id }, ...CATEGORY_DETAIL } });
	return data.category ?? null;
}

/** The categories that can hold others: those at the top. */
export async function fetchParentOptions(): Promise<ParentOption[]> {
	const data = await gqlClient.query({ categories: PARENT_OPTION });
	return data.categories.filter((category) => category.parent === null);
}

/** Adds a category. */
export async function createCategory(input: CategoryInput): Promise<CategoryDetail> {
	const data = await gqlClient.mutation({
		createCategory: { __args: { input }, ...CATEGORY_DETAIL }
	});
	return data.createCategory;
}

/** Changes a category. */
export async function editCategory(id: string, input: CategoryInput): Promise<CategoryDetail> {
	const data = await gqlClient.mutation({
		editCategory: { __args: { id, input }, ...CATEGORY_DETAIL }
	});
	return data.editCategory;
}

/** Removes a category for good; its subcategories move up. */
export async function deleteCategory(id: string): Promise<void> {
	await gqlClient.mutation({ deleteCategory: { __args: { id } } });
}

/** Adds back the default categories the library lacks; returns how many. */
export async function addDefaultCategories(): Promise<number> {
	const data = await gqlClient.mutation({ addDefaultCategories: true });
	return data.addDefaultCategories;
}

/** A kind of media a category page or the categories page shows. */
export interface CategoryKind {
	kind: MediaKind;
	label: string;
	// What's counted, by name.
	noun: string;
}

/** The kinds categories show, in tab order. */
export const CATEGORY_KINDS: CategoryKind[] = [
	{ kind: 'video', label: 'Videos', noun: 'videos' },
	{ kind: 'gallery', label: 'Galleries', noun: 'galleries' },
	{ kind: 'audio', label: 'Audio', noun: 'audio' },
	{ kind: 'story', label: 'Stories', noun: 'stories' }
];

/** The kind named in a URL's ?kind=, videos when it names none of them. */
export function categoryKindOf(param: string | null): CategoryKind {
	return CATEGORY_KINDS.find((entry) => entry.kind === param) ?? CATEGORY_KINDS[0];
}

/** Terms typed as a comma-separated list. */
export function parseTerms(text: string): string[] {
	return text
		.split(',')
		.map((term) => term.trim())
		.filter((term) => term !== '');
}
