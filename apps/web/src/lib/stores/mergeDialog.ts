// Drives the global MergeDialog (mounted once in the layout): any page can offer
// "merge into…" for a performer, studio or tag.
import { writable } from 'svelte/store';
import type { MergeKind } from '$lib/merge';

export interface MergeDialogTarget {
	kind: MergeKind;
	id: string;
	name: string;
}

export const mergeDialogTarget = writable<MergeDialogTarget | null>(null);

/** Opens the dialog to merge the record into another one. */
export function openMergeDialog(kind: MergeKind, id: string, name: string): void {
	mergeDialogTarget.set({ kind, id, name });
}

export function closeMergeDialog(): void {
	mergeDialogTarget.set(null);
}
