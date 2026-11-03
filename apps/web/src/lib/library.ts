// Reads of the core records. Each row type is derived from the selection
// that fetches it, so it can't drift from the API.
import type {
	FieldsSelection,
	FilterInput,
	Gallery,
	Listed as ListedResult,
	Performer,
	Source,
	Stream,
	Studio,
	Tag,
	Trickplay as TrickplayResult,
	Video
} from '@playingwithclouds/veil-sdk/generated';
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

const NAMED = { id: true, name: true } as const;

const TRICKPLAY = {
	sheets: { uri: true },
	columns: true,
	rows: true,
	tileWidth: true,
	tileHeight: true,
	intervalSeconds: true
} as const;

/** A video as a card shows it. previewFrames are the stills its card scrubs through. */
export const VIDEO_CARD = {
	id: true,
	title: true,
	date: true,
	durationSeconds: true,
	cover: { uri: true },
	preview: { uri: true },
	previewFrames: { uri: true },
	trickplay: TRICKPLAY,
	width: true,
	height: true,
	studio: { ...NAMED, image: { uri: true } },
	performers: NAMED,
	sources: { url: true },
	// progress is the resume point in seconds; rating is 0..10.
	userState: { progress: true, rating: true, favorite: true, viewCount: true }
} as const;

/** A gallery as a card shows it. */
export const GALLERY_CARD = {
	id: true,
	title: true,
	date: true,
	imageCount: true,
	cover: { uri: true },
	studio: NAMED,
	performers: NAMED,
	sources: { url: true },
	userState: { favorite: true, rating: true }
} as const;

/** A performer or studio as a card shows it. */
export const MEMBER = {
	id: true,
	name: true,
	image: { uri: true },
	videoCount: true,
	galleryCount: true
} as const;

/** One site a record was found on. */
export const SOURCE = { plugin: true, url: true, fetchedAt: true } as const;

export const STREAM = {
	id: true,
	url: true,
	mimeType: true,
	quality: true,
	source: { plugin: true }
} as const;

/** A comment from a site, as the video page lists it. */
const COMMENT = {
	id: true,
	author: true,
	text: true,
	postedAt: true,
	likes: true,
	replyTo: true,
	source: { plugin: true }
} as const;

/** Everything the video page shows. */
const VIDEO_DETAIL = {
	...VIDEO_CARD,
	details: true,
	studio: { ...NAMED, image: { uri: true } },
	performers: MEMBER,
	suggestedPerformers: { performer: MEMBER, reason: true },
	tags: NAMED,
	sources: SOURCE,
	streams: STREAM,
	// Local copies (downloads, cam recordings); they play before any site's stream.
	files: { id: true, url: true, mimeType: true, quality: true, missing: true },
	segments: { id: true, startSeconds: true, endSeconds: true, title: true, tag: NAMED },
	comments: COMMENT,
	gallery: { id: true, imageCount: true, cover: { uri: true } }
} as const;

/** Everything the gallery page shows. */
const GALLERY_DETAIL = {
	...GALLERY_CARD,
	details: true,
	performers: { id: true, name: true, image: { uri: true } },
	tags: NAMED,
	images: { id: true, position: true, width: true, height: true, asset: { uri: true } },
	video: { id: true, title: true },
	sources: SOURCE,
	userState: { favorite: true, rating: true }
} as const;

/** A studio's page. */
const STUDIO_DETAIL = {
	...MEMBER,
	aliases: true,
	details: true,
	url: true,
	parent: NAMED,
	sources: SOURCE
} as const;

/** What a browse or search stored. */
const LISTED = {
	videos: VIDEO_CARD,
	galleries: GALLERY_CARD,
	performers: MEMBER,
	studios: MEMBER,
	// Counted only: their pages load their own cards.
	audios: { id: true },
	stories: { id: true }
} as const;

/** A named reference: a performer, studio or tag. */
export type NamedRef = FieldsSelection<Tag, typeof NAMED>;
export type Trickplay = FieldsSelection<TrickplayResult, typeof TRICKPLAY>;
export type VideoCardRow = FieldsSelection<Video, typeof VIDEO_CARD>;
export type GalleryCardRow = FieldsSelection<Gallery, typeof GALLERY_CARD>;
export type MemberRow = FieldsSelection<Performer, typeof MEMBER>;
export type SourceRow = FieldsSelection<Source, typeof SOURCE>;
export type StreamRow = FieldsSelection<Stream, typeof STREAM>;
export type VideoDetail = FieldsSelection<Video, typeof VIDEO_DETAIL>;
export type GalleryDetail = FieldsSelection<Gallery, typeof GALLERY_DETAIL>;
export type StudioDetail = FieldsSelection<Studio, typeof STUDIO_DETAIL>;
export type Listed = FieldsSelection<ListedResult, typeof LISTED>;

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

