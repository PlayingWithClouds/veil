import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, MEMBER, type VideoCardRow, type MemberRow } from '$lib/library';

/** A recommended video and why. */
export interface RecommendedVideo {
	video: VideoCardRow;
	reason: string | null;
}

/** A saved filter as its chip shows it. */
export interface SavedFilterChip {
	id: string;
	name: string;
}

/** A saved filter with the filter it plays. */
export interface SavedFilterDetail extends SavedFilterChip {
	filter: FilterInput;
}

/** Videos the user stopped partway, most recently watched first. */
export async function fetchContinueWatching(limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		history: {
			__args: { kind: 'video', limit: limit * 2 },
			progress: true,
			record: { on_Video: { __typename: true, ...VIDEO_CARD } }
		}
	});
	const started = data.history.flatMap((state) => {
		if (!state.progress || state.record?.__typename !== 'Video') return [];
		return [state.record];
	});
	return started.slice(0, limit);
}

/** Videos watched to the end before, most recently watched first. */
export async function fetchWatchAgain(limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		history: {
			__args: { kind: 'video', limit: limit * 2 },
			progress: true,
			record: { on_Video: { __typename: true, ...VIDEO_CARD } }
		}
	});
	const finished = data.history.flatMap((state) => {
		if (state.progress || state.record?.__typename !== 'Video') return [];
		return [state.record];
	});
	return finished.slice(0, limit);
}

/** Recommended videos with their first reason, only those filter matches. */
export async function fetchRecommendedVideos(
	limit = 20,
	filter: FilterInput = {}
): Promise<RecommendedVideo[]> {
	const data = await gqlClient.query({
		recommendations: {
			__args: { limit, filter },
			record: { on_Video: { __typename: true, ...VIDEO_CARD } },
			evidence: { detail: true }
		}
	});
	return data.recommendations.flatMap((match) => {
		if (match.record?.__typename !== 'Video') return [];
		return [{ video: match.record, reason: match.evidence[0]?.detail ?? null }];
	});
}

/** New videos from every follow, without repeats. */
export async function fetchFollowedVideos(perFollow = 10): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({ follows: { id: true, enabled: true } });
	const enabled = data.follows.filter((follow) => follow.enabled);
	const pages = await Promise.all(
		enabled.map((follow) =>
			gqlClient.query({
				newSince: {
					__args: { followId: follow.id, limit: perFollow },
					on_Video: { __typename: true, ...VIDEO_CARD }
				}
			})
		)
	);
	const seen = new Set<string>();
	return pages.flatMap((page) =>
		page.newSince.flatMap((record) => {
			if (record.__typename !== 'Video' || seen.has(record.id)) return [];
			seen.add(record.id);
			return [record];
		})
	);
}

/** The user's favourite performers, most recently marked first. */
export async function fetchFavoritePerformers(limit = 30): Promise<MemberRow[]> {
	const data = await gqlClient.query({
		userStates: {
			__args: { filter: { kind: 'performer', favorite: true }, limit },
			record: { on_Performer: { __typename: true, ...MEMBER } }
		}
	});
	return data.userStates.flatMap((state) => {
		if (state.record?.__typename !== 'Performer') return [];
		return [state.record];
	});
}

/** The saved filters pinned to home. */
export async function fetchPinnedFilters(): Promise<SavedFilterChip[]> {
	const data = await gqlClient.query({ savedFilters: { id: true, name: true, pinned: true } });
	return data.savedFilters.filter((saved) => saved.pinned);
}

/** One saved filter with the filter it plays. */
export async function fetchSavedFilter(id: string): Promise<SavedFilterDetail | null> {
	const data = await gqlClient.query({
		savedFilter: {
			__args: { id },
			id: true,
			name: true,
			filter: {
				kinds: true,
				text: true,
				performerIds: true,
				studioIds: true,
				categoryIds: true,
				includeTagIds: true,
				excludeTagIds: true,
				sources: true,
				dateFrom: true,
				dateTo: true,
				minDurationSeconds: true,
				maxDurationSeconds: true,
				minRating: true,
				status: true,
				inLibrary: true,
				favorite: true,
				unwatched: true,
				vr: true,
				fields: { fieldDefId: true, operator: true, value: true },
				sort: { field: true, descending: true }
			}
		}
	});
	return data.savedFilter ?? null;
}
