// Finding the way through a story's chapters, and the reading position as
// UserState.progress keeps it: the chapter's position plus how far through
// it reading got (3.4 is 40% into the chapter at position 3).

/** A chapter as navigation knows it. */
export interface ChapterLink {
	id: string;
	position: number;
	title?: string | null;
	parent?: string | null;
	choice?: string | null;
}

/** Where reading stopped: a chapter by position, and how far into it (0..1). */
export interface ReadingPoint {
	position: number;
	fraction: number;
}

/** A chapter in the contents, with how deep its branch goes. */
export interface ContentsEntry<Chapter extends ChapterLink> {
	chapter: Chapter;
	depth: number;
}

// Kept below 1, so a chapter read to the end doesn't count as the next one.
const MAX_FRACTION = 0.999;
// Words read per minute, for the time left.
const WORDS_PER_MINUTE = 230;

/** Whether the story branches: a chapter is reached by a choice. */
export function isBranching(chapters: ChapterLink[]): boolean {
	return chapters.some((chapter) => Boolean(chapter.parent));
}

/** The chapters a chapter's choices lead to, in story order. */
export function choicesOf<Chapter extends ChapterLink>(
	chapters: Chapter[],
	chapterId: string
): Chapter[] {
	return chapters.filter((chapter) => chapter.parent === chapterId);
}

/**
 * What follows a chapter: in a branching story the chapters its choices lead
 * to, else the next chapter. Empty at an ending.
 */
export function followingChapters<Chapter extends ChapterLink>(
	chapters: Chapter[],
	current: Chapter
): Chapter[] {
	if (isBranching(chapters)) return choicesOf(chapters, current.id);
	const index = chapters.findIndex((chapter) => chapter.id === current.id);
	const next = chapters[index + 1];
	if (index === -1 || !next) return [];
	return [next];
}

/** What comes before a chapter: its parent in a branching story, else the chapter before. */
export function previousChapter<Chapter extends ChapterLink>(
	chapters: Chapter[],
	current: Chapter
): Chapter | null {
	if (current.parent) return chapters.find((chapter) => chapter.id === current.parent) ?? null;
	if (isBranching(chapters)) return null;
	const index = chapters.findIndex((chapter) => chapter.id === current.id);
	if (index <= 0) return null;
	return chapters[index - 1];
}

/** The choices that led from the first chapter to this one, first choice first. */
export function pathTo<Chapter extends ChapterLink>(
	chapters: Chapter[],
	current: Chapter
): Chapter[] {
	const byId = new Map(chapters.map((chapter) => [chapter.id, chapter]));
	const path: Chapter[] = [];
	const visited = new Set<string>();
	let step: Chapter | undefined = current;
	while (step?.parent && !visited.has(step.id)) {
		visited.add(step.id);
		path.unshift(step);
		step = byId.get(step.parent);
	}
	return path;
}

/**
 * The chapters in reading order with their depth: a linear story flat, a
 * branching one depth first from its roots (chapters without a parent), each
 * choice's branch under it.
 */
export function contentsOf<Chapter extends ChapterLink>(
	chapters: Chapter[]
): ContentsEntry<Chapter>[] {
	if (!isBranching(chapters)) return chapters.map((chapter) => ({ chapter, depth: 0 }));
	const known = new Set(chapters.map((chapter) => chapter.id));
	const roots = chapters.filter((chapter) => !chapter.parent || !known.has(chapter.parent));
	const entries: ContentsEntry<Chapter>[] = [];
	const visited = new Set<string>();
	// An explicit stack: real branching stories run hundreds of choices deep.
	const stack = roots.reverse().map((chapter) => ({ chapter, depth: 0 }));
	while (stack.length > 0) {
		const entry = stack.pop();
		if (!entry || visited.has(entry.chapter.id)) continue;
		visited.add(entry.chapter.id);
		entries.push(entry);
		const children = choicesOf(chapters, entry.chapter.id).reverse();
		for (const child of children) stack.push({ chapter: child, depth: entry.depth + 1 });
	}
	return entries;
}

/** A reading point as progress. */
export function progressOf(point: ReadingPoint): number {
	return point.position + Math.min(Math.max(point.fraction, 0), MAX_FRACTION);
}

/** Progress as a reading point; null when nothing was read. */
export function readingPointOf(progress: number | null | undefined): ReadingPoint | null {
	if (progress === null || progress === undefined || progress < 0) return null;
	const position = Math.floor(progress);
	return { position, fraction: progress - position };
}

/** The minutes it takes to read what's left of words, at least 1. */
export function minutesLeft(words: number, fraction: number): number {
	const left = words * (1 - Math.min(Math.max(fraction, 0), 1));
	return Math.max(1, Math.round(left / WORDS_PER_MINUTE));
}

/** A chapter's text as paragraphs: blank lines part them, single line breaks stay. */
export function paragraphsOf(text: string): string[] {
	return text
		.split(/\n\s*\n/)
		.map((paragraph) => paragraph.trim())
		.filter((paragraph) => paragraph !== '');
}
