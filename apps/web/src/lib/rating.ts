// Like/dislike, stored on the 1–10 user_rating scale. Thumbs map to
// the extremes so the recommendation taste signal (centered at 5) gets a strong
// positive or negative pull.
import { gqlClient } from '$lib/veil';

export type Verdict = 'up' | 'down' | null;

export const RATING_UP = 10;
export const RATING_DOWN = 1;

// Interprets a stored numeric rating as a thumbs verdict.
export function ratingToVerdict(rating: number | null): Verdict {
	if (rating === null) return null;
	if (rating > 5) return 'up';
	if (rating < 5) return 'down';
	return null;
}

export async function fetchVerdict(mediaId: string): Promise<Verdict> {
	try {
		const data = await gqlClient.query({
			userRating: { __args: { mediaId }, rating: true }
		});
		const rating = data.userRating?.rating ?? null;
		return ratingToVerdict(rating as number | null);
	} catch {
		return null;
	}
}

export async function setVerdict(mediaId: string, verdict: Exclude<Verdict, null>): Promise<void> {
	const rating = verdict === 'up' ? RATING_UP : RATING_DOWN;
	await gqlClient.mutation({
		upsertUserRating: {
			__args: { input: { media: mediaId, rating } },
			rating: true
		}
	});
}

export async function clearVerdict(mediaId: string): Promise<void> {
	await gqlClient.mutation({
		deleteUserRating: { __args: { mediaId } }
	});
}
