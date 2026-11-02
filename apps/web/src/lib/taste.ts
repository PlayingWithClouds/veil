import { gqlClient } from '$lib/veil';

/** One liked (weight > 0) or disliked (< 0) entity; weight is in [-1, 1]. */
export interface TasteEntry {
	id: string;
	name: string;
	image: { uri: string } | null;
	weight: number;
}

/** What the user likes and dislikes, per dimension: liked first, then disliked. */
export interface Taste {
	signalCount: number;
	tags: TasteEntry[];
	performers: TasteEntry[];
	studios: TasteEntry[];
	sites: TasteEntry[];
}

const ENTRY = { id: true, name: true, image: { uri: true }, weight: true } as const;

/** The user's taste, from the event log. */
export async function fetchTaste(): Promise<Taste> {
	const data = await gqlClient.query({
		taste: { signalCount: true, tags: ENTRY, performers: ENTRY, studios: ENTRY, sites: ENTRY }
	});
	return data.taste;
}
