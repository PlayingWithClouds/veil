import { browser } from '$app/environment';
import { writable } from 'svelte/store';

/** Which kinds of media Home shows: videos, images (galleries), or both. */
export type MediaFilter = 'all' | 'videos' | 'images';

const MEDIA_FILTER_STORAGE_KEY = 'veil:media-filter';

/** The filter chosen last time, defaulting to both. */
function loadMediaFilter(): MediaFilter {
	if (!browser) return 'all';
	const stored = localStorage.getItem(MEDIA_FILTER_STORAGE_KEY);
	if (stored === 'videos' || stored === 'images') return stored;
	return 'all';
}

// Set from the top bar's toggle; Home's feed and search results follow it.
export const mediaFilter = writable<MediaFilter>(loadMediaFilter());

mediaFilter.subscribe((filter) => {
	if (!browser) return;
	localStorage.setItem(MEDIA_FILTER_STORAGE_KEY, filter);
});
