import { gqlClient } from '$lib/veil';
import { apiBase } from '$lib/server';
import type { Image } from '$lib/images';
import type { TagRef } from '$lib/tags';

export interface GalleryCategory {
	id: string;
	name: string;
	poster: string | null;
}

// Fetches PornPics' category index (each with a preview thumbnail).
export async function fetchGalleryCategories(): Promise<GalleryCategory[]> {
	const result = await gqlClient.query({
		pluginCategories: {
			__args: { plugin: 'pornpics' },
			id: true,
			name: true,
			poster: true
		}
	});
	return (result.pluginCategories ?? []) as GalleryCategory[];
}

export interface GallerySummary {
	id: string;
	title: string;
	coverPath: string | null;
	imageCount: number;
}

export interface GalleryDetail {
	id: string;
	title: string;
	details: string | null;
	date: string | null;
	coverPath: string | null;
	imageCount: number;
	tags: TagRef[];
	images: Image[];
}

export async function fetchGalleries(limit: number, offset: number): Promise<GallerySummary[]> {
	const result = await gqlClient.query({
		galleries: {
			__args: { limit, offset },
			id: true,
			title: true,
			coverPath: true,
			imageCount: true
		}
	});
	return (result.galleries ?? []) as GallerySummary[];
}

const GALLERY_DETAIL_SELECT = {
	id: true,
	title: true,
	details: true,
	date: true,
	coverPath: true,
	imageCount: true,
	tags: { id: true, name: true },
	images: {
		id: true,
		filePath: true,
		title: true,
		width: true,
		height: true,
		position: true,
		galleryId: true,
		tags: { id: true, name: true }
	}
} as const;

export async function loadGalleryDetail(id: string): Promise<GalleryDetail | null> {
	const result = await gqlClient.query({
		gallery: { __args: { id }, ...GALLERY_DETAIL_SELECT }
	});
	return (result.gallery ?? null) as GalleryDetail | null;
}

/**
 * Records a gallery visit: a stub gallery (from search/browse) has its page
 * fetched from its site so its images land. Resolves to the updated gallery.
 */
export async function ensureGalleryImages(id: string): Promise<GalleryDetail | null> {
	const result = await gqlClient.mutation({
		ensureGalleryImages: { __args: { galleryId: id }, ...GALLERY_DETAIL_SELECT }
	});
	return (result.ensureGalleryImages ?? null) as GalleryDetail | null;
}

// Seeds new galleries from the PornPics plugin. Uses the search SSE endpoint,
// which ingests each discovered gallery as a stub and enqueues a background
// scrape that fills in its full-size images. Resolves once the stream ends.
export async function seedGalleries(query: string): Promise<void> {
	const url = `${apiBase()}/api/search?q=${encodeURIComponent(query)}&sources=pornpics&limit=24`;
	await new Promise<void>((resolve) => {
		const source = new EventSource(url);
		const finish = () => {
			source.close();
			resolve();
		};
		source.addEventListener('done', finish);
		source.addEventListener('error', finish);
	});
}
