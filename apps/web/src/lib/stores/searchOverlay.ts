import { writable } from 'svelte/store';

/** Whether the phone's full-screen search overlay is open. */
export const searchOverlayOpen = writable(false);
