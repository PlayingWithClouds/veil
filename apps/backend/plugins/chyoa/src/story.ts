import { parse } from "node-html-parser";
import type { HTMLElement } from "node-html-parser";
import type { ScenePerformer, ScrapeResult, Story, StoryChapter } from "@playingwithclouds/veil-sdk";
import {
  SITE_URL,
  absoluteUrl,
  chapterIdOf,
  externalIdOf,
  fetchReaderPage,
  normalizeText,
  storyIdOf,
  storyUrl,
} from "./site.ts";
import { countWords, decodeEntities, plainText } from "./text.ts";

const MONTHS = ["jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"];

/** "Nov 2, 2026" as "2026-11-02"; undefined for any other text. */
export function writtenDate(text: string): string | undefined {
  const match = text.match(/([A-Za-z]{3})[a-z]*\.?\s+(\d{1,2}),\s+(\d{4})/);
  if (!match) return undefined;
  const month = MONTHS.indexOf(match[1].toLowerCase());
  if (month < 0) return undefined;
  const monthNumber = String(month + 1).padStart(2, "0");
  return `${match[3]}-${monthNumber}-${match[2].padStart(2, "0")}`;
}

/** The options a chapter offers, each as the chapter it leads to (titled like the option). */
export function choicesOf(root: HTMLElement): StoryChapter[] {
  const choices: StoryChapter[] = [];
  for (const link of root.querySelectorAll("a.rd-story-branch-link")) {
    const url = absoluteUrl(link.getAttribute("href") ?? "");
    const text = link.querySelector(".rd-story-branch-link__text");
    if (chapterIdOf(url) === undefined || !text) continue;
    const choice = normalizeText(text.text);
    choices.push({ title: choice, url, choice });
  }
  return choices;
}

/** The heading of the chapter on the page: the story's title on its first chapter. */
function headingOf(root: HTMLElement): string {
  const heading = root.querySelector(".rd-story-chapter h1");
  if (!heading) return "";
  return normalizeText(heading.text);
}

/** The chapter's text, or an empty string when the page shows none. */
function proseOf(root: HTMLElement): string {
  const prose = root.querySelector(".chapter-content");
  if (!prose) return "";
  return plainText(prose.innerHTML);
}

/** A chapter page as a chapter with its text, or undefined when the page holds none. */
export function chapterFromPage(html: string, url: string): StoryChapter | undefined {
  const root = parse(html);
  const title = headingOf(root);
  if (title === "") return undefined;
  const text = proseOf(root);
  const chapter: StoryChapter = { title, url, text, word_count: countWords(text) };
  const choices = choicesOf(root);
  if (choices.length > 0) chapter.choices = choices;
  return chapter;
}

/** The address of the page's chapter without the "/bookmark" action, or undefined. */
function chapterUrlOf(root: HTMLElement): string | undefined {
  const href = root.querySelector("a.bookmark-link")?.getAttribute("href");
  if (href === undefined) return undefined;
  return href.replace(/\/bookmark$/, "");
}

