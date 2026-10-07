import { gqlClient } from '$lib/veil';

export interface PerformerInfo {
	id: string;
	name: string;
	aliases: string[];
	details: string | null;
	gender: string | null;
	birthdate: string | null;
	country: string | null;
	ethnicity: string | null;
	eyeColor: string | null;
	hairColor: string | null;
	heightCm: number | null;
	weightKg: number | null;
	measurements: string | null;
	careerLength: string | null;
	url: string | null;
	imagePath: string | null;
	favorite: boolean;
	sceneCount: number;
	tags: { id: string; name: string }[];
}

export interface PerformerSceneCard {
	id: string;
	title: string;
	posterPath: string | null;
	date: string | null;
	durationSeconds: number | null;
}

export interface PerformerDetail {
	performer: PerformerInfo | null;
	scenes: PerformerSceneCard[];
}

const EMPTY_DETAIL: PerformerDetail = { performer: null, scenes: [] };

export interface PerformerCard {
	id: string;
	name: string;
	imagePath: string | null;
	birthdate: string | null;
	sceneCount: number;
	favorite: boolean;
}

// Lists performers for the index grid (ordered by name server-side).
export async function loadPerformers(limit = 120): Promise<PerformerCard[]> {
	try {
		const data = await gqlClient.query({
			performers: {
				__args: { limit },
				id: true,
				name: true,
				imagePath: true,
				birthdate: true,
				sceneCount: true,
				favorite: true
			}
		});
		return (data.performers ?? []) as unknown as PerformerCard[];
	} catch {
		return [];
	}
}

// Loads a performer's profile plus the scenes they appear in.
export async function loadPerformerDetail(id: string): Promise<PerformerDetail> {
	try {
		const [performerData, scenesData] = await Promise.all([
			gqlClient.query({
				performer: {
					__args: { id },
					id: true,
					name: true,
					aliases: true,
					details: true,
					gender: true,
					birthdate: true,
					country: true,
					ethnicity: true,
					eyeColor: true,
					hairColor: true,
					heightCm: true,
					weightKg: true,
					measurements: true,
					careerLength: true,
					url: true,
					imagePath: true,
					favorite: true,
					sceneCount: true,
					tags: { id: true, name: true }
				}
			}),
			gqlClient.query({
				scenes: {
					__args: { performerId: id, limit: 60 },
					id: true,
					title: true,
					posterPath: true,
					date: true,
					durationSeconds: true
				}
			})
		]);
		if (!performerData.performer) return EMPTY_DETAIL;
		return {
			performer: performerData.performer as unknown as PerformerInfo,
			scenes: (scenesData.scenes ?? []) as unknown as PerformerSceneCard[]
		};
	} catch {
		return EMPTY_DETAIL;
	}
}

// Toggle a performer as a favorite.
export async function setPerformerFavorite(id: string, favorite: boolean): Promise<void> {
	await gqlClient.mutation({
		setPerformerFavorite: { __args: { id, favorite }, id: true }
	});
}

// Queue an enrich job to fill in the performer's photo, measurements and bio.
export async function enrichPerformer(id: string): Promise<void> {
	await gqlClient.mutation({ enrichPerformer: { __args: { id } } });
}
