import type { Component } from 'svelte';
import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
import UsersIcon from 'phosphor-svelte/lib/UsersIcon';
import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import { audiosUrl, galleriesUrl, pluginsUrl, storiesUrl } from '$lib/routes';

/** One top-level destination, in the sidebar and the phone tab bar. */
export interface NavLink {
	label: string;
	icon: Component;
	href: string;
	// Match the route exactly rather than by prefix.
	exact?: boolean;
}

/** The top-level pages, in menu order. */
export const navLinks: NavLink[] = [
	{ label: 'Home', icon: HouseIcon, href: '/', exact: true },
	{ label: 'Galleries', icon: ImagesIcon, href: galleriesUrl() },
	{ label: 'Audio', icon: HeadphonesIcon, href: audiosUrl() },
	{ label: 'Stories', icon: BookOpenTextIcon, href: storiesUrl() },
	{ label: 'Performers', icon: UsersIcon, href: '/performers' },
	{ label: 'Studios', icon: FilmSlateIcon, href: '/studios' },
	{ label: 'Taste', icon: HeartIcon, href: '/taste' },
	{ label: 'Plugins', icon: PlugsIcon, href: pluginsUrl() }
];

/** Where the top bar's search goes and what it says it searches. */
export interface SearchScope {
	path: string;
	placeholder: string;
}

/** The search scope of a page: gallery, audio and story pages search their own sites. */
export function searchScopeOf(pathname: string): SearchScope {
	if (pathname.startsWith(galleriesUrl()))
		return { path: galleriesUrl(), placeholder: 'Search galleries' };
	if (pathname.startsWith(audiosUrl())) return { path: audiosUrl(), placeholder: 'Search audio' };
	if (pathname.startsWith(storiesUrl()))
		return { path: storiesUrl(), placeholder: 'Search stories' };
	return { path: '/', placeholder: 'Search the sites' };
}

/** Whether the link points at the page at pathname. */
export function isNavActive(link: NavLink, pathname: string): boolean {
	if (link.exact) return pathname === link.href;
	return pathname.startsWith(link.href);
}

/** The index of the top-level page at pathname, or -1 on detail pages. */
export function activeNavIndex(pathname: string): number {
	return navLinks.findIndex((link) => link.href === pathname);
}
