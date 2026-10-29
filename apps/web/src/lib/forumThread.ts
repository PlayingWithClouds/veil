import { gqlClient } from '$lib/veil';

/** A link a comment gives; scenePlugin names the plugin that can play it in the app. */
export interface ForumLink {
	url: string;
	scenePlugin: string | null;
}

/** A performer a comment tags, with the library's performer of that name. */
export interface ForumName {
	name: string;
	performerId: string | null;
}

/** One comment of a forum thread. */
export interface ForumComment {
	id: string;
	author: string | null;
	/** ISO 8601. */
	date: string | null;
	/** Plain text, line breaks kept. */
	text: string;
	/** Whether this is the thread's accepted answer. */
	accepted: boolean;
	likes: number | null;
	names: ForumName[];
	/** Scene or movie titles the comment tags. */
	titles: string[];
	links: ForumLink[];
	/** Pictures posted in the comment. */
	images: string[];
	/** Other threads of the same forum the comment points to. */
	posts: ForumPostRef[];
}

/** A thread of the same forum a comment points to. */
export interface ForumPostRef {
	title: string;
	url: string;
	thumbnailUrl: string | null;
}

/** One thread of an identification forum: the question and its comments. */
export interface ForumThread {
	url: string;
	title: string;
	description: string | null;
	imageUrl: string | null;
	videoUrl: string | null;
	author: string | null;
	date: string | null;
	tags: string[];
	solved: boolean;
	/** The plugin that read the thread. */
	plugin: string;
	comments: ForumComment[];
}

/** A thread a forum's text search found. */
export interface ForumPostHit {
	title: string;
	url: string;
	thumbnailUrl: string | null;
	date: string | null;
	solved: boolean;
	author: string | null;
	/** The start of the question or answer, as the forum's search shows it. */
	snippet: string | null;
	plugin: string;
}

/** Threads of the identification forums (NameThatPorn) matching a query; none when the forums fail. */
export async function fetchForumSearch(query: string): Promise<ForumPostHit[]> {
	const data = await gqlClient.query({
		forumSearch: {
			__args: { query },
			title: true,
			url: true,
			thumbnailUrl: true,
			date: true,
			solved: true,
			author: true,
			snippet: true,
			plugin: true
		}
	});
	return data.forumSearch as unknown as ForumPostHit[];
}

/** Reads a forum thread live through the plugin serving its URL. */
export async function fetchForumThread(url: string): Promise<ForumThread> {
	const data = await gqlClient.query({
		forumThread: {
			__args: { url },
			url: true,
			title: true,
			description: true,
			imageUrl: true,
			videoUrl: true,
			author: true,
			date: true,
			tags: true,
			solved: true,
			plugin: true,
			comments: {
				id: true,
				author: true,
				date: true,
				text: true,
				accepted: true,
				likes: true,
				names: { name: true, performerId: true },
				titles: true,
				links: { url: true, scenePlugin: true },
				images: true,
				posts: { title: true, url: true, thumbnailUrl: true }
			}
		}
	});
	return data.forumThread as unknown as ForumThread;
}
