import type { ScrapeResult, Story, StoryChapter } from "@playingwithclouds/veil-sdk";
import {
  authorUrl,
  fetchSeries,
  fetchSeriesWorks,
  fetchStoryPage,
  seriesOf,
  seriesUrl,
  targetOf,
  workUrl,
  type ApiSeries,
  type ApiStoryPage,
  type ApiWork,
} from "./api.ts";
import { categoryName } from "./categories.ts";
import { isoDateOf, joinPages, plainPageText } from "./text.ts";

/** The identity of a standalone work. */
export function workExternalId(workId: number): string {
  return `story-${workId}`;
}

/** The identity of a series, which is one story made of its works. */
export function seriesExternalId(seriesId: number): string {
  return `series-${seriesId}`;
}

/** Serves "story:find": a work or a series by URL, as one story with its chapters listed. */
export async function findStory(url: string): Promise<ScrapeResult[]> {
  const story = await storyAt(url);
  return [{ type: "story", story }];
}

/** The story a work or series URL stands for. */
async function storyAt(url: string): Promise<Story> {
  const target = targetOf(url);
  if (target.kind === "series") {
    return seriesStory(target.slugOrId, undefined);
  }
  if (target.kind !== "work") {
    throw new Error(`Not a Literotica story or series URL: ${url}`);
  }
  const firstPage = await fetchStoryPage(target.slug, 1);
  const series = seriesOf(firstPage.submission);
  if (series === undefined) {
    return standaloneStory(firstPage);
  }
  return seriesStory(series.meta.id, firstPage);
}

/** A work that belongs to no series: a story of one chapter. */
function standaloneStory(firstPage: ApiStoryPage): Story {
  const work = firstPage.submission;
  return {
    type: "story",
    external_id: workExternalId(work.id),
    source_url: workUrl(work.url),
    title: work.title,
    details: nonEmpty(work.description),
    date: isoDateOf(work.date_approve),
    performers: [performerOf(work)],
    tags: tagsOf([work]),
    word_count: work.words_count,
    complete: true,
    chapters: [chapterOf(work, loadedText(work, firstPage))],
  };
}

/** A series: one story whose chapters are its works. `openedPage` is the work page the caller already fetched. */
async function seriesStory(slugOrId: string | number, openedPage: ApiStoryPage | undefined): Promise<Story> {
  const series = await fetchSeries(slugOrId);
  const works = (await fetchSeriesWorks(series.id)).filter((work) => work.type === "story");
  if (works.length === 0) {
    throw new Error(`Series ${series.title} has no stories`);
  }
  return {
    type: "story",
    external_id: seriesExternalId(series.id),
    source_url: seriesUrl(series),
    title: series.title,
    details: seriesDetails(series, works),
    date: isoDateOf(works[0].date_approve),
    performers: [performerOf(works[0])],
    tags: tagsOf(works),
    word_count: totalWords(works),
    complete: series.state === "completed",
    chapters: works.map((work) => chapterOf(work, loadedText(work, openedPage))),
  };
}

/** The series blurb, else its introduction, else the first work's description. */
function seriesDetails(series: ApiSeries, works: ApiWork[]): string | undefined {
  return nonEmpty(series.description) ?? nonEmpty(series.introduction) ?? nonEmpty(works[0].description);
}

/** The text of a work when the page already fetched holds all of it, which is when the work has one page. */
function loadedText(work: ApiWork, openedPage: ApiStoryPage | undefined): string | undefined {
  if (openedPage === undefined || openedPage.submission.id !== work.id) {
    return undefined;
  }
  if (openedPage.meta.pages_count !== 1) {
    return undefined;
  }
  return plainPageText(openedPage.pageText);
}

/** One chapter entry for a work. */
function chapterOf(work: ApiWork, text: string | undefined): StoryChapter {
  const chapter: StoryChapter = { title: work.title, url: workUrl(work.url), word_count: work.words_count };
  if (text !== undefined) {
    chapter.text = text;
  }
  return chapter;
}

/** The author as the story's performer; listing cards credit the same identity as their studio. */
export function performerOf(work: ApiWork): { name: string; external_id: string; source_url: string } {
  return { name: work.authorname, external_id: work.authorname, source_url: authorUrl(work.authorname) };
}

/** The category name first, then the works' own tags, without repeats. Purely numeric tags arrive as numbers. */
function tagsOf(works: ApiWork[]): string[] {
  const ownTags = works.flatMap((work) => (work.tags ?? []).map((entry) => String(entry.tag).trim()));
  return withoutRepeats([categoryName(works[0].category_info.pageUrl), ...ownTags]).filter((tag) => tag !== "");
}

/** The names in order, dropping later ones that differ only in case. */
function withoutRepeats(names: string[]): string[] {
  const seen = new Set<string>();
  return names.filter((name) => {
    const key = name.toLowerCase();
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
}

/** The sum of the works' word counts. */
function totalWords(works: ApiWork[]): number {
  return works.reduce((sum, work) => sum + (work.words_count ?? 0), 0);
}

/** The text, or undefined when it is missing or blank. */
function nonEmpty(text: string | null | undefined): string | undefined {
  if (text === undefined || text === null || text.trim() === "") {
    return undefined;
  }
  return text.trim();
}

/** Serves "story:chapter": one work's whole text, every page joined. */
export async function readChapter(url: string): Promise<StoryChapter> {
  const target = targetOf(url);
  if (target.kind !== "work") {
    throw new Error(`Not a Literotica story URL: ${url}`);
  }
  const firstPage = await fetchStoryPage(target.slug, 1);
  const pages = [firstPage.pageText];
  for (let page = 2; page <= firstPage.meta.pages_count; page++) {
    pages.push((await fetchStoryPage(target.slug, page)).pageText);
  }
  return chapterOf(firstPage.submission, joinPages(pages));
}
