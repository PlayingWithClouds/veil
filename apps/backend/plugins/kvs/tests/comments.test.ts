import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { extractComments } from "../src/comments.ts";

const stockTheme = `
<div id="video_comments_video_comments_items">
  <div class="item " data-comment-id="1">
    <div class="text">
      <span class="comment-options"><span class="comment-rating positive">1</span></span>
      <span class="comment-info">
        <a class="username" href="/members/1/">CassandraB</a>
        2 months ago
      </span>
      <p class="comment-text">&laquo;<span class="original-text"><img src='w.png' alt=':woot:'/><img src='t.png' alt=':P'/> hi &amp; bye<br>second line</span>&raquo;</p>
    </div>
  </div>
  <div class="item dim-comment" data-comment-id="2">
    <div class="text">
      <span class="comment-options"><span class="comment-rating negative">-1</span></span>
      <span class="comment-info"><a class="username" href="/members/2/">Spammer</a> 10 months ago</span>
      <p class="comment-text">plain text without original span</p>
    </div>
  </div>
  <div class="item" data-comment-id="3"><p class="comment-text"><span class="original-text"> </span></p></div>
</div>`;

const whoreshubTheme = `
<div class="item post-wrap" data-comment-id="9">
  <div class="comments-top"><a class="name" href="/members/5/">Analgod69</a>
    <span class="added"><span class="text">13 seconds ago</span></span>
    <span class="comment-rating rating positive">2</span></div>
  <div class="comments-center"><p class="comment-text"><span class="original-text">works</span></p></div>
</div>`;

test("reads the stock theme's author, relative date, text and rating", () => {
  const comments = extractComments(parse(stockTheme));
  expect(comments).toHaveLength(2);
  expect(comments[0].text).toBe(":woot: :P hi & bye second line");
  expect(comments[0].author).toBe("CassandraB");
  expect(comments[0].likes).toBe(1);
  expect(comments[0].date).toBeDefined();
  expect(comments[1].text).toBe("plain text without original span");
  expect(comments[1].likes).toBe(-1);
  expect(comments.every((comment) => comment.reply_to === undefined)).toBe(true);
});

test("reads the whoreshub variant (.name, .added)", () => {
  const comments = extractComments(parse(whoreshubTheme));
  expect(comments).toEqual([{ text: "works", author: "Analgod69", date: expect.any(String), likes: 2 }]);
});

test("a page without comment items gives none", () => {
  expect(extractComments(parse("<div class='list-comments hidden'></div>"))).toEqual([]);
});
