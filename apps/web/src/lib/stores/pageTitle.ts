import { writable } from 'svelte/store';

/** What a detail page shows in the phone top bar once its hero scrolls away. */
export interface PageTitle {
	title: string;
	imageUrl: string | null;
	// Performers get a round photo, studios their logo on a rounded tile.
	imageShape: 'avatar' | 'logo';
}

/** Set by detail pages (EntityHero) while they are shown; null elsewhere. */
export const pageTitle = writable<PageTitle | null>(null);

/**
 * Whether the detail page's hero is on screen: the top bar then stays
 * transparent over it and leaves the name to the hero.
 */
export const pageHeroVisible = writable(false);
