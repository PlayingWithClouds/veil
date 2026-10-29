import { relativeDate } from "@playingwithclouds/veil-sdk";
import type { SceneComment } from "@playingwithclouds/veil-sdk";
import { decodeEntities, fetchJson, ORIGINS, type Site } from "./http.ts";

/** One post of the comments endpoint; replies nest the same shape. */
interface ThreadPost {
  name?: string;
  message?: string;
  time_diff?: string;
  votes?: { nb?: number };
  replies?: ThreadPosts;
}

/** A page of posts: their ids in site order and the posts by id. */
interface ThreadPosts {
  ids?: string[];
  posts?: Record<string, ThreadPost> | ThreadPost[];
}

/**
 * The top comments under a video (first page, replies after the comment they
 * answer). Both sites load them from the same endpoint by the numeric video
 * id. A failing request yields none: comments are a bonus to the scene.
 */
export async function fetchComments(site: Site, numericId: string, pageUrl: string): Promise<SceneComment[]> {
  try {
    const answer = await fetchJson(`${ORIGINS[site]}/threads/video-comments/get-posts/top/${numericId}/0/0`, pageUrl);
    return parseComments(answer);
  } catch {
    return [];
  }
}

/** Flattens the endpoint's answer into Scene.comments, each reply pointing at its comment. */
export function parseComments(answer: { posts?: ThreadPosts }): SceneComment[] {
  const comments: SceneComment[] = [];
  for (const post of orderedPosts(answer.posts)) {
    const comment = toComment(post);
    if (!comment) continue;
    const index = comments.length;
    comments.push(comment);
    for (const reply of orderedPosts(post.replies)) {
      const answerComment = toComment(reply);
      if (answerComment) comments.push({ ...answerComment, reply_to: index });
    }
  }
  return comments;
}

/** The posts of a page in the order of its ids. */
function orderedPosts(page: ThreadPosts | undefined): ThreadPost[] {
  if (!page || !page.posts || Array.isArray(page.posts)) return [];
  const byId = page.posts;
  return (page.ids || []).map((id) => byId[id]).filter((post): post is ThreadPost => post !== undefined);
}

/** One post as a comment: HTML stripped, relative date made absolute. */
function toComment(post: ThreadPost): SceneComment | undefined {
  const text = plainText(post.message || "");
  if (text === "") return undefined;
  const comment: SceneComment = { text };
  if (post.name) comment.author = post.name;
  const date = relativeDate(post.time_diff || "");
  if (date) comment.date = date;
  if (typeof post.votes?.nb === "number") comment.likes = post.votes.nb;
  return comment;
}

/** A message as plain text: line breaks kept, tags dropped, emoticon images as their alt text. */
function plainText(message: string): string {
  return decodeEntities(
    message
      .replace(/<br\s*\/?>/gi, "\n")
      .replace(/<img[^>]*\balt="([^"]*)"[^>]*>/gi, "$1")
      .replace(/<[^>]+>/g, ""),
  ).trim();
}
