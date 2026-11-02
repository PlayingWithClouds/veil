// One toast at a time saying what just happened, with an Undo button. A new
// toast settles the one before it.
import { writable } from 'svelte/store';

const TOAST_MS = 6000;

/** A toast: its message and what Undo does; no Undo button without one. */
export interface UndoToast {
	message: string;
	undo?: () => void;
	// A page to go on to, e.g. where the suggestions it mentions are.
	link?: { label: string; href: string };
	// Runs when the toast goes away without Undo, e.g. to send what was held back.
	settle?: () => void;
}

export const undoToast = writable<UndoToast | null>(null);

let current: UndoToast | null = null;
let timer: ReturnType<typeof setTimeout> | null = null;

/** Shows a toast, settling the one it replaces. */
export function showUndo(toast: UndoToast): void {
	settleCurrent();
	current = toast;
	undoToast.set(toast);
	timer = setTimeout(settleCurrent, TOAST_MS);
}

/** Runs the toast's undo and closes it. */
export function undoCurrent(): void {
	if (!current) return;
	const toast = current;
	close();
	toast.undo?.();
}

/** Closes the toast without undoing: what it did stays. */
export function settleCurrent(): void {
	if (!current) return;
	const toast = current;
	close();
	toast.settle?.();
}

/** Forgets the shown toast and its timer. */
function close(): void {
	if (timer !== null) {
		clearTimeout(timer);
		timer = null;
	}
	current = null;
	undoToast.set(null);
}
