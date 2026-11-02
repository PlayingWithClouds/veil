/**
 * Perceptual hash of a video: its sampled frames are tiled into one sheet
 * (the trickplay sheet) and the sheet is hashed like pHash, 64 bits from the
 * lowest frequencies of a 32×32 DCT. Re-encodes, resizes and watermarks move
 * few bits, so the backend counts hashes a few bits apart as the same media.
 */

/** Side of the grayscale square the DCT runs on. */
const HASH_INPUT = 32;

/** Side of the low-frequency block the bits come from. */
const HASH_BLOCK = 8;

/** Luminance (Rec. 601) of every pixel of an RGBA image. */
export function grayscale(rgba: ArrayLike<number>, width: number, height: number): Float64Array {
	const gray = new Float64Array(width * height);
	for (let pixel = 0; pixel < gray.length; pixel++) {
		gray[pixel] =
			0.299 * rgba[pixel * 4] + 0.587 * rgba[pixel * 4 + 1] + 0.114 * rgba[pixel * 4 + 2];
	}
	return gray;
}

/** Shrinks a grayscale image to size×size by averaging the pixels each target pixel covers. */
export function shrink(
	gray: Float64Array,
	width: number,
	height: number,
	size: number
): Float64Array {
	const shrunk = new Float64Array(size * size);
	for (let row = 0; row < size; row++) {
		const top = Math.floor((row * height) / size);
		const bottom = Math.max(top + 1, Math.floor(((row + 1) * height) / size));
		for (let column = 0; column < size; column++) {
			const left = Math.floor((column * width) / size);
			const right = Math.max(left + 1, Math.floor(((column + 1) * width) / size));
			shrunk[row * size + column] = average(gray, width, left, top, right, bottom);
		}
	}
	return shrunk;
}

/** Mean of a rectangle of a grayscale image. */
function average(
	gray: Float64Array,
	width: number,
	left: number,
	top: number,
	right: number,
	bottom: number
): number {
	let sum = 0;
	for (let y = top; y < bottom; y++) {
		for (let x = left; x < right; x++) sum += gray[y * width + x];
	}
	return sum / ((right - left) * (bottom - top));
}

/** The 2D DCT-II of a size×size block. */
export function dct2d(block: Float64Array, size: number): Float64Array {
	const rows = new Float64Array(size * size);
	for (let row = 0; row < size; row++) {
		dct1d(block, row * size, 1, rows, row * size, 1, size);
	}
	const result = new Float64Array(size * size);
	for (let column = 0; column < size; column++) {
		dct1d(rows, column, size, result, column, size, size);
	}
	return result;
}

/** One-dimensional DCT-II over a strided run of values. */
function dct1d(
	input: Float64Array,
	inputStart: number,
	inputStep: number,
	output: Float64Array,
	outputStart: number,
	outputStep: number,
	size: number
) {
	for (let frequency = 0; frequency < size; frequency++) {
		let sum = 0;
		for (let index = 0; index < size; index++) {
			sum +=
				input[inputStart + index * inputStep] *
				Math.cos((Math.PI * (2 * index + 1) * frequency) / (2 * size));
		}
		output[outputStart + frequency * outputStep] = sum;
	}
}

/**
 * The 64-bit perceptual hash of a grayscale image as 16 hex digits: a bit
 * per low-frequency DCT coefficient, set when it's above their median.
 */
export function perceptualHash(gray: Float64Array, width: number, height: number): string {
	const coefficients = dct2d(shrink(gray, width, height, HASH_INPUT), HASH_INPUT);
	const low: number[] = [];
	for (let row = 0; row < HASH_BLOCK; row++) {
		for (let column = 0; column < HASH_BLOCK; column++)
			low.push(coefficients[row * HASH_INPUT + column]);
	}
	const middle = median(low);
	let hash = 0n;
	for (const coefficient of low) {
		hash <<= 1n;
		if (coefficient > middle) hash |= 1n;
	}
	return hash.toString(16).padStart(16, '0');
}

/** The median of values. */
function median(values: number[]): number {
	const sorted = [...values].sort((first, second) => first - second);
	const half = Math.floor(sorted.length / 2);
	if (sorted.length % 2 === 1) return sorted[half];
	return (sorted[half - 1] + sorted[half]) / 2;
}

/** How many bits two 16-hex-digit hashes differ in. */
export function hashDistance(first: string, second: string): number {
	let differing = BigInt(`0x${first}`) ^ BigInt(`0x${second}`);
	let count = 0;
	while (differing > 0n) {
		count += Number(differing & 1n);
		differing >>= 1n;
	}
	return count;
}
