import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractComments } from "../src/comments.ts";

/** A trimmed eporner comment box at the given indent. */
function box(id: number, indent: number, author: string, time: string, textHtml: string, votes: string): string {
  return `<div class="commbox" style="margin-left:${indent}px" id="c${id}" itemprop="comment">
    <span class="commboxtime f16" itemprop="commentTime"><span><a itemprop="creator">${author}</a></span>
      <div class="flag us"></div> <small>${time}</small></span>
    ${textHtml}
    <div class="commboxvotebar"><div class="commboxvote"> <span class="votegreen">${votes}</span></div></div></div>`;
}

const html = `
  ${box(1, 0, "Creamking", "3 weeks ago", `<p itemprop="commentText" id="ct1">Cameraman&#039;s trash..</p>`, "+6")}
  ${box(2, 20, "TheGame", "2 days ago", `<p itemprop="commentText" id="ct2">reply one</p>`, "0")}
  ${box(3, 40, "Third", "1 day ago", `<p itemprop="commentText" id="ct3">reply to reply</p>`, "1")}
  ${box(4, 20, "Fourth", "1 day ago", `<p itemprop="commentText" id="ct4">second reply</p>`, "1")}
  ${box(5, 0, "Hidden", "1 month ago", `<p itemprop="commentText" id="ct5" class="moderated"><span class="comment-hidden-label">[Comment hidden]</span><span class="comment-hidden-text" hidden>real text</span></p>`, "-4")}
  ${box(6, 0, "Empty", "1 month ago", `<p itemprop="commentText" id="ct6"> </p>`, "0")}`;

test("reads author, date, text and votes; skips empty comments", () => {
  const comments = extractComments(parse(html));
  expect(comments.map((comment) => comment.text)).toEqual([
    "Cameraman's trash..",
    "reply one",
    "reply to reply",
    "second reply",
    "real text",
  ]);
  expect(comments[0].author).toBe("Creamking");
  expect(comments[0].likes).toBe(6);
  expect(comments[0].date).toBeDefined();
  expect(comments[4].likes).toBe(-4);
});

test("indented comments answer the nearest earlier comment with a smaller indent", () => {
  const comments = extractComments(parse(html));
  expect(comments.map((comment) => comment.reply_to)).toEqual([undefined, 0, 1, 0, undefined]);
});
