// Styles of video the search filter can pick all sites of at once: cams,
// studio productions, amateur uploads, JAV, fetish. A site is named by its
// filter value (siteChoices: a plugin's name, or "plugin/host" for one site
// of a plugin serving several). "Tubes" is every site no other style claims.
import type { SiteChoice } from '$lib/plugins';

/** A style and the sites carrying it, as the filter offers it. */
export interface SiteStyle {
	key: string;
	label: string;
	// Filter values of the installed sites of this style.
	sites: string[];
}

/** The styles with their sites; sites not installed are skipped. */
const STYLES: { key: string; label: string; sites: string[] }[] = [
	{
		key: 'cams',
		label: 'Cams',
		sites: [
			'showcamrips',
			'kvs/cloudbate.com',
			'kvs/camstreams.tv',
			'kvs/stream-leak.com',
			'kvs/camgirlsleak.com'
		]
	},
	{
		key: 'professional',
		label: 'Professional',
		sites: [
			'aylo',
			'hqporner',
			'porn-star',
			'kvs/pornhat.com',
			'kvs/perfectgirls.xxx',
			'kvs/xxxfiles.com'
		]
	},
	{
		key: 'amateur',
		label: 'Amateur',
		sites: [
			'kvs/xgroovy.com',
			'kvs/fapnfuck.com',
			'kvs/fuqster.com',
			'kvs/fapgoat.com',
			'kvs/w1mp.com',
			'kvs/xozilla.com'
		]
	},
	{ key: 'jav', label: 'JAV', sites: ['missav', 'kvs/javbangers.com'] },
	{
		key: 'fetish',
		label: 'Fetish',
		sites: ['kvs/severeporn.com', 'kvs/analdin.com', 'kvs/whoreshub.com']
	}
];

/**
 * The styles among the offered sites, each with the sites it picks, then
 * "Tubes" for the rest. A style with none of its sites installed is left out.
 */
export function siteStyles(choices: SiteChoice[]): SiteStyle[] {
	const offered = new Set(choices.map((choice) => choice.value));
	const claimed = new Set(STYLES.flatMap((style) => style.sites));
	const styles = STYLES.map((style) => ({
		...style,
		sites: style.sites.filter((site) => offered.has(site))
	}));
	styles.push({
		key: 'tubes',
		label: 'Tubes',
		sites: choices.map((choice) => choice.value).filter((value) => !claimed.has(value))
	});
	return styles.filter((style) => style.sites.length > 0);
}
