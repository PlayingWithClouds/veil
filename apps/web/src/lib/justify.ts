// Justified photo rows: photos keep their proportions and their order, each
// row as tall as it needs to be to fill the width exactly.

/** One row of photos: those from start up to (not including) end, and its height. */
export interface PhotoRow {
	start: number;
	end: number;
	height: number;
	// Whether the row fills the width; the last one may not, and keeps the target height.
	full: boolean;
}

/**
 * Lays photos with the given aspect ratios (width over height) out in rows
 * of width pixels with gap between photos: each row takes photos until they
 * would reach the width at targetHeight, then shrinks to fill it exactly.
 */
export function justifiedRows(
	ratios: number[],
	width: number,
	targetHeight: number,
	gap: number
): PhotoRow[] {
	if (width <= 0) return [];
	const rows: PhotoRow[] = [];
	let start = 0;
	let ratioSum = 0;
	for (let index = 0; index < ratios.length; index++) {
		ratioSum += ratios[index];
		const gaps = gap * (index - start);
		if (ratioSum * targetHeight + gaps < width) continue;
		rows.push({ start, end: index + 1, height: (width - gaps) / ratioSum, full: true });
		start = index + 1;
		ratioSum = 0;
	}
	if (start < ratios.length) {
		rows.push({ start, end: ratios.length, height: targetHeight, full: false });
	}
	return rows;
}
