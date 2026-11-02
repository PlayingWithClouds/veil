// The user's own organisation: categories, collections and the built-in
// watchlist.
import type { Category, Collection, FieldsSelection } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';

/** A category as a tile shows it. */
const CATEGORY_TILE = {
	id: true,
	name: true,
	cover: { uri: true },
	videoCount: true,
	galleryCount: true,
	parent: { id: true }
} as const;

/** The collection fields its tile shows; the stacked covers come from its items. */
const COLLECTION_TILE = { id: true, name: true, cover: { uri: true }, itemCount: true } as const;

export type CategoryCardRow = FieldsSelection<Category, typeof CATEGORY_TILE>;

/** A collection as a tile shows it, with the covers of its first items. */
export type CollectionCardRow = FieldsSelection<Collection, typeof COLLECTION_TILE> & {
	covers: string[];
};

// Covers stacked on a collection's tile.
const STACKED_COVERS = 3;

/** The top-level categories, in their order. */
export async function fetchCategories(): Promise<CategoryCardRow[]> {
	const data = await gqlClient.query({ categories: CATEGORY_TILE });
	return data.categories.filter((category) => category.parent === null);
}

/** The user's own collections, with their first covers. */
export async function fetchCollections(): Promise<CollectionCardRow[]> {
	const data = await gqlClient.query({
		collections: {
			...COLLECTION_TILE,
			kind: true,
			items: {
				__args: { limit: STACKED_COVERS },
				record: {
					on_Video: { __typename: true, cover: { uri: true } },
					on_Gallery: { __typename: true, cover: { uri: true } }
				}
			}
		}
	});
	return data.collections
		.filter((collection) => collection.kind === 'user')
		.map((collection) => ({
			id: collection.id,
			name: collection.name,
			cover: collection.cover,
			itemCount: collection.itemCount,
			covers: collection.items.flatMap((item) => {
				const record = item.record;
				if (record?.__typename !== 'Video' && record?.__typename !== 'Gallery') return [];
				if (!record.cover) return [];
				return [record.cover.uri];
			})
		}));
}

/** The videos on the watchlist, in its order. */
export async function fetchWatchlist(limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		builtInCollection: {
			__args: { kind: 'watchlist' },
			items: {
				__args: { limit },
				record: { on_Video: { __typename: true, ...VIDEO_CARD } }
			}
		}
	});
	return data.builtInCollection.items.flatMap((item) => {
		if (item.record?.__typename !== 'Video') return [];
		return [item.record];
	});
}

/** Puts a video on the watchlist. */
export async function addToWatchlist(videoId: string): Promise<void> {
	const data = await gqlClient.query({
		builtInCollection: { __args: { kind: 'watchlist' }, id: true }
	});
	await gqlClient.mutation({
		addToCollection: {
			__args: {
				collectionId: data.builtInCollection.id,
				subjects: [{ kind: 'video', id: videoId }]
			},
			id: true
		}
	});
}
