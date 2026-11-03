import { expect, test } from 'bun:test';
import { justifiedRows } from '../src/lib/justify';

test('full rows fill the width exactly and keep the order', () => {
	const ratios = [1, 1, 1, 1, 2.1];
	const rows = justifiedRows(ratios, 410, 200, 10);
	expect(rows.map((row) => [row.start, row.end, row.full])).toEqual([
		[0, 2, true],
		[2, 4, true],
		[4, 5, true]
	]);
	for (const row of rows) {
		const widths = ratios.slice(row.start, row.end).map((ratio) => ratio * row.height);
		const used = widths.reduce((sum, width) => sum + width, 0) + 10 * (row.end - row.start - 1);
		expect(used).toBeCloseTo(410);
	}
});

test('a last row that does not reach the width keeps the target height', () => {
	const rows = justifiedRows([1, 1, 0.5], 410, 200, 10);
	expect(rows.at(-1)).toEqual({ start: 2, end: 3, height: 200, full: false });
});

test('no width lays out nothing', () => {
	expect(justifiedRows([1, 1], 0, 200, 10)).toEqual([]);
});
