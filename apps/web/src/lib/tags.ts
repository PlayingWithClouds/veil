// Tag index + detail data. The tag detail page is the hub for everything a tag
// touches: scenes (direct and inherited via studio/performer), galleries,
// images, performers, studios, and collections.
import { gqlClient } from '$lib/veil';
import { sceneToCard, SCENE_ROW_FIELDS, type SceneRow } from '$lib/filters';
import type { SceneCardData } from '$lib/components/SceneCard.svelte';

export interface TagRef {
	id: string;
	name: string;
}

/**
 * A tag name in one consistent casing: lowercase, except short all-caps words
 * that are acronyms ("POV", "MILF", "JAV"). Sites disagree on casing, so the
 * stored spelling is whatever arrived first.
 */
export function tagLabel(name: string): string {
	return name
		.split(' ')
		.map((word) => {
			const isAcronym = word.length >= 2 && word.length <= 5 && word === word.toUpperCase() && /[A-Z]/.test(word);
			if (isAcronym) return word;
			return word.toLowerCase();
		})
		.join(' ');
}

export interface TagSummary extends TagRef {
	description: string | null;
	category: string | null;
	sceneCount: number;
}

export interface TagDetail extends TagSummary {
	aliases: string[];
}

const TAG_FIELDS = {
	id: true,
	name: true,
	description: true,
	category: true,
	sceneCount: true
} as const;

export async function fetchTagIndex(search?: string): Promise<TagSummary[]> {
	try {
		const data = await gqlClient.query({
			tags: { __args: { search: search || undefined }, ...TAG_FIELDS }
		});
		return (data.tags ?? []) as unknown as TagSummary[];
	} catch {
		return [];
	}
}

export async function fetchTag(id: string): Promise<TagDetail | null> {
	try {
		const data = await gqlClient.query({
			tag: { __args: { id }, ...TAG_FIELDS, aliases: true }
		});
		return (data.tag ?? null) as unknown as TagDetail | null;
	} catch {
		return null;
	}
}

export interface TaggedScenes {
	direct: SceneCardData[];
	inherited: SceneCardData[];
}

export interface TaggedPerformer {
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
}

export interface TaggedStudio {
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
}

export interface TaggedGallery {
	id: string;
	title: string;
	coverPath: string | null;
	imageCount: number;
}

export interface TaggedCollection {
	id: string;
	name: string;
	coverPath: string | null;
	itemCount: number;
}

export async function fetchPerformersForTag(tagId: string, limit = 30): Promise<TaggedPerformer[]> {
	try {
		const data = await gqlClient.query({
			performers: {
				__args: { tagId, limit },
				id: true,
				name: true,
				imagePath: true,
				sceneCount: true
			}
		});
		return (data.performers ?? []) as unknown as TaggedPerformer[];
	} catch {
		return [];
	}
}

export async function fetchStudiosForTag(tagId: string, limit = 30): Promise<TaggedStudio[]> {
	try {
		const data = await gqlClient.query({
			studios: {
				__args: { tagId, limit },
				id: true,
				name: true,
				imagePath: true,
				sceneCount: true
			}
		});
		return (data.studios ?? []) as unknown as TaggedStudio[];
	} catch {
		return [];
	}
}

export async function fetchGalleriesForTag(tagId: string, limit = 30): Promise<TaggedGallery[]> {
	try {
		const data = await gqlClient.query({
			galleries: {
				__args: { tagId, limit },
				id: true,
				title: true,
				coverPath: true,
				imageCount: true
			}
		});
		return (data.galleries ?? []) as unknown as TaggedGallery[];
	} catch {
		return [];
	}
}

export async function fetchCollectionsForTag(
	tagId: string,
	limit = 30
): Promise<TaggedCollection[]> {
	try {
		const data = await gqlClient.query({
			collections: {
				__args: { tagId, limit },
				id: true,
				name: true,
				coverPath: true,
				itemCount: true
			}
		});
		return (data.collections ?? []) as unknown as TaggedCollection[];
	} catch {
		return [];
	}
}

// Scenes matching a tag, split into direct matches and ones inherited from the
// scene's studio or performers. The backend ranks direct first and reports the
// split via tagMatch.
export async function fetchScenesForTag(tagId: string, limit = 60): Promise<TaggedScenes> {
	try {
		const data = await gqlClient.query({
			scenes: {
				__args: { tagId, limit },
				...SCENE_ROW_FIELDS
			}
		});
		const rows = (data.scenes ?? []) as unknown as (SceneRow & { tagMatch: string | null })[];
		const direct: SceneCardData[] = [];
		const inherited: SceneCardData[] = [];
		for (const row of rows) {
			if (row.tagMatch === 'inherited') {
				inherited.push(sceneToCard(row));
			} else {
				direct.push(sceneToCard(row));
			}
		}
		return { direct, inherited };
	} catch {
		return { direct: [], inherited: [] };
	}
}
