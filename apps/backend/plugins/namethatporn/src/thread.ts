import { parse, type HTMLElement } from "node-html-parser";
import { relativeDate } from "@playingwithclouds/veil-sdk";
import type { ForumComment, ForumPost, ForumPostRef } from "@playingwithclouds/veil-sdk";
import { fetchPage } from "./http.ts";
import { SITE_ORIGIN } from "./parse.ts";

/** Fetches and parses one thread; only the site's post pages are accepted. */
export async function findPost(url: string): Promise<ForumPost> {
  if (!url.startsWith(`${SITE_ORIGIN}/post/`)) throw new Error(`not a NameThatPorn post: ${url}`);
  return parseThread(await fetchPage(url), url);
}

/** The site wraps every outgoing link in this redirect. */
const REDIRECT_PREFIX = "https://redir.me?";

/** Web addresses written into a comment's text without a link. */
const BARE_URL = /\b(?:https?:\/\/|www\.)[^\s<>"']+/gi;

/** The date of an element: its datetime attribute, else its relative text ("edited 9 years ago"). */
function dateOf(element: HTMLElement | null): string | undefined {
  if (!element) return undefined;
  const stamped = element.getAttribute("datetime");
  if (stamped) return new Date(stamped).toISOString();
  return relativeDate(element.text);
}

/** The real address behind a link of the site: unwrapped from its redirect; site-internal links are dropped. */
function outgoingLink(href: string | undefined): string | undefined {
  if (!href) return undefined;
  if (href.startsWith(REDIRECT_PREFIX)) return href.slice(REDIRECT_PREFIX.length);
  if (href.startsWith("https://") || href.startsWith("http://")) return href;
  return undefined;
}

/** A bare address as a link: "www.x.com/a." → "https://www.x.com/a". */
function normalizedBareUrl(found: string): string {
  const trimmed = found.replace(/[.,;:!?)\]]+$/, "");
  if (trimmed.toLowerCase().startsWith("www.")) return `https://${trimmed}`;
  return trimmed;
}

/** Appends value unless it is already there. */
function pushUnique(values: string[], value: string) {
  if (!values.includes(value)) values.push(value);
}

/** The cards a comment shows for other threads of the site, whose text isn't the comment's. */
const POST_CARD = /<a\b[^>]*\bntpref_wrapper\b[^>]*>[\s\S]*?<\/a>/gi;

/** The background picture of a card ("background-image: url('…')"). */
const BACKGROUND_URL = /url\(['"]?([^'")]+)['"]?\)/;

/**
 * The text of a comment body: line breaks kept, runs of blank lines folded,
 * cards of other threads left out (see postRefs).
 */
function plainText(body: HTMLElement): string {
  const withBreaks = body.innerHTML.replace(POST_CARD, "\n").replace(/<br\s*\/?>/gi, "\n");
  return parse(withBreaks)
    .text.split("\n")
    .map((line) => line.trim())
    .join("\n")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}

/** The other threads of the site a comment body shows as cards. */
function postRefs(body: HTMLElement): ForumPostRef[] {
  const refs: ForumPostRef[] = [];
  for (const card of body.querySelectorAll("a.ntpref_wrapper")) {
    const href = card.getAttribute("href");
    if (!href || !href.startsWith("/post/")) continue;
    const url = SITE_ORIGIN + href;
    if (refs.some((known) => known.url === url)) continue;
    const ref: ForumPostRef = {
      title: card.querySelector(".ntpref_title")?.text.trim() || card.getAttribute("title") || "Post",
      url,
    };
    const thumbnail = card.querySelector(".ntpref_image")?.getAttribute("style")?.match(BACKGROUND_URL);
    if (thumbnail) ref.thumbnail_url = thumbnail[1];
    refs.push(ref);
  }
  return refs;
}

