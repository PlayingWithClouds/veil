// Media beyond videos and galleries: audio, stories, creator posts, and the
// sites' own groups of videos (movies, series, compilations).
import { gqlClient } from '$lib/veil';
import type { NamedRef } from '$lib/library';

/** An audio as a card shows it. */
export interface AudioCardRow {
	id: string;
	title: string;
	date: string | null;
	durationSeconds: number | null;
	cover: { uri: string } | null;
	studio: (NamedRef & { image: { uri: string } | null }) | null;
	performers: (NamedRef & { image: { uri: string } | null })[];
	sources: { url: string }[];
}

/** A story as a card shows it. */
export interface StoryCardRow {
	id: string;
	title: string;
	date: string | null;
	chapterCount: number;
	complete: boolean;
	cover: { uri: string } | null;
	performers: NamedRef[];
	sources: { url: string }[];
}

/** A creator post as a card shows it. */
export interface PostCardRow {
	id: string;
	title: string;
	date: string | null;
	// Paywalled for the logged-in account.
	locked: boolean;
	cover: { uri: string } | null;
	creator: (NamedRef & { image: { uri: string } | null }) | null;
	sources: { url: string }[];
}

/** A site group (movie, series, compilation) as a card shows it. */
export interface GroupCardRow {
	id: string;
	title: string;
	date: string | null;
	cover: { uri: string } | null;
	studio: NamedRef | null;
}

const CREDIT = { id: true, name: true, image: { uri: true } } as const;

const AUDIO_CARD = {
	id: true,
	title: true,
	date: true,
	durationSeconds: true,
	cover: { uri: true },
	studio: CREDIT,
	performers: CREDIT,
	sources: { url: true }
} as const;

const STORY_CARD = {
	id: true,
	title: true,
	date: true,
	chapterCount: true,
	complete: true,
	cover: { uri: true },
	performers: { id: true, name: true },
	sources: { url: true }
} as const;

const POST_CARD = {
	id: true,
	title: true,
	date: true,
	locked: true,
	cover: { uri: true },
	creator: CREDIT,
	sources: { url: true }
} as const;

const GROUP_CARD = {
	id: true,
	title: true,
	date: true,
	cover: { uri: true },
	studio: { id: true, name: true }
} as const;

/** Newest audio, sites taking turns. */
export async function fetchAudios(limit = 20, offset = 0): Promise<AudioCardRow[]> {
	const data = await gqlClient.query({ audios: { __args: { limit, offset }, ...AUDIO_CARD } });
	return data.audios;
}

/** Newest stories, sites taking turns. */
export async function fetchStories(limit = 20, offset = 0): Promise<StoryCardRow[]> {
	const data = await gqlClient.query({ stories: { __args: { limit, offset }, ...STORY_CARD } });
	return data.stories;
}

/** Newest creator posts, sites taking turns. */
export async function fetchPosts(limit = 20, offset = 0): Promise<PostCardRow[]> {
	const data = await gqlClient.query({ posts: { __args: { limit, offset }, ...POST_CARD } });
	return data.posts;
}

/** Newest site groups. */
export async function fetchGroups(limit = 20, offset = 0): Promise<GroupCardRow[]> {
	const data = await gqlClient.query({ groups: { __args: { limit, offset }, ...GROUP_CARD } });
	return data.groups;
}
