// The scene that keeps playing in the bottom-right corner after the user leaves
// its page for a non-scene page. The scene page hands playback over on leave
// and takes it back (position and source) when reopened. Persisted to
// localStorage so it survives a reload.
import { browser } from '$app/environment';
import { get, writable } from 'svelte/store';
import { mainPlayback, type PlaybackHandoff } from '$lib/playback';

// The handoff back to the scene page; position is kept current while the mini
// player runs.
export interface MiniPlayback extends PlaybackHandoff {
	playing: boolean;
	muted: boolean;
}

const MINI_PLAYER_STORAGE_KEY = 'veil:mini-player';

/** The mini player left open before the last reload, if any. */
function loadMiniPlayback(): MiniPlayback | null {
	if (!browser) return null;
	const stored = localStorage.getItem(MINI_PLAYER_STORAGE_KEY);
	if (!stored) return null;
	try {
		return JSON.parse(stored) as MiniPlayback;
	} catch {
		return null;
	}
}

export const miniPlayer = writable<MiniPlayback | null>(loadMiniPlayback());

miniPlayer.subscribe((playback) => {
	if (!browser) return;
	if (!playback) {
		localStorage.removeItem(MINI_PLAYER_STORAGE_KEY);
		return;
	}
	localStorage.setItem(MINI_PLAYER_STORAGE_KEY, JSON.stringify(playback));
});

/** Records the mini player's current position, play and mute state. */
export function updateMiniPlayback(position: number, playing: boolean, muted: boolean): void {
	miniPlayer.update((playback) => {
		if (!playback) return playback;
		return { ...playback, position, playing, muted };
	});
}

/**
 * Closes the mini player as a scene page opens. Returns its playback when it
 * was playing that same scene, so the page can take the running video over;
 * a different scene's video is stopped.
 */
export function takeMiniPlayback(sceneId: string): MiniPlayback | null {
	const playback = get(miniPlayer);
	if (!playback) return null;
	miniPlayer.set(null);
	if (playback.sceneId === sceneId) return playback;
	mainPlayback().stop();
	return null;
}
