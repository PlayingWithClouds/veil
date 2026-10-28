import { test, expect, describe } from "bun:test";
import { parse } from "node-html-parser";
import { commentText, extractComments } from "../src/comments.ts";

/** A trimmed pornhub comment block. */
function block(id: string, author: string, date: string, text: string, votes: string): string {
  return `<div class="commentBlock closestComment commentTag${id} clearfix">
    <div class="usernameWrap"><a class="usernameLink" href="/users/x">${author}</a></div>
    <div class="date">  ${date}  </div>
    <div class="commentMessage"><span>${text}</span>
      <div class="actionButtonsBlock"><span class="rightMargin voteTotal">${votes}</span></div></div>
  </div>`;
}

describe("pornhub comments", () => {
  const html = `
    ${block("aaa", "alice", "3 months ago", "Nice &amp; hot<br>video", "+4")}
    <div class="nestedBlock childrenOfaaa"><div class="clearfix">
      ${block("bbb", "bob", "2 days ago", "Thanks <img src=\"x.png\" alt=\":)\">", "0")}
    </div></div>
    ${block("ccc", "carol", "1 year ago", "   ", "1")}
    <div class="commentBlock commentTag[[id]]"><div class="commentMessage"><span>[[commentMessage]]</span></div></div>
    ${block("ddd", "dave", "just now", "last", "-2")}`;

  test("reads author, date, text and votes, skipping empty and template blocks", () => {
    const comments = extractComments(parse(html));
    expect(comments.map((comment) => comment.text)).toEqual(["Nice & hot video", "Thanks :)", "last"]);
    expect(comments[0].author).toBe("alice");
    expect(comments[0].likes).toBe(4);
    expect(comments[0].date).toBeDefined();
    expect(comments[2].likes).toBe(-2);
  });

  test("a block inside a nestedBlock answers the comment its childrenOf class names", () => {
    const comments = extractComments(parse(html));
    expect(comments[0].reply_to).toBeUndefined();
    expect(comments[1].reply_to).toBe(0);
    expect(comments[2].reply_to).toBeUndefined();
  });

  test("commentText turns emoticons into alt text and decodes entities", () => {
    expect(commentText('a&nbsp;<img src="e.png" alt=":P"/> &#39;b&#x27;&lt;')).toBe("a :P 'b'<");
  });
});
