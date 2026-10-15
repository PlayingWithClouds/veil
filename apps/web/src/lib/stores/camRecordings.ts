import { writable } from 'svelte/store';
import { fetchActiveCamRecordings, type CamRecording } from '$lib/cams';

// How often the backend is asked which rooms are being recorded.
const POLL_INTERVAL_MS = 10_000;

let pollTimer: ReturnType<typeof setInterval> | null = null;

/** Fetches the recordings in progress and publishes them. */
export async function refreshCamRecordings() {
	activeStore.set(await fetchActiveCamRecordings());
}

// Polled only while something subscribes (the sidebar, the Cams page, a room
// page), so their indicators stay current without a permanent poll.
const activeStore = writable<CamRecording[]>([], () => {
	void refreshCamRecordings();
	pollTimer = setInterval(refreshCamRecordings, POLL_INTERVAL_MS);
	return () => {
		if (pollTimer) clearInterval(pollTimer);
		pollTimer = null;
	};
});

/** The live-cam recordings in progress, with their duration and size so far. */
export const activeCamRecordings = { subscribe: activeStore.subscribe };
