import { readable } from 'svelte/store';
import { browser } from '$app/environment';
import { fetchPlugins, type SitePlugin } from '$lib/plugins';

// The loaded plugins, fetched once per page load, so any card can name the
// site a record came from by its source URL (see pluginForUrl).
export const sites = readable<SitePlugin[]>([], (set) => {
	if (!browser) return;
	fetchPlugins()
		.then(set)
		.catch(() => set([]));
});
