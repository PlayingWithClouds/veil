import type { HTMLElement } from "node-html-parser";
import { relativeDate, type SceneComment } from "@playingwithclouds/veil-sdk";

const MAX_COMMENTS = 100;

/**
 * The comments a KVS video page renders as `[data-comment-id]` items (the
 * stock theme markup, kept by camstreams, cloudbate and whoreshub), in page
 * order. Stock KVS comments are flat, so none is a reply. Themes that load
 * their comments with scripts leave the list empty in the fetched page.
 */
export function extractComments(root: HTMLElement): SceneComment[] {
  const comments: SceneComment[] = [];
  for (const item of root.querySelectorAll("[data-comment-id]")) {
    if (comments.length >= MAX_COMMENTS) break;
    const comment = readComment(item);
    if (comment) comments.push(comment);
  }
  return comments;
}

/** One comment item as a SceneComment; undefined when it has no text. */
function readComment(item: HTMLElement): SceneComment | undefined {
  // `.original-text` is the comment without the theme's quote marks around it.
  let textElement = item.querySelector(".original-text");
  if (!textElement) textElement = item.querySelector(".comment-text");
  if (!textElement) return undefined;
  const text = commentText(textElement.innerHTML);
  if (text === "") return undefined;

  const comment: SceneComment = { text };
  const authorName = commentAuthor(item);
  if (authorName !== "") comment.author = authorName;
  const date = commentDate(item, authorName);
  if (date) comment.date = date;
  const rating = item.querySelector(".comment-rating");
  if (rating) {
    const likes = parseInt(rating.text.trim(), 10);
    if (!Number.isNaN(likes)) comment.likes = likes;
  }
  return comment;
}

/** The commenter's user name (`.username`, or `.name` in the whoreshub theme), "" when anonymous. */
function commentAuthor(item: HTMLElement): string {
  const author = item.querySelector(".username") || item.querySelector("a.name");
  if (!author) return "";
  return author.text.trim();
}

/**
 * The comment's date: the text of `.comment-info` after the user name (stock
 * theme) or `.added` (whoreshub), as an ISO date for relative times
 * ("2 months ago") and absolute ones ("2026-10-28 12:00").
 */
function commentDate(item: HTMLElement, authorName: string): string | undefined {
  const dateElement = item.querySelector(".comment-info") || item.querySelector(".added");
  if (!dateElement) return undefined;
  let dateText = dateElement.text;
  if (authorName !== "") dateText = dateText.replace(authorName, " ");
  dateText = dateText.replace(/\s+/g, " ").trim();
  const relative = relativeDate(dateText);
  if (relative) return relative;
  const absolute = dateText.match(/\d{4}-\d{2}-\d{2}(?:[ T]\d{2}:\d{2}(?::\d{2})?)?/);
  if (!absolute) return undefined;
  const parsed = new Date(absolute[0].replace(" ", "T") + (absolute[0].length > 10 ? "Z" : ""));
  if (Number.isNaN(parsed.getTime())) return undefined;
  return parsed.toISOString();
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
