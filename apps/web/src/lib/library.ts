import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';
import type { ListKind } from '$lib/plugins';

/** Narrows a list of videos or galleries. Empty fields don't filter. */
export interface MediaFilter {
	text?: string;
	performerId?: string;
	studioId?: string;
	tagId?: string;
}

/** A named reference: a performer, studio or tag. */
export interface NamedRef {
	id: string;
	name: string;
}

/** A video as a card shows it. */
export interface VideoCardRow {
	id: string;
	title: string;
	date: string | null;
	durationSeconds: number | null;
	cover: { uri: string } | null;
	preview: { uri: string } | null;
	// Stills the card scrubs through on hover, in order.
	previewFrames: { uri: string }[] | null;
	trickplay: Trickplay | null;
	width: number | null;
	height: number | null;
	studio: (NamedRef & { image: { uri: string } | null }) | null;
	performers: NamedRef[];
	sources: { url: string }[];
	// progress is the resume point in seconds; rating is 0..10.
	userState: {
		progress: number | null;
		rating: number | null;
		favorite: boolean;
		viewCount: number;
	};
}

/** A video's scrubbing thumbnails: tiles of frames packed into sprite sheets. */
export interface Trickplay {
	sheets: { uri: string }[];
	columns: number;
	rows: number;
	tileWidth: number;
	tileHeight: number;
	intervalSeconds: number;
}

/** A gallery as a card shows it. */
export interface GalleryCardRow {
	id: string;
	title: string;
	date: string | null;
	imageCount: number;
	cover: { uri: string } | null;
	studio: NamedRef | null;
	sources: { url: string }[];
}

/** A performer or studio as a card shows it. */
export interface MemberRow {
	id: string;
	name: string;
	image: { uri: string } | null;
	videoCount: number;
	galleryCount: number;
}

/** One site a record was found on. */
export interface SourceRow {
	plugin: string;
	url: string;
	fetchedAt: string | null;
}

/** A playable stream a source offers. */
export interface StreamRow {
	id: string;
	url: string;
	mimeType: string;
	quality: string | null;
	source: { plugin: string } | null;
}

/** Everything the video page shows. */
export interface VideoDetail extends VideoCardRow {
	details: string | null;
	performers: (NamedRef & { image: { uri: string } | null })[];
	tags: NamedRef[];
	sources: SourceRow[];
	streams: StreamRow[];
}

/** Everything the gallery page shows. */
export interface GalleryDetail extends GalleryCardRow {
	details: string | null;
	performers: NamedRef[];
	tags: NamedRef[];
	images: { id: string; position: number; asset: { uri: string } }[];
	sources: SourceRow[];
}

/** A performer's page. */
export interface PerformerDetail extends MemberRow {
	aliases: string[];
	details: string | null;
	links: string[];
	attributes: Record<string, string | number | null>;
	tags: NamedRef[];
	sources: SourceRow[];
}

/** A studio's page. */
export interface StudioDetail extends MemberRow {
	aliases: string[];
	details: string | null;
	url: string | null;
	parent: NamedRef | null;
	sources: SourceRow[];
}

/** What a browse or search stored. */
export interface Listed {
	videos: VideoCardRow[];
	galleries: GalleryCardRow[];
	performers: MemberRow[];
	studios: MemberRow[];
}

export const VIDEO_CARD = {
	id: true,
	title: true,
	date: true,
	durationSeconds: true,
	cover: { uri: true },
	preview: { uri: true },
	previewFrames: { uri: true },
	trickplay: {
		sheets: { uri: true },
		columns: true,
		rows: true,
		tileWidth: true,
		tileHeight: true,
		intervalSeconds: true
	},
	width: true,
	height: true,
	studio: { id: true, name: true, image: { uri: true } },
	performers: { id: true, name: true },
	sources: { url: true },
	userState: { progress: true, rating: true, favorite: true, viewCount: true }
} as const;

export const GALLERY_CARD = {
	id: true,
	title: true,
	date: true,
	imageCount: true,
	cover: { uri: true },
	studio: { id: true, name: true },
	sources: { url: true }
} as const;

export const MEMBER = {
	id: true,
	name: true,
	image: { uri: true },
	videoCount: true,
	galleryCount: true
} as const;

const SOURCE = { plugin: true, url: true, fetchedAt: true } as const;

const LISTED = {
	videos: VIDEO_CARD,
	galleries: GALLERY_CARD,
	performers: MEMBER,
	studios: MEMBER
} as const;

/** Maps a page's filter onto the API's FilterInput. */
function filterInput(filter: MediaFilter): FilterInput {
	const input: FilterInput = {};
	if (filter.text) {
		input.text = filter.text;
	}
	if (filter.performerId) {
		input.performerIds = [filter.performerId];
	}
	if (filter.studioId) {
		input.studioIds = [filter.studioId];
	}
	if (filter.tagId) {
		input.includeTagIds = [filter.tagId];
	}
	return input;
}

