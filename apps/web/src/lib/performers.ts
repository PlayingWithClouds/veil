// A performer's page: reading them from their sites, the user's edits,
// probable duplicates and merging.
import { gqlClient } from '$lib/veil';
import { MEMBER, type MemberRow, type NamedRef, type SourceRow } from '$lib/library';

/** A performer's body and career facts; null where nothing is known. */
export interface PerformerAttributes {
	gender: string | null;
	birthdate: string | null;
	deathDate: string | null;
	country: string | null;
	ethnicity: string | null;
	eyeColor: string | null;
	hairColor: string | null;
	heightCm: number | null;
	weightKg: number | null;
	measurements: string | null;
	fakeTits: string | null;
	tattoos: string | null;
	piercings: string | null;
	careerStart: number | null;
	careerEnd: number | null;
}

/** Everything the performer page shows and edits. */
export interface PerformerDetail extends MemberRow {
	aliases: string[];
	details: string | null;
	links: string[];
	attributes: PerformerAttributes;
	tags: NamedRef[];
	sources: SourceRow[];
	// Which fields the user set; sites don't change those until unlocked.
	fieldSources: { field: string; locked: boolean }[];
}

/** The user's change to a performer; omitted fields stay as they are. */
export interface PerformerEdit {
	name?: string;
	aliases?: string[];
	details?: string;
	links?: string[];
	// Replaces every attribute.
	attributes?: PerformerAttributes;
	// "" clears the picture.
	imageUrl?: string;
}

const PERFORMER_DETAIL = {
	...MEMBER,
	aliases: true,
	details: true,
	links: true,
	attributes: {
		gender: true,
		birthdate: true,
		deathDate: true,
		country: true,
		ethnicity: true,
		eyeColor: true,
		hairColor: true,
		heightCm: true,
		weightKg: true,
		measurements: true,
		fakeTits: true,
		tattoos: true,
		piercings: true,
		careerStart: true,
		careerEnd: true
	},
	tags: { id: true, name: true },
	sources: { plugin: true, url: true, fetchedAt: true },
	fieldSources: { field: true, locked: true }
} as const;

/** One performer as stored, without reading any site. */
export async function fetchPerformer(id: string): Promise<PerformerDetail | null> {
	const data = await gqlClient.query({ performer: { __args: { id }, ...PERFORMER_DETAIL } });
	return data.performer ?? null;
}

/**
 * Reads the performer's pages on the sites they came from (the unread ones,
 * or all of them with refresh) and returns them as stored after.
 */
export async function readPerformer(id: string, refresh = false): Promise<PerformerDetail | null> {
	const data = await gqlClient.mutation({
		fetchPerformer: { __args: { id, refresh }, ...PERFORMER_DETAIL }
	});
	return data.fetchPerformer ?? null;
}

/** Saves the user's change; every field it sets is locked against sites. */
export async function editPerformer(
	id: string,
	edit: PerformerEdit
): Promise<PerformerDetail | null> {
	const data = await gqlClient.mutation({
		editPerformer: { __args: { id, edit }, ...PERFORMER_DETAIL }
	});
	return data.editPerformer ?? null;
}

/** Releases a field the user set, so the next site that describes it takes it over. */
export async function unlockPerformerField(id: string, field: string): Promise<void> {
	await gqlClient.mutation({
		lockField: {
			__args: { subject: { kind: 'performer', id }, field, locked: false },
			field: true
		}
	});
}

/** Performers that are probably the same person: a name or alias in common. */
export async function fetchPerformerDuplicates(id: string): Promise<MemberRow[]> {
	const data = await gqlClient.query({ performerDuplicates: { __args: { id }, ...MEMBER } });
	return data.performerDuplicates;
}

/**
 * Folds mergeId into keepId for good: its sources, credits and the user's
 * state move over, its name becomes an alias, and it is gone.
 */
export async function mergePerformers(keepId: string, mergeId: string): Promise<void> {
	await gqlClient.mutation({
		merge: {
			__args: { kind: 'performer', keepId, mergeId },
			on_Performer: { id: true }
		}
	});
}
