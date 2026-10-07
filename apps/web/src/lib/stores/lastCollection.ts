// Remembers the collection most recently saved to, so a card bookmark can
// one-tap add to it. Persisted in localStorage.
import { browser } from '$app/environment';

const STORAGE_KEY = 'veil:last-collection';

export function getLastCollection(): string | null {
	if (!browser) return null;
	return localStorage.getItem(STORAGE_KEY);
}

export function setLastCollection(collectionId: string): void {
	if (!browser) return;
	localStorage.setItem(STORAGE_KEY, collectionId);
}
