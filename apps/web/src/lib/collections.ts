import { gqlClient } from '$lib/veil';
import { sceneUrl, galleryUrl, performerUrl, studioUrl } from '$lib/routes';
import type { TagRef } from '$lib/tags';

// Entity types a collection can contain. Collections cannot contain collections.
export type CollectionEntityType = 'scene' | 'gallery' | 'image' | 'performer' | 'studio';

export interface Collection {
	id: string;
	name: string;
	details: string | null;
	itemCount: number;
	coverPath: string | null;
	origin: 'user' | 'scraped';
	tags: TagRef[];
	createdAt: string;
	updatedAt: string;
}

export interface CollectionMember {
	mediaId: string;
	mediaType: CollectionEntityType;
	position: number;
	title: string;
	posterPath: string | null;
}

const COLLECTION_FIELDS = {
	id: true,
	name: true,
	details: true,
	itemCount: true,
	coverPath: true,
	origin: true,
	tags: { id: true, name: true },
	createdAt: true,
	updatedAt: true
} as const;

export async function fetchCollections(origin?: 'user' | 'scraped'): Promise<Collection[]> {
	try {
		const data = await gqlClient.query({
			collections: { __args: { origin }, ...COLLECTION_FIELDS }
		});
		return (data.collections ?? []) as unknown as Collection[];
	} catch {
		return [];
	}
}

export async function fetchCollection(id: string): Promise<Collection | null> {
	try {
		const data = await gqlClient.query({
			collection: { __args: { id }, ...COLLECTION_FIELDS }
		});
		return (data.collection ?? null) as unknown as Collection | null;
	} catch {
		return null;
	}
}

// Members in manual sort order, shaped as cards for mixed-type rendering.
export async function fetchCollectionMembers(collectionId: string): Promise<CollectionMember[]> {
	try {
		const data = await gqlClient.query({
			collectionMembers: {
				__args: { collectionId },
				mediaId: true,
				mediaType: true,
				position: true,
				title: true,
				posterPath: true
			}
		});
		return (data.collectionMembers ?? []) as unknown as CollectionMember[];
	} catch {
		return [];
	}
}

// Ids of the user's own collections that already contain the given record.
export async function collectionIdsForEntity(mediaId: string): Promise<string[]> {
	try {
		const data = await gqlClient.query({
			collectionIdsForMedia: { __args: { mediaId } }
		});
		return (data.collectionIdsForMedia ?? []) as unknown as string[];
	} catch {
		return [];
	}
}

export async function createCollection(name: string): Promise<Collection | null> {
	try {
		const data = await gqlClient.mutation({
			createCollection: { __args: { name }, ...COLLECTION_FIELDS }
		});
		return (data.createCollection ?? null) as unknown as Collection | null;
	} catch {
		return null;
	}
}

export async function renameCollection(collectionId: string, name: string): Promise<void> {
	await gqlClient.mutation({ renameCollection: { __args: { collectionId, name }, id: true } });
}

export async function deleteCollection(collectionId: string): Promise<void> {
	await gqlClient.mutation({ deleteCollection: { __args: { collectionId } } });
}

export async function addToCollection(collectionId: string, mediaId: string): Promise<void> {
	await gqlClient.mutation({ addToCollection: { __args: { collectionId, mediaId } } });
}

export async function removeFromCollection(collectionId: string, mediaId: string): Promise<void> {
	await gqlClient.mutation({ removeFromCollection: { __args: { collectionId, mediaId } } });
}

export async function reorderCollection(collectionId: string, mediaIds: string[]): Promise<void> {
	await gqlClient.mutation({ reorderCollection: { __args: { collectionId, mediaIds } } });
}

export async function setCollectionTags(collectionId: string, tagIds: string[]): Promise<void> {
	await gqlClient.mutation({ setCollectionTags: { __args: { collectionId, tagIds }, id: true } });
}

// Route for a member card click, by entity type. Images have no detail route;
// they open in a lightbox instead.
export function memberUrl(member: CollectionMember): string {
	switch (member.mediaType) {
		case 'scene':
			return sceneUrl(member.mediaId);
		case 'gallery':
			return galleryUrl(member.mediaId);
		case 'performer':
			return performerUrl(member.mediaId);
		case 'studio':
			return studioUrl(member.mediaId);
		default:
			return '';
	}
}