/** Pictures and links of a comment body: picture links go to images, the rest to links, then bare addresses. */
function linksAndImages(body: HTMLElement, text: string): { links: string[]; images: string[] } {
  const links: string[] = [];
  const images: string[] = [];
  for (const anchor of body.querySelectorAll("a")) {
    if (anchor.classList.contains("item_img_container")) {
      const picture = anchor.querySelector("img")?.getAttribute("data-src") || outgoingLink(anchor.getAttribute("href"));
      if (picture) pushUnique(images, picture);
      continue;
    }
    const link = outgoingLink(anchor.getAttribute("href"));
    if (link) pushUnique(links, link);
  }
  for (const found of text.match(BARE_URL) || []) {
    const link = normalizedBareUrl(found);
    if (!images.includes(link)) pushUnique(links, link);
  }
  return { links, images };
}

/** The texts of the tags of one kind (".name_syntax", ".movie_syntax") in a comment body, without repeats. */
function tagTexts(body: HTMLElement, selector: string): string[] {
  const texts: string[] = [];
  for (const tag of body.querySelectorAll(`${selector} .name`)) {
    const text = tag.text.trim();
    if (text !== "") pushUnique(texts, text);
  }
  return texts;
}

/** Reads one comment block of a thread. */
function parseComment(wrapper: HTMLElement, acceptedId: string | undefined): ForumComment | undefined {
  const id = wrapper.getAttribute("data-commentid");
  const body = wrapper.querySelector(".id_answer");
  if (!id || !body) return undefined;
  const text = plainText(body);
  const comment: ForumComment = {
    id,
    text,
    accepted: id === acceptedId,
    names: tagTexts(body, ".name_syntax"),
    titles: tagTexts(body, ".movie_syntax"),
    posts: postRefs(body),
    ...linksAndImages(body, text),
  };
  const author = wrapper.querySelector(".id_answer_bottom_line .ud_unme")?.text.trim();
  if (author) comment.author = author;
  const date = dateOf(wrapper.querySelector(".idatl_date"));
  if (date) comment.date = date;
  const likes = Number(wrapper.querySelector(".idabres_upvote span")?.text);
  if (Number.isFinite(likes)) comment.likes = likes;
  return comment;
}

/**
 * Parses a thread page: the question (title, description, picture or clip,
 * tags) and its comments in the site's order. The accepted answer is shown
 * twice on the page, on top and in the thread; only the thread's copy is kept,
 * marked accepted.
 */
export function parseThread(html: string, url: string): ForumPost {
  const page = parse(html);
  const acceptedId = page.querySelector("#correct_answer")?.getAttribute("data-commentid");
  const post: ForumPost = {
    url,
    title: page.querySelector("#id_title_top")?.text.trim() || "Untitled post",
    tags: page.querySelectorAll("#id_tags a").map((tag) => tag.text.trim()).filter((tag) => tag !== ""),
    solved: acceptedId !== undefined,
    comments: [],
  };
  const description = page.querySelector("#id_description")?.text.trim();
  if (description) post.description = description;
  const imageUrl = page.querySelector("#id_media_wrapper img.item_img")?.getAttribute("src");
  if (imageUrl) post.image_url = imageUrl;
  const videoUrl = page.querySelector('#id_media_wrapper video source[type="video/mp4"]')?.getAttribute("src");
  if (videoUrl) {
    post.video_url = videoUrl;
    // The site keeps a still of every clip beside it, under the same name.
    if (!post.image_url) post.image_url = videoUrl.replace(/\.mp4$/, ".jpg");
  }
  const author = page.querySelector("#id_resources_wrapper .ud_unme")?.text.trim();
  if (author) post.author = author;
  const date = dateOf(page.querySelector(".item_date time"));
  if (date) post.date = date;

  for (const wrapper of page.querySelectorAll(".id_answer_wrapper")) {
    if (!wrapper.getAttribute("id")?.startsWith("answer_")) continue;
    const comment = parseComment(wrapper, acceptedId);
    if (comment) post.comments.push(comment);
  }
  return post;
}
