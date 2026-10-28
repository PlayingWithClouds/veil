// Why an up-next item is in the watch page's rail, and the filter groups the
// rail offers above the list.
import type { PluginSearchResult } from '$lib/search';

// site: the site's own related list; cast: found by a performer of this scene;
// similar: found by a tag or the title; feed: the personal feed or site browse.
export type UpNextGroup = 'site' | 'cast' | 'similar' | 'feed';

/** Which filter group an up-next item belongs to. */
export function upNextGroup(item: PluginSearchResult): UpNextGroup {
	if (item.relatedVia === 'site') return 'site';
	if (item.relatedVia === 'performer') return 'cast';
	if (item.relatedVia === 'tag' || item.relatedVia === 'title') return 'similar';
	return 'feed';
}

/** The first of the scene's performers the item credits too (names compared case-insensitively). */
export function sharedPerformer(item: PluginSearchResult, cast: string[]): string | null {
	if (!item.performerNames) return null;
	const credited = new Set(item.performerNames.map((name) => name.toLowerCase()));
	const shared = cast.find((name) => credited.has(name.toLowerCase()));
	if (shared === undefined) return null;
	return shared;
}

/** The short "why" line on a rail card, e.g. "With Nuria Millan" or "Similar tags". */
export function upNextReason(item: PluginSearchResult, cast: string[]): string {
	const performer = sharedPerformer(item, cast);
	if (performer) return `With ${performer}`;
	if (item.relatedVia === 'site') return "Site's pick";
	if (item.relatedVia === 'performer') return 'Same performer';
	if (item.relatedVia === 'tag') return 'Similar tags';
	if (item.relatedVia === 'title') return 'Similar title';
	return 'For you';
}

/** Chip label of a filter group; the cast chip names the performer when there is one. */
export function upNextGroupLabel(group: UpNextGroup, cast: string[]): string {
	if (group === 'site') return "Site's picks";
	if (group === 'cast') {
		if (cast.length === 1) return `With ${cast[0]}`;
		return 'Same performers';
	}
	if (group === 'similar') return 'Similar';
	return 'For you';
}
