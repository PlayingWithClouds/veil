import { readable } from 'svelte/store';
import { browser } from '$app/environment';
import { fetchSearchPlugins, type SearchPlugin } from '$lib/search';

// The searchable sites, fetched once per page load, so any card can name the
// site a record came from by its source URL (see pluginForUrl).
export const sites = readable<SearchPlugin[]>([], (set) => {
	if (browser) fetchSearchPlugins().then(set);
});

/** A plugin's display name ("EPorner" for "eporner"), else the name as given. */
export function siteDisplayName(pluginName: string, list: SearchPlugin[]): string {
	const plugin = list.find((entry) => entry.name === pluginName);
	if (plugin?.displayName) return plugin.displayName;
	return pluginName;
}
