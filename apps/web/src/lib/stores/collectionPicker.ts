// The record the "Add to collection" dialog is open for, if any. Card menus
// open it from plain TS; the layout shows the one dialog.
import { writable } from 'svelte/store';
import type { CollectableRef } from '$lib/curation';

/** The record being added, with its title for the dialog's heading. */
export interface CollectionPick {
	subject: CollectableRef;
	title: string;
}

export const collectionPick = writable<CollectionPick | null>(null);

/** Opens the "Add to collection" dialog for a record. */
export function pickCollectionFor(subject: CollectableRef, title: string): void {
	collectionPick.set({ subject, title });
}
