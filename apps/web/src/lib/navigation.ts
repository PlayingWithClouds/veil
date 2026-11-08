import type { Component } from 'svelte';
import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
import UsersIcon from 'phosphor-svelte/lib/UsersIcon';
import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import BookmarksSimpleIcon from 'phosphor-svelte/lib/BookmarksSimpleIcon';
import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
import {
	audiosUrl,
	camsUrl,
	categoriesUrl,
	galleriesUrl,
	libraryUrl,
	pluginsUrl,
	storiesUrl
} from '$lib/routes';

/** One top-level destination, in the sidebar and the phone tab bar. */
export interface NavLink {
	label: string;
	icon: Component;
	href: string;
	// Match the route exactly rather than by prefix.
	exact?: boolean;
	// The sidebar section it's listed under.
	group: NavGroup;
}

/** The sidebar's sections, in order. */
export const NAV_GROUPS = ['Watch', 'Library', 'You'] as const;
export type NavGroup = (typeof NAV_GROUPS)[number];

/** The top-level pages, in menu order (the phone tab bar's too). */
export const navLinks: NavLink[] = [
	{ label: 'Home', icon: HouseIcon, href: '/', exact: true, group: 'Watch' },
	{ label: 'Cams', icon: VideoCameraIcon, href: camsUrl(), group: 'Watch' },
	{ label: 'Galleries', icon: ImagesIcon, href: galleriesUrl(), group: 'Watch' },
	{ label: 'Categories', icon: SquaresFourIcon, href: categoriesUrl(), group: 'Library' },
	{ label: 'Audio', icon: HeadphonesIcon, href: audiosUrl(), group: 'Watch' },
	{ label: 'Stories', icon: BookOpenTextIcon, href: storiesUrl(), group: 'Watch' },
	{ label: 'Performers', icon: UsersIcon, href: '/performers', group: 'Library' },
	{ label: 'Studios', icon: FilmSlateIcon, href: '/studios', group: 'Library' },
	{ label: 'Library', icon: BookmarksSimpleIcon, href: libraryUrl(), group: 'You' },
	{ label: 'Taste', icon: HeartIcon, href: '/taste', group: 'You' },
	{ label: 'Plugins', icon: PlugsIcon, href: pluginsUrl(), group: 'You' }
];

/** Where the top bar's search goes and what it says it searches. */
export interface SearchScope {
	path: string;
	placeholder: string;
	// Whether the library filter's chip row shows under the search; cams have their own filters.
	libraryFilter: boolean;
}

/** The search scope of a page: cam, gallery, audio and story pages search their own sites. */
export function searchScopeOf(pathname: string): SearchScope {
	if (pathname.startsWith(camsUrl()))
		return { path: camsUrl(), placeholder: 'Search cams', libraryFilter: false };
	if (pathname.startsWith(galleriesUrl()))
		return { path: galleriesUrl(), placeholder: 'Search galleries', libraryFilter: true };
	if (pathname.startsWith(audiosUrl()))
		return { path: audiosUrl(), placeholder: 'Search audio', libraryFilter: true };
	if (pathname.startsWith(storiesUrl()))
		return { path: storiesUrl(), placeholder: 'Search stories', libraryFilter: true };
	return { path: '/', placeholder: 'Search the sites', libraryFilter: true };
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
