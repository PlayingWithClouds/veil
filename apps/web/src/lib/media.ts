// Media beyond videos and galleries: audio, stories, creator posts, and the
// sites' own groups of videos (movies, series, compilations).
import type {
	Audio,
	FieldsSelection,
	Group,
	Post,
	Story,
	StoryChapter
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { SOURCE, STREAM } from '$lib/library';

const CREDIT = { id: true, name: true, image: { uri: true } } as const;
const NAMED = { id: true, name: true } as const;

/** An audio as a card shows it. */
export const AUDIO_CARD = {
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
export const STORY_CARD = {
	id: true,
	title: true,
	date: true,
	chapterCount: true,
	complete: true,
	cover: { uri: true },
	performers: { id: true, name: true },
	sources: { url: true }
} as const;

/** An audio's page: what the card shows, its description, tags, streams and resume point. */
const AUDIO_DETAIL = {
	...AUDIO_CARD,
	details: true,
	tags: NAMED,
	sources: SOURCE,
	streams: STREAM,
	userState: { progress: true }
} as const;

/** A chapter as the chapter list and the reader's navigation know it, without its text. */
const CHAPTER = {
	id: true,
	position: true,
	title: true,
	wordCount: true,
	parent: true,
	choice: true
} as const;

/** A story's page: what the card shows, its description, chapters and where reading stopped. */
const STORY_DETAIL = {
	...STORY_CARD,
	details: true,
	wordCount: true,
	studio: NAMED,
	tags: NAMED,
	sources: SOURCE,
	chapters: CHAPTER,
	userState: { progress: true }
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
export type AudioDetail = FieldsSelection<Audio, typeof AUDIO_DETAIL>;
export type StoryCardRow = FieldsSelection<Story, typeof STORY_CARD>;
export type StoryDetail = FieldsSelection<Story, typeof STORY_DETAIL>;
export type ChapterRow = FieldsSelection<StoryChapter, typeof CHAPTER>;
export type ChapterText = ChapterRow & { text: string };
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

/** Stories the user stopped partway, most recently read first. */
export async function fetchContinueReading(limit = 20): Promise<StoryCardRow[]> {
	const data = await gqlClient.query({
		history: {
			__args: { kind: 'story', limit: limit * 2 },
			progress: true,
			record: { on_Story: { __typename: true, ...STORY_CARD } }
		}
	});
	const started = data.history.flatMap((state) => {
		if (!state.progress || state.record?.__typename !== 'Story') return [];
		return [state.record];
	});
	return started.slice(0, limit);
}

/** Opens an audio, reading the pages behind its sources that weren't read yet (credits, streams). */
export async function openAudio(id: string): Promise<AudioDetail | null> {
	const data = await gqlClient.mutation({
		fetchMedia: {
			__args: { subject: { kind: 'audio', id } },
			on_Audio: { __typename: true, ...AUDIO_DETAIL }
		}
	});
	if (data.fetchMedia?.__typename !== 'Audio') return null;
	return data.fetchMedia;
}

/** Opens a story, reading the pages behind its sources that weren't read yet (chapters, credits). */
export async function openStory(id: string): Promise<StoryDetail | null> {
	const data = await gqlClient.mutation({
		fetchMedia: {
			__args: { subject: { kind: 'story', id } },
			on_Story: { __typename: true, ...STORY_DETAIL }
		}
	});
	if (data.fetchMedia?.__typename !== 'Story') return null;
	return data.fetchMedia;
}

/** A story's chapters, again: reading a chapter of a branching story can add the ones its choices lead to. */
export async function fetchChapters(storyId: string): Promise<ChapterRow[]> {
	const data = await gqlClient.query({ story: { __args: { id: storyId }, chapters: CHAPTER } });
	return data.story?.chapters ?? [];
}

/** A chapter with its text, read from the site the first time. */
export async function readChapter(storyId: string, position: number): Promise<ChapterText | null> {
	const data = await gqlClient.mutation({
		readChapter: { __args: { storyId, position }, ...CHAPTER, text: true }
	});
	return data.readChapter ?? null;
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
