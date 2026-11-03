// Galleries beyond the plain library list: the overview's rows, the endless
// recommended feed and the galleries related to one.
import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { GALLERY_CARD, type GalleryCardRow } from '$lib/library';
import { listingFilter, type MediaListing } from '$lib/media';
import { endlessFeed, type Feed, type RankedPage } from '$lib/feed';

/** Recommendations ranked among galleries only. */
const GALLERIES_ONLY: FilterInput = { kinds: ['gallery'] };

/** Galleries, most recently added first (or as listing says). */
export async function fetchGalleryListing(
	limit = 20,
	offset = 0,
	listing: MediaListing = {}
): Promise<GalleryCardRow[]> {
	const data = await gqlClient.query({
		galleries: { __args: { filter: listingFilter(listing), limit, offset }, ...GALLERY_CARD }
	});
	return data.galleries;
}

/** Galleries the user opened, most recently viewed first. */
export async function fetchRecentlyViewedGalleries(limit = 20): Promise<GalleryCardRow[]> {
	const data = await gqlClient.query({
		history: {
			__args: { kind: 'gallery', limit },
			record: { on_Gallery: { __typename: true, ...GALLERY_CARD } }
		}
	});
	return data.history.flatMap((state) => {
		if (state.record?.__typename !== 'Gallery') return [];
		return [state.record];
	});
}

/** The favourite galleries, most recently added first. */
export async function fetchFavoriteGalleries(limit = 20): Promise<GalleryCardRow[]> {
	const data = await gqlClient.query({
		galleries: { __args: { filter: { favorite: true }, limit }, ...GALLERY_CARD }
	});
	return data.galleries;
}

/** One page of recommended galleries. */
async function fetchRecommendedGalleries(
	limit: number,
	offset: number
): Promise<RankedPage<GalleryCardRow>> {
	const data = await gqlClient.query({
		recommendations: {
			__args: { filter: GALLERIES_ONLY, limit, offset },
			record: { on_Gallery: { __typename: true, ...GALLERY_CARD } }
		}
	});
	const items = data.recommendations.flatMap((match) => {
		if (match.record?.__typename !== 'Gallery') return [];
		return [match.record];
	});
	return { items, entryCount: data.recommendations.length };
}

/** Galleries like one, most alike first: same performers, studio, rare tags, similar images. */
export async function fetchRelatedGalleries(id: string, limit = 24): Promise<GalleryCardRow[]> {
	const data = await gqlClient.query({
		related: {
			__args: { subject: { kind: 'gallery', id }, limit },
			record: { on_Gallery: { __typename: true, ...GALLERY_CARD } }
		}
	});
	return data.related.flatMap((match) => {
		if (match.record?.__typename !== 'Gallery') return [];
		return [match.record];
	});
}

/** The endless gallery feed: recommended first, then the rest of the library, newest first. */
export function galleryFeed(leaveOut: string[] = []): Feed<GalleryCardRow> {
	return endlessFeed(
		{ ranked: fetchRecommendedGalleries, rest: (limit, offset) => fetchGalleryListing(limit, offset) },
		leaveOut
	);
}
