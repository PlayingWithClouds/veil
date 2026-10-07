import type { PluginSearchResult } from '$lib/search';
import type { SceneCardData } from '$lib/components/SceneCard.svelte';
import { cacheUrl } from '$lib/img';

// Curated adult browse categories. `slug` keys the /nsfw/[category] route;
// `query` is the term handed to plugin search (adult sources index by keyword);
// `imageUrl` is a representative still (Pornhub's static category art, 706x400,
// unsigned and stable — loaded through the image cache proxy).
export interface NsfwCategory {
	slug: string;
	label: string;
	query: string;
	imageUrl: string;
}

function pornhubCategoryImage(categoryId: number): string {
	return `https://ei.phncdn.com/static/images/categories/roku_${categoryId}.jpg`;
}

export const NSFW_CATEGORIES: NsfwCategory[] = [
	{ slug: 'amateur', label: 'Amateur', query: 'amateur', imageUrl: pornhubCategoryImage(3) },
	{ slug: 'anal', label: 'Anal', query: 'anal', imageUrl: pornhubCategoryImage(35) },
	{ slug: 'asian', label: 'Asian', query: 'asian', imageUrl: pornhubCategoryImage(1) },
	// Pornhub has no "Babe" category; Solo Female is the closest visual stand-in.
	{ slug: 'babe', label: 'Babe', query: 'babe', imageUrl: pornhubCategoryImage(492) },
	{ slug: 'big-ass', label: 'Big Ass', query: 'big ass', imageUrl: pornhubCategoryImage(4) },
	{ slug: 'big-tits', label: 'Big Tits', query: 'big tits', imageUrl: pornhubCategoryImage(8) },
	{ slug: 'blonde', label: 'Blonde', query: 'blonde', imageUrl: pornhubCategoryImage(9) },
	{ slug: 'blowjob', label: 'Blowjob', query: 'blowjob', imageUrl: pornhubCategoryImage(13) },
	{ slug: 'brunette', label: 'Brunette', query: 'brunette', imageUrl: pornhubCategoryImage(11) },
	{ slug: 'creampie', label: 'Creampie', query: 'creampie', imageUrl: pornhubCategoryImage(15) },
	{ slug: 'ebony', label: 'Ebony', query: 'ebony', imageUrl: pornhubCategoryImage(17) },
	{ slug: 'latina', label: 'Latina', query: 'latina', imageUrl: pornhubCategoryImage(26) },
	{ slug: 'lesbian', label: 'Lesbian', query: 'lesbian', imageUrl: pornhubCategoryImage(27) },
	{ slug: 'mature', label: 'Mature', query: 'mature', imageUrl: pornhubCategoryImage(28) },
	{ slug: 'milf', label: 'MILF', query: 'milf', imageUrl: pornhubCategoryImage(29) },
	{ slug: 'threesome', label: 'Threesome', query: 'threesome', imageUrl: pornhubCategoryImage(65) }
];

export function categoryBySlug(slug: string): NsfwCategory | undefined {
	return NSFW_CATEGORIES.find((category) => category.slug === slug);
}

// Adapts a plugin search/browse result to the shared scene-card shape, labelled
// with the site it came from. Every image is routed through the image cache
// proxy so hotlink-protected adult CDNs load in the browser.
export function toSceneCard(result: PluginSearchResult): SceneCardData {
	const previewImages = result.previewImages
		.map((url) => cacheUrl(url))
		.filter((url): url is string => url !== null);
	return {
		id: result.sceneId,
		title: result.title,
		bannerUrl: cacheUrl(result.posterUrl),
		previewImages,
		previewVideo: cacheUrl(result.previewVideo),
		durationSeconds: result.durationSeconds,
		channelName: result.studioName,
		channelImageUrl: cacheUrl(result.studioImagePath),
		sourceUrl: result.sourceUrl,
		date: result.date,
		isNew: result.isNew
	};
}
