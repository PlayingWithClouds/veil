// O-counter. Each increment records an event (history-backed) and
// returns the new total; decrement removes the most recent event. Feeds the
// recommender as the strongest taste signal.
import { gqlClient } from '$lib/veil';

export async function fetchOCount(mediaId: string): Promise<number> {
	try {
		const data = await gqlClient.query({
			oCount: { __args: { mediaId } }
		});
		return (data.oCount as number | null) ?? 0;
	} catch {
		return 0;
	}
}

export async function incrementOCount(mediaId: string): Promise<number> {
	const data = await gqlClient.mutation({
		incrementOCount: { __args: { mediaId } }
	});
	return (data.incrementOCount as number | null) ?? 0;
}

export async function decrementOCount(mediaId: string): Promise<number> {
	const data = await gqlClient.mutation({
		decrementOCount: { __args: { mediaId } }
	});
	return (data.decrementOCount as number | null) ?? 0;
}
