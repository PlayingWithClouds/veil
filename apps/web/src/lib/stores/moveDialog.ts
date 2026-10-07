// Drives the global CollectionDialog (mounted once in the layout). Any surface
// can open the collection picker for an entity without owning the dialog markup.
import { writable } from 'svelte/store';
import type { CollectionEntityType } from '$lib/collections';

export interface MoveDialogTarget {
	entityType: CollectionEntityType;
	entityId: string;
	title: string;
}

export const moveDialogTarget = writable<MoveDialogTarget | null>(null);

export function openMoveDialog(
	entityType: CollectionEntityType,
	entityId: string,
	title: string
): void {
	moveDialogTarget.set({ entityType, entityId, title });
}

export function closeMoveDialog(): void {
	moveDialogTarget.set(null);
}
