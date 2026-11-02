import { describe, expect, test } from 'bun:test';
import { SearchLayout, TOP_PLACES } from '../src/lib/searchLayout';

/** Results a, b, … for a site, ids prefixed with the site. */
function results(site: string, count: number) {
	return Array.from({ length: count }, (_, index) => ({ id: `${site}${index}` }));
}

describe('SearchLayout', () => {
	test('gives each site four top places and holds the rest', () => {
		const layout = new SearchLayout();
		expect(layout.add('a', results('a', 6)).map((item) => item.id)).toEqual(['a0', 'a1', 'a2', 'a3']);
		expect(layout.add('b', results('b', 2)).map((item) => item.id)).toEqual(['b0', 'b1']);
		expect(layout.release().map((item) => item.id)).toEqual(['a4', 'a5']);
	});

	test('releases the rest, sites taking turns, once the top is full', () => {
		const layout = new SearchLayout();
		const shown = ['a', 'b', 'c', 'd', 'e'].flatMap((site) => layout.add(site, results(site, 6)));
		expect(shown.slice(0, TOP_PLACES).every((item) => Number(item.id.slice(1)) < 4)).toBe(true);
		expect(shown.slice(TOP_PLACES).map((item) => item.id)).toEqual(['a4', 'b4', 'c4', 'd4', 'e4', 'a5', 'b5', 'c5', 'd5', 'e5']);
	});

	test('never repeats a result, nor one claimed elsewhere', () => {
		const layout = new SearchLayout();
		layout.claim('x');
		expect(layout.add('a', [{ id: 'x' }, { id: 'y' }]).map((item) => item.id)).toEqual(['y']);
		expect(layout.add('b', [{ id: 'y' }])).toEqual([]);
	});

	test('after release, results wait for the next release', () => {
		const layout = new SearchLayout();
		layout.release();
		expect(layout.add('a', results('a', 2))).toEqual([]);
		expect(layout.hasWaiting).toBe(true);
		expect(layout.release().map((item) => item.id)).toEqual(['a0', 'a1']);
	});
});
