// The ⋮ menu entries video and gallery cards (and their pages) share: play
// next and add to a collection,
// favourite, not interested, blocking the record or what it credits, and
// opening it on the site.
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
import UserMinusIcon from 'phosphor-svelte/lib/UserMinusIcon';
import BuildingsIcon from 'phosphor-svelte/lib/BuildingsIcon';
import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
import type { RecordKind } from '@playingwithclouds/veil-sdk/generated';
import type { SheetAction } from '$lib/components/ActionSheet.svelte';
import { setBlocked, setFavorite, setHidden, logNotInterested } from '$lib/userState';
import QueueIcon from 'phosphor-svelte/lib/QueueIcon';
import StackPlusIcon from 'phosphor-svelte/lib/StackPlusIcon';
import { playNext } from '$lib/curation';
import { libraryUrl } from '$lib/routes';
import { pickCollectionFor } from '$lib/stores/collectionPicker';
import { showUndo } from '$lib/stores/undoToast';

// Blockable credits listed in the menu; more performers are blocked from their page.
const MENU_PERFORMERS = 3;

/** A media record as its menu needs it. */
export interface MenuRecord {
	kind: RecordKind;
	id: string;
	title: string;
	favorite: boolean;
	performers: { id: string; name: string }[];
	studio: { id: string; name: string } | null;
	sourceUrl: string | null;
}

/** The record's name in messages and labels: "Video", "Gallery". */
function nounOf(kind: RecordKind): string {
	return kind.charAt(0).toUpperCase() + kind.slice(1);
}

/**
 * Play next and add to a collection, favourite, not interested, the blocks
 * and "Open on the site"; onfavorite gets the new value.
 */
export function recordActions(
	record: MenuRecord,
	onfavorite: (favorite: boolean) => void
): SheetAction[] {
	const noun = nounOf(record.kind);
	const actions: SheetAction[] = [
		...collectionActions(record),
		{
			label: record.favorite ? 'Remove from favourites' : 'Add to favourites',
			icon: HeartIcon,
			run: () => void toggleFavorite(record, onfavorite)
		},
		{
			label: 'Not interested',
			icon: EyeSlashIcon,
			run: () => notInterested(record.kind, record.id, noun)
		},
		{
			label: `Block this ${noun.toLowerCase()}`,
			icon: ProhibitIcon,
			run: () => block(record.kind, record.id, noun),
			destructive: true
		}
	];
	for (const performer of record.performers.slice(0, MENU_PERFORMERS)) {
		actions.push({
			label: `Block ${performer.name}`,
			icon: UserMinusIcon,
			run: () => block('performer', performer.id, performer.name),
			destructive: true
		});
	}
	if (record.studio) {
		const studio = record.studio;
		actions.push({
			label: `Block ${studio.name}`,
			icon: BuildingsIcon,
			run: () => block('studio', studio.id, studio.name),
			destructive: true
		});
	}
	if (record.sourceUrl) {
		const sourceUrl = record.sourceUrl;
		actions.push({
			label: 'Open on the site',
			icon: ArrowSquareOutIcon,
			run: () => window.open(sourceUrl, '_blank', 'noopener')
		});
	}
	return actions;
}

/** "Play next" (videos) and "Add to collection…" (videos and galleries). */
function collectionActions(record: MenuRecord): SheetAction[] {
	if (record.kind !== 'video' && record.kind !== 'gallery') return [];
	const subject = { kind: record.kind, id: record.id };
	const actions: SheetAction[] = [];
	if (record.kind === 'video') {
		actions.push({
			label: 'Play next',
			icon: QueueIcon,
			run: () => void queueNext(record.id)
		});
	}
	actions.push({
		label: 'Add to collection…',
		icon: StackPlusIcon,
		run: () => pickCollectionFor(subject, record.title)
	});
	return actions;
}

/** Puts a video first in the queue and says so. */
async function queueNext(videoId: string): Promise<void> {
	await playNext(videoId);
	showUndo({ message: 'Plays next', link: { label: 'Queue', href: libraryUrl('queue') } });
}

/** Flips the record's favourite mark. */
async function toggleFavorite(
	record: MenuRecord,
	onfavorite: (favorite: boolean) => void
): Promise<void> {
	const favorite = !record.favorite;
	await setFavorite(record.kind, record.id, favorite);
	onfavorite(favorite);
}

/** Hides the record now; the thumbs down is sent once Undo has passed. */
function notInterested(kind: RecordKind, id: string, noun: string): void {
	setHidden(kind, id, true);
	showUndo({
		message: `${noun} hidden. You won’t be recommended it again.`,
		undo: () => setHidden(kind, id, false),
		settle: () => logNotInterested(kind, id)
	});
}

/** Blocks a record, with Undo. */
export function block(kind: RecordKind, id: string, name: string): void {
	void setBlocked(kind, id, true);
	showUndo({ message: `Blocked ${name}`, undo: () => void setBlocked(kind, id, false) });
}