/** Adds a performer by name, or returns the one already called that. */
export async function createPerformer(name: string): Promise<MemberRow> {
	const data = await gqlClient.mutation({ createPerformer: { __args: { name }, ...MEMBER } });
	return data.createPerformer;
}

/** Adds a studio by name, or returns the one already called that. */
export async function createStudio(name: string): Promise<MemberRow> {
	const data = await gqlClient.mutation({ createStudio: { __args: { name }, ...MEMBER } });
	return data.createStudio;
}

/**
 * Opens a video: reads the pages behind its sources that haven't been read
 * yet (credits, description, streams) and returns it.
 */
export async function openVideo(id: string): Promise<VideoDetail | null> {
	const data = await gqlClient.mutation({ fetchVideo: { __args: { id }, ...VIDEO_DETAIL } });
	return data.fetchVideo ?? null;
}

/** The video as stored, without reading its pages again. */
export async function fetchVideoDetail(id: string): Promise<VideoDetail | null> {
	const data = await gqlClient.query({ video: { __args: { id }, ...VIDEO_DETAIL } });
	return data.video ?? null;
}

/**
 * Sets who's in a video, as the user: the credits are locked, so the sites
 * leave them as set.
 */
export async function setVideoPerformers(
	id: string,
	performerIds: string[]
): Promise<VideoDetail | null> {
	const data = await gqlClient.mutation({
		editVideo: { __args: { id, edit: { performerIds } }, ...VIDEO_DETAIL }
	});
	return data.editVideo ?? null;
}

/** Sets a video's studio as the user; null clears it. */
export async function setVideoStudio(
	id: string,
	studioId: string | null
): Promise<VideoDetail | null> {
	const data = await gqlClient.mutation({
		editVideo: { __args: { id, edit: { studioId: studioId ?? '' } }, ...VIDEO_DETAIL }
	});
	return data.editVideo ?? null;
}

/** Videos like this one, most alike first. */
export async function fetchRelatedVideos(id: string, limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		related: {
			__args: { subject: { kind: 'video', id }, limit },
			record: { on_Video: { __typename: true, ...VIDEO_CARD } }
		}
	});
	return data.related.flatMap((match) => {
		if (match.record?.__typename !== 'Video') return [];
		return [match.record];
	});
}

/** Opens a gallery: reads its pages (they list the images) and returns it. */
export async function openGallery(id: string): Promise<GalleryDetail | null> {
	const data = await gqlClient.mutation({ fetchGallery: { __args: { id }, ...GALLERY_DETAIL } });
	return data.fetchGallery ?? null;
}

/** One studio's page. */
export async function fetchStudio(id: string): Promise<StudioDetail | null> {
	const data = await gqlClient.query({ studio: { __args: { id }, ...STUDIO_DETAIL } });
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

/** A page of one site's videos (its search, with query), stored as it comes. */
export async function browseVideos(
	plugin: string,
	limit: number,
	offset: number,
	query = ''
): Promise<VideoCardRow[]> {
	const data = await gqlClient.mutation({
		browse: { __args: { plugin, kind: 'VIDEO', query, limit, offset }, videos: VIDEO_CARD }
	});
	return data.browse.videos;
}

/** Searches every plugin that lists this kind and stores the results. */
export async function search(kind: ListKind, query: string, limit = 20): Promise<Listed> {
	const data = await gqlClient.mutation({ search: { __args: { kind, query, limit }, ...LISTED } });
	return data.search;
}

export type LocalFileRow = VideoDetail['files'][number];

/** A local file as the player plays it, at its backend URL. */
export function playableFile(file: LocalFileRow): { url: string; mimeType: string } {
	return { url: backendUrl(file.url), mimeType: file.mimeType };
}

/** Resolves a stream to a URL the player can load. */
export async function play(url: string): Promise<{ url: string; mimeType: string } | null> {
	const data = await gqlClient.query({ play: { __args: { url }, url: true, mimeType: true } });
	if (!data.play) return null;
	return { url: backendUrl(data.play.url), mimeType: data.play.mimeType };
}
