/**
 * Face alignment for SFace (OpenCV Zoo, face_recognition_sface_2021dec):
 * the five landmarks are mapped onto the ArcFace 112×112 template by a
 * similarity transform, and the crop becomes the recognizer's input. Pure,
 * so it runs in a worker and under `bun test`.
 */
import type { Point } from './detect';

/** Side of the aligned crop SFace takes. */
export const RECOGNIZER_SIZE = 112;

/**
 * Where the five landmarks sit in an aligned 112×112 face (the ArcFace
 * template OpenCV's FaceRecognizerSF aligns to), in YuNet's landmark order.
 */
export const FACE_TEMPLATE: Point[] = [
	[38.2946, 51.6963],
	[73.5318, 51.5014],
	[56.0252, 71.7366],
	[41.5493, 92.3655],
	[70.7299, 92.2041]
];

/**
 * A similarity transform (rotation, uniform scale, translation):
 * x' = a·x − b·y + tx, y' = b·x + a·y + ty.
 */
export interface Similarity {
	a: number;
	b: number;
	tx: number;
	ty: number;
}

/** The least-squares similarity transform taking `from` onto `to`. */
export function similarityTransform(from: Point[], to: Point[]): Similarity {
	const fromMean = mean(from);
	const toMean = mean(to);
	let dot = 0;
	let cross = 0;
	let norm = 0;
	for (let index = 0; index < from.length; index++) {
		const fromX = from[index][0] - fromMean[0];
		const fromY = from[index][1] - fromMean[1];
		const toX = to[index][0] - toMean[0];
		const toY = to[index][1] - toMean[1];
		dot += fromX * toX + fromY * toY;
		cross += fromX * toY - fromY * toX;
		norm += fromX * fromX + fromY * fromY;
	}
	const a = dot / norm;
	const b = cross / norm;
	return {
		a,
		b,
		tx: toMean[0] - (a * fromMean[0] - b * fromMean[1]),
		ty: toMean[1] - (b * fromMean[0] + a * fromMean[1])
	};
}

/** Applies a similarity transform to a point. */
export function applySimilarity(transform: Similarity, [x, y]: Point): Point {
	return [
		transform.a * x - transform.b * y + transform.tx,
		transform.b * x + transform.a * y + transform.ty
	];
}

/** The centroid of points. */
function mean(points: Point[]): Point {
	let x = 0;
	let y = 0;
	for (const point of points) {
		x += point[0];
		y += point[1];
	}
	return [x / points.length, y / points.length];
}

/**
 * The recognizer's input for one face: the frame (RGBA, width×height) warped
 * so the landmarks land on the template, as planar RGB 0..255 (OpenCV's
 * blobFromImage with swapRB on its BGR crop).
 */
export function alignedInput(
	rgba: ArrayLike<number>,
	width: number,
	height: number,
	landmarks: Point[]
): Float32Array {
	const forward = similarityTransform(landmarks, FACE_TEMPLATE);
	const inverse = invert(forward);
	const plane = RECOGNIZER_SIZE * RECOGNIZER_SIZE;
	const input = new Float32Array(3 * plane);
	for (let row = 0; row < RECOGNIZER_SIZE; row++) {
		for (let column = 0; column < RECOGNIZER_SIZE; column++) {
			const [sourceX, sourceY] = applySimilarity(inverse, [column, row]);
			const pixel = row * RECOGNIZER_SIZE + column;
			for (let channel = 0; channel < 3; channel++) {
				input[channel * plane + pixel] = sample(rgba, width, height, sourceX, sourceY, channel);
			}
		}
	}
	return input;
}

/** The inverse of a similarity transform. */
export function invert(transform: Similarity): Similarity {
	const scale = transform.a * transform.a + transform.b * transform.b;
	const a = transform.a / scale;
	const b = -transform.b / scale;
	return {
		a,
		b,
		tx: -(a * transform.tx - b * transform.ty),
		ty: -(b * transform.tx + a * transform.ty)
	};
}

/** Bilinear sample of one channel; outside the frame is black, as warpAffine's border. */
function sample(
	rgba: ArrayLike<number>,
	width: number,
	height: number,
	x: number,
	y: number,
	channel: number
): number {
	const left = Math.floor(x);
	const top = Math.floor(y);
	const right = x - left;
	const down = y - top;
	const topRow =
		(1 - right) * pixel(rgba, width, height, left, top, channel) +
		right * pixel(rgba, width, height, left + 1, top, channel);
	const bottomRow =
		(1 - right) * pixel(rgba, width, height, left, top + 1, channel) +
		right * pixel(rgba, width, height, left + 1, top + 1, channel);
	return (1 - down) * topRow + down * bottomRow;
}

/** One channel of one pixel, 0 outside the frame. */
function pixel(
	rgba: ArrayLike<number>,
	width: number,
	height: number,
	x: number,
	y: number,
	channel: number
): number {
	if (x < 0 || y < 0 || x >= width || y >= height) return 0;
	return rgba[(y * width + x) * 4 + channel];
}

/** Scales a vector to unit length; an all-zero vector stays as it is. */
export function normalize(vector: ArrayLike<number>): number[] {
	let norm = 0;
	for (let index = 0; index < vector.length; index++) norm += vector[index] * vector[index];
	norm = Math.sqrt(norm);
	const unit: number[] = [];
	for (let index = 0; index < vector.length; index++) {
		if (norm === 0) {
			unit.push(vector[index]);
			continue;
		}
		unit.push(vector[index] / norm);
	}
	return unit;
}

/** Cosine similarity of two unit vectors. */
export function cosine(first: ArrayLike<number>, second: ArrayLike<number>): number {
	let dot = 0;
	for (let index = 0; index < first.length; index++) dot += first[index] * second[index];
	return dot;
}
