import type { HTMLElement } from "node-html-parser";
import { relativeDate, type SceneComment } from "@playingwithclouds/veil-sdk";

const MAX_COMMENTS = 100;

/**
 * The comments Pornhub renders under a video, in page order. Replies sit in a
 * `.nestedBlock.childrenOf<id>` that follows the comment `commentTag<id>`
 * they answer. RedTube, YouPorn and Tube8 load comments with scripts, so
 * their pages carry none.
 */
export function extractComments(doc: HTMLElement): SceneComment[] {
  const comments: SceneComment[] = [];
  const indexByCommentId = new Map<string, number>();
  for (const block of doc.querySelectorAll(".commentBlock")) {
    if (comments.length >= MAX_COMMENTS) break;
    const commentId = classSuffix(block, "commentTag");
    // The page's reply template carries "[[id]]" placeholders instead of data.
    if (commentId === "" || commentId.includes("[[")) continue;
    const comment = readComment(block);
    if (!comment) continue;
    const parentIndex = parentCommentIndex(block, indexByCommentId);
    if (parentIndex !== undefined) comment.reply_to = parentIndex;
    indexByCommentId.set(commentId, comments.length);
    comments.push(comment);
  }
  return comments;
}

/** Index of the comment a reply block answers, from the enclosing `childrenOf<id>` block. */
function parentCommentIndex(block: HTMLElement, indexByCommentId: Map<string, number>): number | undefined {
  const nested = block.closest(".nestedBlock");
  if (!nested) return undefined;
  return indexByCommentId.get(classSuffix(nested, "childrenOf"));
}

/** The rest of the class token that starts with `prefix` ("commentTagabc" gives "abc"), or "". */
function classSuffix(element: HTMLElement, prefix: string): string {
  for (const token of (element.getAttribute("class") || "").split(/\s+/)) {
    if (token.startsWith(prefix)) return token.slice(prefix.length);
  }
  return "";
}

/** One comment block as a SceneComment; undefined when its text is empty. */
function readComment(block: HTMLElement): SceneComment | undefined {
  const messageElement = block.querySelector(".commentMessage span");
  if (!messageElement) return undefined;
  const text = commentText(messageElement.innerHTML);
  if (text === "") return undefined;

  const comment: SceneComment = { text };
  const author = block.querySelector(".usernameLink");
  if (author && author.text.trim() !== "") comment.author = author.text.trim();
  const dateElement = block.querySelector(".date");
  if (dateElement) {
    const date = relativeDate(dateElement.text);
    if (date) comment.date = date;
  }
  const votes = block.querySelector(".voteTotal");
  if (votes) {
    const likes = parseInt(votes.text.trim(), 10);
    if (!Number.isNaN(likes)) comment.likes = likes;
  }
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
