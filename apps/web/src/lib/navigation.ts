import type { Component } from 'svelte';
import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
import UsersIcon from 'phosphor-svelte/lib/UsersIcon';
import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
import { pluginsUrl } from '$lib/routes';

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
	{ label: 'Performers', icon: UsersIcon, href: '/performers' },
	{ label: 'Studios', icon: FilmSlateIcon, href: '/studios' },
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
