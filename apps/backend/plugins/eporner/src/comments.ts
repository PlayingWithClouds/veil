import type { HTMLElement } from "node-html-parser";
import { relativeDate, type SceneComment } from "@playingwithclouds/veil-sdk";

const MAX_COMMENTS = 100;
/** Indent per reply level in the comment's inline `margin-left`. */
const INDENT_STEP_PIXELS = 20;

/**
 * The comments under an eporner video page, in page order. A reply is
 * indented further (`style="margin-left:20px"` per level) than the comment it
 * answers and follows it, so it belongs to the nearest earlier comment with a
 * smaller indent.
 */
export function extractComments(doc: HTMLElement): SceneComment[] {
  const comments: SceneComment[] = [];
  // Index and indent of the latest comment on each level above the current one.
  const ancestors: { index: number; indent: number }[] = [];
  for (const box of doc.querySelectorAll(".commbox")) {
    if (comments.length >= MAX_COMMENTS) break;
    const comment = readComment(box);
    if (!comment) continue;
    const indent = indentOf(box);
    while (ancestors.length > 0 && ancestors[ancestors.length - 1].indent >= indent) ancestors.pop();
    if (ancestors.length > 0) comment.reply_to = ancestors[ancestors.length - 1].index;
    ancestors.push({ index: comments.length, indent });
    comments.push(comment);
  }
  return comments;
}

/** The comment's left indent in pixels, 0 when unset. */
function indentOf(box: HTMLElement): number {
  const match = (box.getAttribute("style") || "").match(/margin-left:\s*(\d+)/);
  if (!match) return 0;
  return Math.floor(parseInt(match[1], 10) / INDENT_STEP_PIXELS) * INDENT_STEP_PIXELS;
}

/** One `.commbox` as a SceneComment; undefined when it has no text. */
function readComment(box: HTMLElement): SceneComment | undefined {
  const textElement = box.querySelector('[itemprop="commentText"]');
  if (!textElement) return undefined;
  // Comments voted down are collapsed behind a label; the real text stays in the page.
  const hiddenText = textElement.querySelector(".comment-hidden-text");
  let html = textElement.innerHTML;
  if (hiddenText) html = hiddenText.innerHTML;
  const text = commentText(html);
  if (text === "") return undefined;

  const comment: SceneComment = { text };
  const author = box.querySelector('[itemprop="creator"]');
  if (author && author.text.trim() !== "") comment.author = author.text.trim();
  const dateElement = box.querySelector(".commboxtime small");
  if (dateElement) {
    const date = relativeDate(dateElement.text);
    if (date) comment.date = date;
  }
  const votes = box.querySelector(".commboxvote span");
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
