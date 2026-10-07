import type { Component } from 'svelte';
import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
import BellIcon from 'phosphor-svelte/lib/BellIcon';
import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
import FolderSimpleIcon from 'phosphor-svelte/lib/FolderSimpleIcon';
import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
import { subscriptionsUrl } from '$lib/routes';

/** A destination of the phone tab bar. */
export type Tab = {
	label: string;
	icon: Component;
	href: string;
	// Match the route exactly rather than by prefix.
	exact?: boolean;
};

/** The phone tab bar, in order; swiping sideways steps through them. */
export const tabs: Tab[] = [
	{ label: 'Home', icon: HouseIcon, href: '/', exact: true },
	{ label: 'Subscriptions', icon: BellIcon, href: subscriptionsUrl() },
	{ label: 'Library', icon: FilmStripIcon, href: '/library' },
	{ label: 'Collections', icon: FolderSimpleIcon, href: '/collections' },
	{ label: 'History', icon: ClockCounterClockwiseIcon, href: '/history' }
];

/** Whether the tab points at the page at this path. */
export function isTabActive(tab: Tab, pathname: string): boolean {
	if (tab.exact) return pathname === tab.href;
	return pathname.startsWith(tab.href);
}

/** Index of the tab showing this path, or -1 on pages without a tab. */
export function activeTabIndex(pathname: string): number {
	return tabs.findIndex((tab) => isTabActive(tab, pathname));
}
