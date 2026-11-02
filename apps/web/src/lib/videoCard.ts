// What a video card shows beyond its row: the quality label and the actions
// in its ⋮ menu.
import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
import UserMinusIcon from 'phosphor-svelte/lib/UserMinusIcon';
import BuildingsIcon from 'phosphor-svelte/lib/BuildingsIcon';
import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
import type { RecordKind } from '@playingwithclouds/veil-sdk/generated';
import type { SheetAction } from '$lib/components/ActionSheet.svelte';
import type { VideoCardRow } from '$lib/library';
import { addToWatchlist } from '$lib/curation';
import { setBlocked, setFavorite, setHidden, logNotInterested } from '$lib/userState';
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

// Blockable credits listed in the menu; more performers are blocked from their page.
const MENU_PERFORMERS = 3;

/** The ⋮ menu of a video card. */
export function videoCardActions(video: VideoCardRow, hooks: VideoCardHooks): SheetAction[] {
	const favorite = video.userState.favorite;
	const actions: SheetAction[] = [
		{ label: 'Add to watchlist', icon: BookmarkSimpleIcon, run: () => void watchlist(video) },
		{
			label: favorite ? 'Remove from favourites' : 'Add to favourites',
			icon: HeartIcon,
			run: () => void toggleFavorite(video, hooks)
		},
		{ label: 'Not interested', icon: EyeSlashIcon, run: () => notInterested(video) },
		{
			label: 'Block this video',
			icon: ProhibitIcon,
			run: () => block('video', video.id, 'Video'),
			destructive: true
		}
	];
	for (const performer of video.performers.slice(0, MENU_PERFORMERS)) {
		actions.push({
			label: `Block ${performer.name}`,
			icon: UserMinusIcon,
			run: () => block('performer', performer.id, performer.name),
			destructive: true
		});
	}
	if (video.studio) {
		const studio = video.studio;
		actions.push({
			label: `Block ${studio.name}`,
			icon: BuildingsIcon,
			run: () => block('studio', studio.id, studio.name),
			destructive: true
		});
	}
	const sourceUrl = video.sources[0]?.url;
	if (sourceUrl) {
		actions.push({
			label: 'Open on the site',
			icon: ArrowSquareOutIcon,
			run: () => window.open(sourceUrl, '_blank', 'noopener')
		});
	}
	return actions;
}

/** Puts the video on the watchlist and says so. */
async function watchlist(video: VideoCardRow): Promise<void> {
	await addToWatchlist(video.id);
	showUndo({ message: 'Added to watchlist' });
}

/** Flips the video's favourite mark. */
async function toggleFavorite(video: VideoCardRow, hooks: VideoCardHooks): Promise<void> {
	const favorite = !video.userState.favorite;
	await setFavorite('video', video.id, favorite);
	hooks.onfavorite(favorite);
}

/** Hides the video now; the thumbs down is sent once Undo has passed. */
function notInterested(video: VideoCardRow): void {
	setHidden('video', video.id, true);
	showUndo({
		message: 'Video hidden. You won’t be recommended it again.',
		undo: () => setHidden('video', video.id, false),
		settle: () => logNotInterested('video', video.id)
	});
}

/** Blocks a record, with Undo. */
function block(kind: RecordKind, id: string, name: string): void {
	void setBlocked(kind, id, true);
	showUndo({ message: `Blocked ${name}`, undo: () => void setBlocked(kind, id, false) });
}
