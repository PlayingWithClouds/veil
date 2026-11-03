import type { Component } from 'svelte';
import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
import UsersIcon from 'phosphor-svelte/lib/UsersIcon';
import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
import { audiosUrl, pluginsUrl, storiesUrl } from '$lib/routes';

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
	{ label: 'Galleries', icon: ImagesIcon, href: '/galleries' },
	{ label: 'Audio', icon: HeadphonesIcon, href: audiosUrl() },
	{ label: 'Stories', icon: BookOpenTextIcon, href: storiesUrl() },
	{ label: 'Performers', icon: UsersIcon, href: '/performers' },
	{ label: 'Studios', icon: FilmSlateIcon, href: '/studios' },
	{ label: 'Taste', icon: HeartIcon, href: '/taste' },
	{ label: 'Plugins', icon: PlugsIcon, href: pluginsUrl() }
];

/** Whether the link points at the page at pathname. */
export function isNavActive(link: NavLink, pathname: string): boolean {
	if (link.exact) return pathname === link.href;
	return pathname.startsWith(link.href);
}

/** The index of the top-level page at pathname, or -1 on detail pages. */
export function activeNavIndex(pathname: string): number {
	return navLinks.findIndex((link) => link.href === pathname);
}
