import type {
	FieldsSelection,
	Taste as TasteResult,
	TasteEntry as TasteEntryResult
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** One liked (weight > 0) or disliked (< 0) entity; weight is in [-1, 1]. */
const ENTRY = { id: true, name: true, image: { uri: true }, weight: true } as const;

/** What the user likes and dislikes, per dimension: liked first, then disliked. */
const TASTE = {
	signalCount: true,
	tags: ENTRY,
	performers: ENTRY,
	studios: ENTRY,
	sites: ENTRY
} as const;

export type TasteEntry = FieldsSelection<TasteEntryResult, typeof ENTRY>;
export type Taste = FieldsSelection<TasteResult, typeof TASTE>;

/** The user's taste, from the event log. */
export async function fetchTaste(): Promise<Taste> {
	const data = await gqlClient.query({ taste: TASTE });
	return data.taste;
}
