import { describe, expect, test } from 'bun:test';
import {
	contentsOf,
	followingChapters,
	isBranching,
	paragraphsOf,
	pathTo,
	previousChapter,
	progressOf,
	readingPointOf
} from '../src/lib/storyReading';

const linear = [
	{ id: 'a', position: 0 },
	{ id: 'b', position: 1 },
	{ id: 'c', position: 2 }
];

// a → (b → d, c); positions in the order the chapters were found.
const branching = [
	{ id: 'a', position: 0 },
	{ id: 'b', position: 1, parent: 'a', choice: 'Stay' },
	{ id: 'c', position: 2, parent: 'a', choice: 'Leave' },
	{ id: 'd', position: 3, parent: 'b', choice: 'Knock' }
];

describe('navigation', () => {
	test('a linear story goes chapter by chapter', () => {
		expect(isBranching(linear)).toBe(false);
		expect(followingChapters(linear, linear[0]).map((chapter) => chapter.id)).toEqual(['b']);
		expect(followingChapters(linear, linear[2])).toEqual([]);
		expect(previousChapter(linear, linear[1])?.id).toBe('a');
		expect(previousChapter(linear, linear[0])).toBeNull();
	});

	test('a branching story follows the choices and goes back to the parent', () => {
		expect(isBranching(branching)).toBe(true);
		expect(followingChapters(branching, branching[0]).map((chapter) => chapter.id)).toEqual([
			'b',
			'c'
		]);
		expect(followingChapters(branching, branching[2])).toEqual([]);
		expect(previousChapter(branching, branching[3])?.id).toBe('b');
		expect(previousChapter(branching, branching[0])).toBeNull();
		expect(pathTo(branching, branching[3]).map((chapter) => chapter.choice)).toEqual([
			'Stay',
			'Knock'
		]);
	});

	test('contents list each branch under its choice', () => {
		expect(contentsOf(branching).map((entry) => `${entry.chapter.id}${entry.depth}`)).toEqual([
			'a0',
			'b1',
			'd2',
			'c1'
		]);
		expect(contentsOf(linear).map((entry) => entry.depth)).toEqual([0, 0, 0]);
	});
});

describe('reading point', () => {
	test('round-trips through progress and never reaches the next chapter', () => {
		expect(readingPointOf(progressOf({ position: 3, fraction: 0.4 }))).toEqual({
			position: 3,
			fraction: expect.closeTo(0.4, 5)
		});
		expect(Math.floor(progressOf({ position: 3, fraction: 1 }))).toBe(3);
		expect(readingPointOf(null)).toBeNull();
	});
});

test('paragraphs part at blank lines and keep single line breaks', () => {
	expect(paragraphsOf('One\nline two\n\n\n  Two  \n \nThree')).toEqual([
		'One\nline two',
		'Two',
		'Three'
	]);
});
