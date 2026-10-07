import { browser } from '$app/environment';
import { writable } from 'svelte/store';

const COLLAPSED_STORAGE_KEY = 'veil:sidebar-collapsed';

/** Whether the sidebar was collapsed to the icon rail last time. */
function loadCollapsed(): boolean {
	if (!browser) return false;
	return localStorage.getItem(COLLAPSED_STORAGE_KEY) === 'true';
}

// Full sidebar (false) or the narrow icon rail (true), toggled from the top bar.
export const sidebarCollapsed = writable(loadCollapsed());

sidebarCollapsed.subscribe((collapsed) => {
	if (!browser) return;
	localStorage.setItem(COLLAPSED_STORAGE_KEY, String(collapsed));
});