/** Newest videos matching filter. */
export async function fetchVideos(
	filter: MediaFilter = {},
	limit = 40,
	offset = 0
): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		videos: { __args: { filter: filterInput(filter), limit, offset }, ...VIDEO_CARD }
	});
	return data.videos;
}

/** Videos matching a full API filter, e.g. a saved filter's. */
export async function fetchVideosMatching(
	filter: FilterInput,
	limit = 40,
	offset = 0
): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		videos: { __args: { filter, limit, offset }, ...VIDEO_CARD }
	});
	return data.videos;
}

/** Newest galleries matching filter. */
export async function fetchGalleries(
	filter: MediaFilter = {},
	limit = 40,
	offset = 0
): Promise<GalleryCardRow[]> {
	const data = await gqlClient.query({
		galleries: { __args: { filter: filterInput(filter), limit, offset }, ...GALLERY_CARD }
	});
	return data.galleries;
}

/** Performers by name. */
export async function fetchPerformers(text = '', limit = 60, offset = 0): Promise<MemberRow[]> {
	const data = await gqlClient.query({
		performers: { __args: { text, limit, offset }, ...MEMBER }
	});
	return data.performers;
}

/** Studios by name. */
export async function fetchStudios(text = '', limit = 60, offset = 0): Promise<MemberRow[]> {
	const data = await gqlClient.query({ studios: { __args: { text, limit, offset }, ...MEMBER } });
	return data.studios;
}

/**
 * Opens a video: reads the pages behind its sources that haven't been read
 * yet (credits, description, streams) and returns it.
 */
export async function openVideo(id: string): Promise<VideoDetail | null> {
	const data = await gqlClient.mutation({
		fetchVideo: {
			__args: { id },
			...VIDEO_CARD,
			details: true,
			performers: { id: true, name: true, image: { uri: true } },
			tags: { id: true, name: true },
			sources: SOURCE,
			streams: { id: true, url: true, mimeType: true, quality: true, source: { plugin: true } }
		}
	});
	return data.fetchVideo ?? null;
}

/** Opens a gallery: reads its pages (they list the images) and returns it. */
export async function openGallery(id: string): Promise<GalleryDetail | null> {
	const data = await gqlClient.mutation({
		fetchGallery: {
			__args: { id },
			...GALLERY_CARD,
			details: true,
			performers: { id: true, name: true },
			tags: { id: true, name: true },
			images: { id: true, position: true, asset: { uri: true } },
			sources: SOURCE
		}
	});
	return data.fetchGallery ?? null;
}

/** One performer's page. */
export async function fetchPerformer(id: string): Promise<PerformerDetail | null> {
	const data = await gqlClient.query({
		performer: {
			__args: { id },
			...MEMBER,
			aliases: true,
			details: true,
			links: true,
			attributes: {
				gender: true,
				birthdate: true,
				country: true,
				ethnicity: true,
				eyeColor: true,
				hairColor: true,
				heightCm: true,
				weightKg: true,
				measurements: true,
				tattoos: true,
				piercings: true,
				careerStart: true,
				careerEnd: true
			},
			tags: { id: true, name: true },
			sources: SOURCE
		}
	});
	return data.performer ?? null;
}

/** One studio's page. */
export async function fetchStudio(id: string): Promise<StudioDetail | null> {
	const data = await gqlClient.query({
		studio: {
			__args: { id },
			...MEMBER,
			aliases: true,
			details: true,
			url: true,
			parent: { id: true, name: true },
			sources: SOURCE
		}
	});
	return data.studio ?? null;
}

/** Stores one page of a plugin's catalog (no query) or search results. */
export async function browse(
	plugin: string,
	kind: ListKind,
	query = '',
	limit = 40,
	offset = 0
): Promise<Listed> {
	const data = await gqlClient.mutation({
		browse: { __args: { plugin, kind, query, limit, offset }, ...LISTED }
	});
	return data.browse;
}

/** Searches every plugin that lists this kind and stores the results. */
export async function search(kind: ListKind, query: string, limit = 20): Promise<Listed> {
	const data = await gqlClient.mutation({ search: { __args: { kind, query, limit }, ...LISTED } });
	return data.search;
}

/** Resolves a stream to a URL the player can load. */
export async function play(url: string): Promise<{ url: string; mimeType: string } | null> {
	const data = await gqlClient.query({ play: { __args: { url }, url: true, mimeType: true } });
	if (!data.play) return null;
	return { url: backendUrl(data.play.url), mimeType: data.play.mimeType };
}
