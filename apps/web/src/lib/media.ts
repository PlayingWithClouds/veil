// Media beyond videos and galleries: audio, stories, creator posts, and the
// sites' own groups of videos (movies, series, compilations).
import type {
	Audio,
	FieldsSelection,
	Group,
	Post,
	Story
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

const CREDIT = { id: true, name: true, image: { uri: true } } as const;

/** An audio as a card shows it. */
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

/** A story as a card shows it. */
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

/** A creator post as a card shows it; locked when paywalled for the logged-in account. */
const POST_CARD = {
	id: true,
	title: true,
	date: true,
	locked: true,
	cover: { uri: true },
	creator: CREDIT,
	sources: { url: true }
} as const;

/** A site group (movie, series, compilation) as a card shows it. */
const GROUP_CARD = {
	id: true,
	title: true,
	date: true,
	cover: { uri: true },
	studio: { id: true, name: true }
} as const;

export type AudioCardRow = FieldsSelection<Audio, typeof AUDIO_CARD>;
export type StoryCardRow = FieldsSelection<Story, typeof STORY_CARD>;
export type PostCardRow = FieldsSelection<Post, typeof POST_CARD>;
export type GroupCardRow = FieldsSelection<Group, typeof GROUP_CARD>;

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
