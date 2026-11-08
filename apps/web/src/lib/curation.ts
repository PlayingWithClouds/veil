// The user's own organisation: collections and the built-in watchlist
// (categories are in categories.ts).
import type {
	Collection,
	CollectionKind,
	FieldsSelection
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { GALLERY_CARD, VIDEO_CARD, type GalleryCardRow, type VideoCardRow } from '$lib/library';

/** The collection fields its tile shows; the stacked covers come from its items. */
const COLLECTION_TILE = { id: true, name: true, cover: { uri: true }, itemCount: true } as const;

/** A collection as a tile shows it, with the covers of its first items. */
export type CollectionCardRow = FieldsSelection<Collection, typeof COLLECTION_TILE> & {
	covers: string[];
};

// Covers stacked on a collection's tile.
const STACKED_COVERS = 3;

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

/** What a collection's page shows of it, with its videos and galleries. */
const COLLECTION_PAGE = {
	id: true,
	name: true,
	kind: true,
	details: true,
	itemCount: true,
	items: {
		record: {
			on_Video: { __typename: true, ...VIDEO_CARD },
			on_Gallery: { __typename: true, ...GALLERY_CARD }
		}
	}
} as const;

/** One video or gallery in a collection; other kinds of records aren't shown. */
export type CollectionEntry =
	{ kind: 'video'; video: VideoCardRow } | { kind: 'gallery'; gallery: GalleryCardRow };

/** A collection with its videos and galleries, in its order. */
export interface CollectionPage {
	id: string;
	name: string;
	kind: CollectionKind;
	details: string | null;
	itemCount: number;
	entries: CollectionEntry[];
}

type CollectionPageRow = FieldsSelection<Collection, typeof COLLECTION_PAGE>;

/** A collection by id, with its items; null when it's gone. */
export async function fetchCollection(id: string): Promise<CollectionPage | null> {
	const data = await gqlClient.query({ collection: { __args: { id }, ...COLLECTION_PAGE } });
	if (!data.collection) return null;
	return toCollectionPage(data.collection);
}

/** The watchlist or the queue, with its items. */
export async function fetchBuiltInCollection(kind: CollectionKind): Promise<CollectionPage> {
	const data = await gqlClient.query({
		builtInCollection: { __args: { kind }, ...COLLECTION_PAGE }
	});
	return toCollectionPage(data.builtInCollection);
}

/** A collection as its page shows it: only the videos and galleries of its items. */
function toCollectionPage(collection: CollectionPageRow): CollectionPage {
	return {
		id: collection.id,
		name: collection.name,
		kind: collection.kind,
		details: collection.details ?? null,
		itemCount: collection.itemCount,
		entries: collection.items.flatMap((item): CollectionEntry[] => {
			const record = item.record;
			if (record?.__typename === 'Video') return [{ kind: 'video', video: record }];
			if (record?.__typename === 'Gallery') return [{ kind: 'gallery', gallery: record }];
			return [];
		})
	};
}

/** Takes one video or gallery out of a collection. */
export async function removeFromCollection(collectionId: string, entry: CollectionEntry) {
	await removeRecordFromCollection(collectionId, { kind: entry.kind, id: entryId(entry) });
}

/** The id of the record an entry shows. */
export function entryId(entry: CollectionEntry): string {
	if (entry.kind === 'video') return entry.video.id;
	return entry.gallery.id;
}

/** Creates an empty collection and returns its id. */
export async function createCollection(name: string): Promise<string> {
	const data = await gqlClient.mutation({
		createCollection: { __args: { name }, id: true }
	});
	return data.createCollection.id;
}

/** A video or gallery a collection can hold. */
export interface CollectableRef {
	kind: 'video' | 'gallery';
	id: string;
}

/** Adds a record to the end of a collection; one already in it moves there. */
export async function addToCollection(collectionId: string, subject: CollectableRef) {
	await gqlClient.mutation({
		addToCollection: { __args: { collectionId, subjects: [subject] }, id: true }
	});
}

/** Takes a record out of a collection. */
export async function removeRecordFromCollection(collectionId: string, subject: CollectableRef) {
	await gqlClient.mutation({
		removeFromCollection: { __args: { collectionId, subjects: [subject] }, id: true }
	});
}

/** The ids of the user's own collections a record is in. */
export async function fetchCollectionIdsWith(subject: CollectableRef): Promise<Set<string>> {
	const data = await gqlClient.query({
		collectionsWith: { __args: { subject }, id: true, kind: true }
	});
	return new Set(
		data.collectionsWith
			.filter((collection) => collection.kind === 'user')
			.map((collection) => collection.id)
	);
}

/** Puts a video first in the queue, so it plays next. */
export async function playNext(videoId: string): Promise<void> {
	const data = await gqlClient.query({
		builtInCollection: { __args: { kind: 'queue' }, id: true }
	});
	await gqlClient.mutation({
		addToCollection: {
			__args: {
				collectionId: data.builtInCollection.id,
				subjects: [{ kind: 'video', id: videoId }],
				position: 0
			},
			id: true
		}
	});
}
