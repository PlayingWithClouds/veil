// Impression events of the watch page's up-next list (surface "related"): the
// scene's related videos, then the personal feed and the site results.
import type { ShownEvent } from '$lib/recommendations';
import type { PluginSearchResult } from '$lib/search';

/**
 * The impression of an up-next item at a position of the list. Site results
 * that aren't stored yet have no scene id and are skipped by the logger.
 */
export function relatedImpression(item: PluginSearchResult, position: number): ShownEvent {
	return { mediaId: item.sceneId ?? item.dbId ?? '', surface: 'related', position };
}
