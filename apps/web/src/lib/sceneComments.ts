// The comments under a video on its sites, as stored when the scene's page was
// fetched. Viewers often name the performer there.
import { gqlClient } from '$lib/veil';

/** A library performer a comment names. */
export interface CommentMention {
	id: string;
	name: string;
}

export interface SceneComment {
	// "<plugin>:<position>", unique within the scene.
	id: string;
	plugin: string;
	author: string | null;
	text: string;
	postedAt: string | null;
	likes: number | null;
	// id of the comment this one answers.
	replyTo: string | null;
	mentions: CommentMention[];
}

const COMMENT_FIELDS = {
	id: true,
	plugin: true,
	author: true,
	text: true,
	postedAt: true,
	likes: true,
	replyTo: true,
	mentions: { id: true, name: true }
} as const;

/** The scene's stored comments, each site's in its order. */
export async function fetchSceneComments(sceneId: string): Promise<SceneComment[]> {
	const data = await gqlClient.query({
		scene: { __args: { id: sceneId }, comments: COMMENT_FIELDS }
	});
	return (data.scene?.comments ?? []) as SceneComment[];
}

/** Reads the scene's page on its site again and returns the comments it has now. */
export async function refreshSceneComments(sceneId: string): Promise<SceneComment[]> {
	const data = await gqlClient.mutation({
		refreshSceneComments: { __args: { sceneId }, ...COMMENT_FIELDS }
	});
	return data.refreshSceneComments as SceneComment[];
}

/** A run of comment text, linked to a performer when it names one. */
export interface CommentSegment {
	text: string;
	performer: CommentMention | null;
}

/** Splits a comment's text so every occurrence of a mentioned performer's name can be linked. */
export function commentSegments(text: string, mentions: CommentMention[]): CommentSegment[] {
	const lower = text.toLowerCase();
	const ranges: { start: number; end: number; performer: CommentMention }[] = [];
	for (const performer of mentions) {
		const name = performer.name.toLowerCase();
		let start = lower.indexOf(name);
		while (start >= 0) {
			const end = start + name.length;
			if (!ranges.some((range) => start < range.end && end > range.start)) {
				ranges.push({ start, end, performer });
			}
			start = lower.indexOf(name, end);
		}
	}
	ranges.sort((first, second) => first.start - second.start);
	const segments: CommentSegment[] = [];
	let position = 0;
	for (const range of ranges) {
		if (range.start > position) segments.push({ text: text.slice(position, range.start), performer: null });
		segments.push({ text: text.slice(range.start, range.end), performer: range.performer });
		position = range.end;
	}
	if (position < text.length) segments.push({ text: text.slice(position), performer: null });
	return segments;
}

/** A top-level comment with the replies under it, in site order. */
export interface CommentThread {
	comment: SceneComment;
	replies: SceneComment[];
}

/** Groups comments into threads; a reply whose comment is missing stands on its own. */
export function commentThreads(comments: SceneComment[]): CommentThread[] {
	const threads: CommentThread[] = [];
	const threadOf = new Map<string, CommentThread>();
	for (const comment of comments) {
		const parent = comment.replyTo ? threadOf.get(comment.replyTo) : undefined;
		if (parent) {
			parent.replies.push(comment);
			threadOf.set(comment.id, parent);
			continue;
		}
		const thread = { comment, replies: [] };
		threads.push(thread);
		threadOf.set(comment.id, thread);
	}
	return threads;
}

/** Words that mark a comment as asking or telling who someone is. */
const IDENTITY_PATTERN = /\b(who\s+is|who'?s|name\b|names\b|her\s+name|his\s+name|sauce|source|id\s*\?|identify|she\s+is\s+called|known\s+as|aka\b)/i;

/** Whether a comment asks or tells who is in the video, or names a known performer. */
export function isAboutIdentity(comment: SceneComment): boolean {
	if (comment.mentions.length > 0) return true;
	return IDENTITY_PATTERN.test(comment.text);
}
