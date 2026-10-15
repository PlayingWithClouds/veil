// The split view's panes: which cams and videos are open side by side, in tab
// order, and which one is focused (the only one with sound). Kept in
// localStorage so the view survives navigation and reload; playback itself
// restarts on reload.
import { browser } from '$app/environment';
import { writable } from 'svelte/store';

/** Most panes the split view holds. */
export const MAX_SPLIT_PANES = 4;

const STORAGE_KEY = 'veil:split';

export type SplitPaneKind = 'scene' | 'cam';

export interface SplitPane {
	kind: SplitPaneKind;
	// The scene or cam model id.
	mediaId: string;
	title: string;
}

export interface SplitState {
	panes: SplitPane[];
	focusedIndex: number;
}

/** What adding a pane did, so the caller can tell the user. */
export type SplitAddResult =
	| { outcome: 'added' }
	| { outcome: 'existing' }
	| { outcome: 'replaced'; replacedTitle: string };

/** Identity of a pane: the same stream is never open twice. */
export function splitPaneKey(pane: Pick<SplitPane, 'kind' | 'mediaId'>): string {
	return `${pane.kind}:${pane.mediaId}`;
}

/** Reads the saved state, dropping anything malformed. */
function loadState(): SplitState {
	const empty: SplitState = { panes: [], focusedIndex: 0 };
	if (!browser) return empty;
	try {
		const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? 'null');
		if (!saved || !Array.isArray(saved.panes)) return empty;
		const panes: SplitPane[] = saved.panes
			.filter(
				(pane: SplitPane) =>
					(pane.kind === 'scene' || pane.kind === 'cam') &&
					typeof pane.mediaId === 'string' &&
					typeof pane.title === 'string'
			)
			.slice(0, MAX_SPLIT_PANES);
		const focusedIndex = Number(saved.focusedIndex);
		if (!Number.isInteger(focusedIndex) || focusedIndex < 0 || focusedIndex >= panes.length) {
			return { panes, focusedIndex: 0 };
		}
		return { panes, focusedIndex };
	} catch {
		return empty;
	}
}

const state = writable<SplitState>(loadState());

if (browser) {
	state.subscribe((current) => localStorage.setItem(STORAGE_KEY, JSON.stringify(current)));
}

/** The split view's panes and focus. */
export const splitState = { subscribe: state.subscribe };

/** Index of the pane with the given key, or -1. */
function indexOfKey(panes: SplitPane[], key: string): number {
	return panes.findIndex((pane) => splitPaneKey(pane) === key);
}

/**
 * Adds a pane and focuses it. A stream that is already open is only focused;
 * when all panes are taken the focused one is replaced.
 */
export function addSplitPane(pane: SplitPane): SplitAddResult {
	let result: SplitAddResult = { outcome: 'added' };
	state.update((current) => {
		const existingIndex = indexOfKey(current.panes, splitPaneKey(pane));
		if (existingIndex >= 0) {
			result = { outcome: 'existing' };
			return { ...current, focusedIndex: existingIndex };
		}
		if (current.panes.length < MAX_SPLIT_PANES) {
			return { panes: [...current.panes, pane], focusedIndex: current.panes.length };
		}
		result = { outcome: 'replaced', replacedTitle: current.panes[current.focusedIndex].title };
		const panes = [...current.panes];
		panes[current.focusedIndex] = pane;
		return { panes, focusedIndex: current.focusedIndex };
	});
	return result;
}

/** Closes a pane; focus stays on the same pane, or moves to its neighbour. */
export function closeSplitPane(key: string) {
	state.update((current) => {
		const index = indexOfKey(current.panes, key);
		if (index < 0) return current;
		const panes = current.panes.filter((_, position) => position !== index);
		let focusedIndex = current.focusedIndex;
		if (index < focusedIndex) focusedIndex -= 1;
		focusedIndex = Math.max(0, Math.min(focusedIndex, panes.length - 1));
		return { panes, focusedIndex };
	});
}

/** Focuses a pane, which makes it the audible one. */
export function focusSplitPane(key: string) {
	state.update((current) => {
		const index = indexOfKey(current.panes, key);
		if (index < 0 || index === current.focusedIndex) return current;
		return { ...current, focusedIndex: index };
	});
}

/** Moves a pane to another position (tab drag); the focus follows its pane. */
export function moveSplitPane(fromIndex: number, toIndex: number) {
	state.update((current) => {
		if (fromIndex === toIndex) return current;
		if (fromIndex < 0 || fromIndex >= current.panes.length) return current;
		if (toIndex < 0 || toIndex >= current.panes.length) return current;
		const focusedKey = splitPaneKey(current.panes[current.focusedIndex]);
		const panes = [...current.panes];
		const [moved] = panes.splice(fromIndex, 1);
		panes.splice(toIndex, 0, moved);
		return { panes, focusedIndex: indexOfKey(panes, focusedKey) };
	});
}

/** Closes every pane. */
export function clearSplitPanes() {
	state.set({ panes: [], focusedIndex: 0 });
}
