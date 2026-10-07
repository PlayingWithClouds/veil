import { gqlClient } from '$lib/veil';

export interface StudioInfo {
	id: string;
	name: string;
	aliases: string[];
	url: string | null;
	details: string | null;
	imagePath: string | null;
	sceneCount: number;
	parent: { id: string; name: string } | null;
	tags: { id: string; name: string }[];
}

export interface StudioSceneCard {
	id: string;
	title: string;
	posterPath: string | null;
	date: string | null;
	durationSeconds: number | null;
}

export interface StudioDetail {
	studio: StudioInfo | null;
	scenes: StudioSceneCard[];
}

const EMPTY_DETAIL: StudioDetail = { studio: null, scenes: [] };

export interface StudioCard {
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
}

// Lists studios for the index grid (ordered by name server-side).
export async function loadStudios(limit = 120): Promise<StudioCard[]> {
	try {
		const data = await gqlClient.query({
			studios: {
				__args: { limit },
				id: true,
				name: true,
				imagePath: true,
				sceneCount: true
			}
		});
		return (data.studios ?? []) as unknown as StudioCard[];
	} catch {
		return [];
	}
}

// Loads a studio's profile plus the scenes it produced.
export async function loadStudioDetail(id: string): Promise<StudioDetail> {
	try {
		const [studioData, scenesData] = await Promise.all([
			gqlClient.query({
				studio: {
					__args: { id },
					id: true,
					name: true,
					aliases: true,
					url: true,
					details: true,
					imagePath: true,
					sceneCount: true,
					parent: { id: true, name: true },
					tags: { id: true, name: true }
				}
			}),
			gqlClient.query({
				scenes: {
					__args: { studioId: id, limit: 60 },
					id: true,
					title: true,
					posterPath: true,
					date: true,
					durationSeconds: true
				}
			})
		]);
		if (!studioData.studio) return EMPTY_DETAIL;
		return {
			studio: studioData.studio as unknown as StudioInfo,
			scenes: (scenesData.scenes ?? []) as unknown as StudioSceneCard[]
		};
	} catch {
		return EMPTY_DETAIL;
	}
}