/** The address of the story, from the page's "Start Over" link. */
function storyUrlOf(root: HTMLElement): string | undefined {
  for (const link of root.querySelectorAll(".rd-story-navigation__links a")) {
    const href = link.getAttribute("href") ?? "";
    if (/\/story\/[^/?#]+$/.test(href)) return href;
  }
  return undefined;
}

/** The value of a row ("Published", "Category") of the story's details list. */
function detailOf(root: HTMLElement, name: string): HTMLElement | undefined {
  for (const row of root.querySelectorAll("dl div")) {
    const label = row.querySelector("dt");
    if (!label || normalizeText(label.text) !== name) continue;
    const value = row.querySelector("dd");
    if (value) return value;
  }
  return undefined;
}

/** The author credited in the first chapter's byline, or undefined. */
function authorOf(root: HTMLElement): ScenePerformer | undefined {
  const link = root.querySelector('.rd-story-reading-byline a[href*="/user/"]');
  if (!link) return undefined;
  const name = normalizeText(link.text);
  if (name === "") return undefined;
  return {
    name,
    external_id: `chyoa-user-${name.toLowerCase()}`,
    source_url: absoluteUrl(link.getAttribute("href") ?? "").replace(/(?:%20)+$/, ""),
  };
}

/** The category, then the tags the author gave, once each. */
function tagsOf(root: HTMLElement): string[] {
  const names = [detailOf(root, "Category")?.text ?? ""];
  for (const link of root.querySelectorAll('[id^="story-tags-"] a')) {
    names.push(link.text);
  }
  const tags: string[] = [];
  for (const name of names) {
    const tag = normalizeText(name);
    if (tag !== "" && !tags.includes(tag)) tags.push(tag);
  }
  return tags;
}

/** The story's blurb: the "About" text, else the page description. */
function descriptionOf(root: HTMLElement): string {
  const about = root.querySelector("#about p");
  if (about) return normalizeText(about.text);
  const meta = root.querySelector('meta[property="og:description"]')?.getAttribute("content");
  if (meta === undefined) return "";
  return normalizeText(decodeEntities(meta));
}

/** The first chapter of a story page, with the options it offers as its children. */
function firstChapter(root: HTMLElement, title: string): StoryChapter[] {
  const chapterId = root.querySelector("#content")?.getAttribute("data-chapter-id") ?? "";
  const url = chapterUrlOf(root) ?? `${SITE_URL}/chapter/Introduction.${chapterId}`;
  const text = proseOf(root);
  const first: StoryChapter = { title, url, text, word_count: countWords(text) };
  const chapters = [first];
  for (const choice of choicesOf(root)) {
    chapters.push({ ...choice, parent_url: url });
  }
  return chapters;
}

/** The story's address as the page names itself (the site's slug), else the address it was fetched by. */
function canonicalUrlOf(root: HTMLElement, fetchedUrl: string): string {
  const canonical = root.querySelector('meta[property="og:url"]')?.getAttribute("content");
  if (canonical !== undefined && storyIdOf(canonical) !== undefined) return canonical;
  return fetchedUrl;
}

/** Builds the Story of one story page (its first chapter), or undefined when the page holds none. */
export function storyFromPage(html: string, fetchedUrl: string): Story | undefined {
  const root = parse(html);
  const storyId = root.querySelector("#content")?.getAttribute("data-story-id");
  const title = headingOf(root);
  if (storyId === undefined || title === "") return undefined;

  const story: Story = {
    type: "story",
    external_id: externalIdOf(storyId),
    source_url: canonicalUrlOf(root, fetchedUrl),
    title,
    chapters: firstChapter(root, title),
  };
  const details = descriptionOf(root);
  if (details !== "") story.details = details;
  const published = detailOf(root, "Published")?.text;
  const date = writtenDate(published ?? "");
  if (date !== undefined) story.date = date;
  const author = authorOf(root);
  if (author) story.performers = [author];
  const tags = tagsOf(root);
  if (tags.length > 0) story.tags = tags;
  const cover = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  if (cover !== undefined && cover.includes("/data/covers/")) story.cover_path = cover;
  return story;
}

/** The story page an address names: a story address itself, a chapter address through its "Start Over" link. */
async function storyPageOf(url: string): Promise<{ html: string; sourceUrl: string }> {
  const storyId = storyIdOf(url);
  if (storyId !== undefined) return { html: await fetchReaderPage(storyUrl(storyId)), sourceUrl: url };
  if (chapterIdOf(url) === undefined) throw new Error(`Not a chyoa.com story address: ${url}`);
  const startOver = storyUrlOf(parse(await fetchReaderPage(url)));
  if (startOver === undefined) throw new Error(`No story behind ${url}`);
  return { html: await fetchReaderPage(startOver), sourceUrl: startOver };
}

/** Serves "story:find": the story, its first chapter and the chapters that chapter's options lead to. */
export async function findStory(url: string): Promise<ScrapeResult[]> {
  if (storyIdOf(url) === undefined && chapterIdOf(url) === undefined) return [];
  const page = await storyPageOf(url);
  const story = storyFromPage(page.html, page.sourceUrl);
  if (!story) throw new Error(`No story at ${url}`);
  return [{ type: "story", story }];
}

/** Serves "story:chapter": a chapter's text and the chapters its options lead to. */
export async function readChapter(url: string): Promise<StoryChapter> {
  if (storyIdOf(url) === undefined && chapterIdOf(url) === undefined) {
    throw new Error(`Not a chyoa.com chapter address: ${url}`);
  }
  const chapter = chapterFromPage(await fetchReaderPage(url), url);
  if (!chapter) throw new Error(`No chapter at ${url}`);
  return chapter;
}
