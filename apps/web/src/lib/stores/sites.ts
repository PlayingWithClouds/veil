import { get, readable } from 'svelte/store';
import { browser } from '$app/environment';
import type { ListKind } from '@playingwithclouds/veil-sdk/generated';
import { fetchPlugins, listers, type SitePlugin } from '$lib/plugins';

// The loaded plugins, fetched once per page load, so any card can name the
// site a record came from by its source URL (see pluginForUrl).
export const sites = readable<SitePlugin[]>([], (set) => {
	if (!browser) return;
	fetchPlugins()
		.then(set)
		.catch(() => set([]));
});

/**
 * The names of the enabled plugins that list a kind, as loaded now. Empty
 * until the plugins have loaded, which needs a component subscribed to
 * sites (every page listing media uses $sites).
 */
export function listerNames(kind: ListKind): string[] {
	return listers(get(sites), kind).map((site) => site.name);
}
