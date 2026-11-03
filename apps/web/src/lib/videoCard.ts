// What a video card shows beyond its row: the quality label and the actions
// in its ⋮ menu.
import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
import type { SheetAction } from '$lib/components/ActionSheet.svelte';
import type { VideoCardRow } from '$lib/library';
import { addToWatchlist } from '$lib/curation';
import { recordActions } from '$lib/recordActions';
import { showUndo } from '$lib/stores/undoToast';

// Frame heights (the shorter side) and their labels, best first.
const QUALITY_STEPS: [number, string][] = [
	[2160, '4K'],
	[1440, '1440p'],
	[1080, '1080p'],
	[720, '720p']
];

/** The quality badge for a frame size, e.g. "1080p"; null below 720p or when unknown. */
export function qualityLabel(width: number | null, height: number | null): string | null {
	if (!width || !height) return null;
	const shorterSide = Math.min(width, height);
	// Encodes cropped a little short of the step still count as it (1080×1072).
	const step = QUALITY_STEPS.find(([minimum]) => shorterSide >= minimum * 0.95);
	if (!step) return null;
	return step[1];
}

/** What a card's menu does beyond the menu itself. */
export interface VideoCardHooks {
	// Called after favourite changes, with the new value.
	onfavorite: (favorite: boolean) => void;
}

/** The ⋮ menu of a video card. */
export function videoCardActions(video: VideoCardRow, hooks: VideoCardHooks): SheetAction[] {
	const record = {
		kind: 'video' as const,
		id: video.id,
		favorite: video.userState.favorite,
		performers: video.performers,
		studio: video.studio,
		sourceUrl: video.sources[0]?.url ?? null
	};
	return [
		{ label: 'Add to watchlist', icon: BookmarkSimpleIcon, run: () => void watchlist(video) },
		...recordActions(record, hooks.onfavorite)
	];
}

/** Puts the video on the watchlist and says so. */
async function watchlist(video: VideoCardRow): Promise<void> {
	await addToWatchlist(video.id);
	showUndo({ message: 'Added to watchlist' });
}
