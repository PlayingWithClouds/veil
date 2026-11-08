// "Up next" beside a video: an endless list of videos like it. First what the
// library relates to it, then what the sites list as related on its pages,
// then what the sites find for its performers and studio by name, then picks
// for the user. Nothing shows twice, the video itself never.
import { chainFeed, offsetSource, onceSource, type Feed, type Source } from '$lib/feed';
import { fetchRecommendedVideos } from '$lib/home';
import {
	fetchRelatedVideos,
	fetchSiteRelated,
	type VideoCardRow,
	type VideoDetail
} from '$lib/library';
import { collectSearchPage } from '$lib/search';

// Videos the library relates to the video, read at once.
const LIBRARY_RELATED = 30;
// Search pages read for one name before the next name's turn.
const PAGES_PER_NAME = 3;
// Picks for the user that end the list.
const PICKS = 120;

/**
 * The endless related list of a video. opened settles once its pages were
 * read, which stores what the sites list as related and names its credits.
 */
export function relatedFeed(
	videoId: string,
	opened: Promise<VideoDetail | null>
): Feed<VideoCardRow> {
	return chainFeed<VideoCardRow>(
		[
			onceSource(() => fetchRelatedVideos(videoId, LIBRARY_RELATED)),
			offsetSource(async (limit, offset) => {
				await opened.catch(() => null);
				return fetchSiteRelated(videoId, limit, offset);
			}),
			nameSearchSource(() => creditNames(opened)),
			onceSource(async () => {
				const picks = await fetchRecommendedVideos(PICKS);
				return picks.map((pick) => pick.video);
			})
		],
		[videoId]
	);
}

/** The video's performers' names, then its studio's, once it's opened. */
async function creditNames(opened: Promise<VideoDetail | null>): Promise<string[]> {
	const video = await opened.catch(() => null);
	if (!video) return [];
	const names = video.performers.map((performer) => performer.name);
	if (video.studio) names.push(video.studio.name);
	return names;
}

/**
 * Pages of a site search for each name in turn, quoted so only results
 * naming it come back; a name's turn ends at a page with nothing or after
 * PAGES_PER_NAME pages.
 */
function nameSearchSource(names: () => Promise<string[]>): Source<VideoCardRow> {
	let remaining: string[] | null = null;
	let page = 0;
	return async () => {
		if (remaining === null) remaining = await names();
		while (remaining.length > 0) {
			if (page < PAGES_PER_NAME) {
				const found = await collectSearchPage(`"${remaining[0]}"`, page);
				page++;
				if (found.length > 0) return found;
			}
			remaining.shift();
			page = 0;
		}
		return null;
	};
}
