import { test, expect } from "bun:test";
import { extractComments } from "../src/comments.ts";
import { parseInitials } from "../src/initials.ts";

const initials = {
  commentsComponent: {
    commentsList: {
      items: [
        { id: "a", text: "Incredible.&nbsp; How &amp; why", replyToId: null, created: 1781326733, likes: null, status: "public", author: { name: "John" } },
        { id: "b", text: "agreed", replyToId: "a", created: 1781326800, likes: 3, status: "public", author: { name: "Mia" } },
        { id: "c", text: "buy followers", replyToId: null, created: 1781326900, isSpam: true, author: { name: "Bot" } },
        { id: "d", text: "  ", replyToId: null, created: 1781327000, author: { name: "Blank" } },
        { id: "e", text: "late parent", replyToId: "f", created: 1781327100, author: { name: "Eve" } },
        { id: "f", text: "listed after its reply", replyToId: null, created: 1781327200, author: { name: "Finn" } },
      ],
    },
  },
};

test("reads text, author, date and likes, skipping spam and empty comments", () => {
  const comments = extractComments(initials);
  expect(comments.map((comment) => comment.text)).toEqual([
    "Incredible. How & why",
    "agreed",
    "late parent",
    "listed after its reply",
  ]);
  expect(comments[0].author).toBe("John");
  expect(comments[0].date).toBe("2026-06-13T04:58:53.000Z");
  expect(comments[0].likes).toBeUndefined();
  expect(comments[1].likes).toBe(3);
});

test("replyToId becomes the index of the kept parent", () => {
  const comments = extractComments(initials);
  expect(comments.map((comment) => comment.reply_to)).toEqual([undefined, 0, 3, undefined]);
});

test("pages without comments give an empty list", () => {
  expect(extractComments(undefined)).toEqual([]);
  expect(extractComments(parseInitials('window.initials={"videoModel":{}};</script>'))).toEqual([]);
});
