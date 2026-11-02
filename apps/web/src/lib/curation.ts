// The user's own organisation: categories, collections and the built-in
// watchlist.
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';

/** A category as a tile shows it. */
export interface CategoryCardRow {
	id: string;
	name: string;
	cover: { uri: string } | null;
	videoCount: number;
	galleryCount: number;
	parent: { id: string } | null;
}

/** A collection as a tile shows it, with the covers of its first items. */
export interface CollectionCardRow {
	id: string;
	name: string;
	cover: { uri: string } | null;
	itemCount: number;
	covers: string[];
}

// Covers stacked on a collection's tile.
const STACKED_COVERS = 3;

/** The top-level categories, in their order. */
export async function fetchCategories(): Promise<CategoryCardRow[]> {
	const data = await gqlClient.query({
		categories: {
			id: true,
			name: true,
			cover: { uri: true },
			videoCount: true,
			galleryCount: true,
			parent: { id: true }
		}
	});
	return data.categories.filter((category) => category.parent === null);
}

/** The user's own collections, with their first covers. */
export async function fetchCollections(): Promise<CollectionCardRow[]> {
	const data = await gqlClient.query({
		collections: {
			id: true,
			name: true,
			kind: true,
			cover: { uri: true },
			itemCount: true,
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
	const data = await gqlClient.query({ builtInCollection: { __args: { kind: 'watchlist' }, id: true } });
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
