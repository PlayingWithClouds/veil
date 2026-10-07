// Content images: gallery pages and standalone images. Browsable, taggable,
// collectible — asset images (posters, logos) never appear here.
import { gqlClient } from '$lib/veil';
import type { TagRef } from '$lib/tags';

export interface Image {
	id: string;
	filePath: string;
	title: string | null;
	width: number | null;
	height: number | null;
	position: number | null;
	galleryId: string | null;
	tags: TagRef[];
}

const IMAGE_FIELDS = {
	id: true,
	filePath: true,
	title: true,
	width: true,
	height: true,
	position: true,
	galleryId: true,
	tags: { id: true, name: true }
} as const;

export interface ImageFilters {
	tagId?: string;
	performerId?: string;
	studioId?: string;
	galleryId?: string;
}

export async function fetchImages(
	filters: ImageFilters,
	limit = 60,
	offset = 0
): Promise<Image[]> {
	try {
		const data = await gqlClient.query({
			images: { __args: { ...filters, limit, offset }, ...IMAGE_FIELDS }
		});
		return (data.images ?? []) as unknown as Image[];
	} catch {
		return [];
	}
}

export async function setImageTags(imageId: string, tagIds: string[]): Promise<void> {
	await gqlClient.mutation({ setImageTags: { __args: { imageId, tagIds }, id: true } });
}
