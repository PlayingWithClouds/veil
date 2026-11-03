import { parse } from "node-html-parser";
import type { HTMLElement } from "node-html-parser";
import type { ScenePerformer, ScrapeResult, Story, StoryChapter } from "@playingwithclouds/veil-sdk";
import { postedDate } from "./dates.ts";
import {
  authorOfProfile,
  externalIdOf,
  fetchPage,
  normalizeText,
  storyIdOf,
  storySlugOf,
  storyUrl,
} from "./site.ts";
import { countWords, plainText } from "./text.ts";

// The vote list and rating sit inside the text's panel, behind this block.
const TEXT_END = /(?:<hr[^>]*>\s*)?<div class="story_info"/i;

/** The panel holding the story text: the one that is neither the introduction nor the comments. */
function textPanel(root: HTMLElement): HTMLElement | undefined {
  for (const panel of root.querySelectorAll(".block_panel")) {
    if (panel.querySelector(".count_comments")) continue;
    if (isIntroduction(panel)) continue;
    return panel;
  }
  return undefined;
}

/** Whether a panel is the author's introduction ("Introduction:" heading). */
function isIntroduction(panel: HTMLElement): boolean {
  const heading = panel.querySelector("h2");
  if (!heading) return false;
  return normalizeText(heading.text).startsWith("Introduction");
}

/** The story text of its panel, cut off before the votes block. */
function storyText(panel: HTMLElement): string {
  const markup = panel.innerHTML;
  const end = markup.search(TEXT_END);
  if (end < 0) return plainText(markup);
  return plainText(markup.slice(0, end));
}

/** The author's introduction, or an empty string when there is none. */
function introductionOf(root: HTMLElement): string {
  for (const panel of root.querySelectorAll(".block_panel")) {
    if (!isIntroduction(panel)) continue;
    return plainText(panel.innerHTML.replace(/<h2[^>]*>[\s\S]*?<\/h2>/i, ""));
  }
  return "";
}

/** The author credited in the byline, or undefined for anonymous stories. */
function authorOf(root: HTMLElement): ScenePerformer | undefined {
  const link = root.querySelector("#top_panel .title_link a");
  if (!link) return undefined;
  return authorOfProfile(link.text, link.getAttribute("href") ?? "");
}

/** The genre and themes under the title, once each. */
function tagsOf(root: HTMLElement): string[] {
  const tags: string[] = [];
  const line = root.querySelector("#top_panel .top_info");
  for (const part of (line?.text ?? "").split(",")) {
    const tag = normalizeText(part);
    if (tag === "" || tags.includes(tag)) continue;
    tags.push(tag);
  }
  return tags;
}

/** The title in the heading, without the "by <author>" byline that shares it. */
function titleOf(root: HTMLElement): string {
  const heading = root.querySelector("#top_panel h2");
  if (!heading) return "";
  const byline = heading.querySelector(".title_link")?.text ?? "";
  return normalizeText(heading.text.replace(byline, ""));
}

/** Builds the Story of one story page, or undefined when the page holds none (unknown ids answer 200 with the shell). */
export function storyFromPage(html: string, storyId: string, sourceUrl: string): Story | undefined {
  const root = parse(html);
  const panel = textPanel(root);
  const title = titleOf(root);
  if (!panel || title === "") return undefined;

  const text = storyText(panel);
  const wordCount = countWords(text);
  const chapter: StoryChapter = { title, url: sourceUrl, text, word_count: wordCount };
  const story: Story = {
    type: "story",
    external_id: externalIdOf(storyId),
    source_url: sourceUrl,
    title,
    complete: true,
    chapters: [chapter],
  };
  if (wordCount > 0) story.word_count = wordCount;

  const details = introductionOf(root);
  if (details !== "") story.details = details;
  const date = postedDate(normalizeText(root.querySelector(".story_date")?.text ?? ""));
  if (date !== undefined) story.date = date;
  const author = authorOf(root);
  if (author) story.performers = [author];
  const tags = tagsOf(root);
  if (tags.length > 0) story.tags = tags;
  return story;
}

/** Fetches a story page and builds its Story; throws for an address that is no story or no longer exists. */
async function loadStory(url: string): Promise<Story> {
  const storyId = storyIdOf(url);
  if (storyId === undefined) throw new Error(`Not a sexstories.com story address: ${url}`);
  const sourceUrl = storyUrl(storyId, storySlugOf(url));
  const story = storyFromPage(await fetchPage(sourceUrl), storyId, sourceUrl);
  if (!story) throw new Error(`No story at ${sourceUrl}`);
  return story;
}

/** Serves "story:find": a story with its text as one chapter (the site never splits a story across pages). */
export async function findStory(url: string): Promise<ScrapeResult[]> {
  if (storyIdOf(url) === undefined) return [];
  return [{ type: "story", story: await loadStory(url) }];
}

/** Serves "story:chapter": the text of a story's only chapter. */
export async function readChapter(url: string): Promise<StoryChapter> {
  const story = await loadStory(url);
  return story.chapters?.[0] ?? { url, text: "" };
}
