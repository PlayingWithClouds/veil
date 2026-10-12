// Manual merge of two records that are the same performer, studio or tag.
import { gqlClient } from '$lib/veil';

export type MergeKind = 'performer' | 'studio' | 'tag';

export interface MergeCandidate {
	id: string;
	name: string;
}

/** Most candidates shown while picking the record to keep. */
const CANDIDATE_LIMIT = 8;

/** Records of `kind` whose name matches `search`, for picking the one to keep. */
export async function searchMergeCandidates(
	kind: MergeKind,
	search: string
): Promise<MergeCandidate[]> {
	const args = { __args: { search, limit: CANDIDATE_LIMIT }, id: true, name: true };
	if (kind === 'performer') {
		const result = await gqlClient.query({ performers: args });
		return (result.performers ?? []) as MergeCandidate[];
	}
	if (kind === 'studio') {
		const result = await gqlClient.query({ studios: args });
		return (result.studios ?? []) as MergeCandidate[];
	}
	const result = await gqlClient.query({ tags: args });
	return (result.tags ?? []) as MergeCandidate[];
}

/** Folds `mergedId` into `survivorId`; returns the survivor's id. Irreversible. */
export async function mergeRecords(survivorId: string, mergedId: string): Promise<string> {
	const result = await gqlClient.mutation({ mergeRecords: { __args: { survivorId, mergedId } } });
	return result.mergeRecords as string;
}
