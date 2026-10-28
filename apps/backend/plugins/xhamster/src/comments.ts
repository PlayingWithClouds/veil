import type { SceneComment } from "@playingwithclouds/veil-sdk";
import type { RawComment, VideoInitials } from "./initials.ts";

const MAX_COMMENTS = 100;

/**
 * The comments in the page state's `commentsComponent.commentsList`, in the
 * site's order. Replies are flat items whose `replyToId` names the comment
 * they answer; `reply_to` is that comment's index among the kept ones.
 */
export function extractComments(initials: VideoInitials | undefined): SceneComment[] {
  if (!initials || !initials.commentsComponent || !initials.commentsComponent.commentsList) return [];
  const items = initials.commentsComponent.commentsList.items;
  if (!items) return [];

  const kept: { item: RawComment; comment: SceneComment }[] = [];
  for (const item of items) {
    if (kept.length >= MAX_COMMENTS) break;
    const comment = readComment(item);
    if (comment) kept.push({ item, comment });
  }

  const indexById = new Map<string, number>();
  kept.forEach((entry, index) => indexById.set(String(entry.item.id), index));
  for (const entry of kept) {
    if (entry.item.replyToId === null || entry.item.replyToId === undefined) continue;
    const parentIndex = indexById.get(String(entry.item.replyToId));
    if (parentIndex !== undefined) entry.comment.reply_to = parentIndex;
  }
  return kept.map((entry) => entry.comment);
}

/** One page-state comment as a SceneComment; undefined for spam, non-public or empty ones. */
function readComment(item: RawComment): SceneComment | undefined {
  if (item.isSpam) return undefined;
  if (item.status !== undefined && item.status !== "public") return undefined;
  const text = commentText(item.text || "");
  if (text === "") return undefined;

  const comment: SceneComment = { text };
  if (item.author && item.author.name) comment.author = item.author.name;
  if (typeof item.created === "number" && item.created > 0) {
    comment.date = new Date(item.created * 1000).toISOString();
  }
  if (typeof item.likes === "number") comment.likes = item.likes;
  return comment;
}

const ENTITIES: Record<string, string> = { amp: "&", lt: "<", gt: ">", quot: '"', apos: "'", nbsp: " " };

/** Plain text of comment HTML: emoticon images as their alt text, tags dropped, entities decoded, whitespace collapsed. */
export function commentText(html: string): string {
  return html
    .replace(/<img\b[^>]*?\balt=(?:"([^"]*)"|'([^']*)')[^>]*>/gi, (_match, doubleQuoted, singleQuoted) => {
      if (doubleQuoted !== undefined) return ` ${doubleQuoted} `;
      return ` ${singleQuoted} `;
    })
    .replace(/<br\s*\/?>/gi, " ")
    .replace(/<[^>]*>/g, "")
    .replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (entity, body: string) => {
      if (body[0] === "#") {
        const codePoint = body[1].toLowerCase() === "x" ? parseInt(body.slice(2), 16) : parseInt(body.slice(1), 10);
        return String.fromCodePoint(codePoint);
      }
      const named = ENTITIES[body.toLowerCase()];
      if (named === undefined) return entity;
      return named;
    })
    .replace(/\s+/g, " ")
    .trim();
}
