// Uploaders: the accounts that put videos on tube sites. They are not studios
// and have no record of their own; a video's source remembers its uploader,
// and their videos are read from their channel page on that site.
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';
import type { Feed } from '$lib/feed';

/** An uploader on one site, as a video page links to it. */
export interface UploaderOnSite {
	plugin: string;
	name: string;
	// The channel page their videos are listed from; null when the site gave none.
	url: string | null;
	image: string | null;
}

/** The source rows a video page reads uploaders from. */
interface SourceWithUploader {
	plugin: string;
	uploader?: { name: string; url?: string | null; image?: { uri: string } | null } | null;
}

/** The distinct uploaders (by site and page) the sources name, in source order. */
export function uploadersOf(sources: SourceWithUploader[]): UploaderOnSite[] {
	const found = new Map<string, UploaderOnSite>();
	for (const source of sources) {
		const uploader = source.uploader;
		if (!uploader) continue;
		const url = uploader.url ?? null;
		const key = `${source.plugin}|${url ?? uploader.name}`;
		if (found.has(key)) continue;
		found.set(key, {
			plugin: source.plugin,
			name: uploader.name,
			url,
			image: uploader.image?.uri ?? null
		});
	}
	return [...found.values()];
}

/** Stores a page of the videos on an uploader's channel page and returns them. */
export async function fetchUploaderVideos(
	plugin: string,
	url: string,
	name: string,
	limit: number,
	offset: number
): Promise<VideoCardRow[]> {
	const data = await gqlClient.mutation({
		fetchUploader: { __args: { plugin, url, name, limit, offset }, ...VIDEO_CARD }
	});
	return data.fetchUploader;
}

/**
 * An endless feed of an uploader's videos. A site that ignores the offset
 * would repeat its first page for ever, so a page with nothing new ends it.
 */
export function uploaderFeed(plugin: string, url: string, name: string): Feed<VideoCardRow> {
	const seen = new Set<string>();
	let offset = 0;
	return async (limit) => {
		const page = await fetchUploaderVideos(plugin, url, name, limit, offset);
		offset += limit;
		const fresh = page.filter((video) => !seen.has(video.id));
		for (const video of fresh) seen.add(video.id);
		return fresh;
	};
}
