// In-progress playback positions, keyed by scene id. Loaded once (and refreshed
// on navigation) so any SceneCard can draw a resume bar
// by looking itself up — no need to thread progress through every list query.
import { writable } from 'svelte/store';
import { gqlClient } from '$lib/veil';

export interface SceneProgress {
	progressSeconds: number;
	durationSeconds: number | null;
}

export const watchProgress = writable<Map<string, SceneProgress>>(new Map());

// Scenes watched to the end, for the "Watched" mark on cards.
export const watchedScenes = writable<Set<string>>(new Set());

interface WatchHistoryRow {
	media: string;
	progressSeconds: number;
	durationSeconds: number | null;
	completed: boolean;
}

// Loads the unfinished watches into watchProgress and the finished ones into
// watchedScenes. Zero-progress rows are skipped: they would draw an empty bar.
export async function loadWatchProgress(): Promise<void> {
	try {
		const data = await gqlClient.query({
			watchHistory: {
				__args: { limit: 500 },
				media: true,
				progressSeconds: true,
				durationSeconds: true,
				completed: true
			}
		});
		const rows = (data.watchHistory ?? []) as unknown as WatchHistoryRow[];
		const next = new Map<string, SceneProgress>();
		const watched = new Set<string>();
		for (const row of rows) {
			if (row.completed) {
				watched.add(row.media);
				continue;
			}
			if (row.progressSeconds <= 0) continue;
			next.set(row.media, {
				progressSeconds: row.progressSeconds,
				durationSeconds: row.durationSeconds
			});
		}
		watchProgress.set(next);
		watchedScenes.set(watched);
	} catch {
		// Best-effort: leave the previous map in place on failure.
	}
}
