/**
 * Face detection with YuNet (OpenCV Zoo, face_detection_yunet_2023mar):
 * building its input from a frame and decoding its outputs into face boxes
 * with five landmarks. Pure, so it runs in a worker and under `bun test`.
 */

/** Side of the square YuNet's exported graph takes. */
export const DETECTOR_SIZE = 640;

/** Strides of YuNet's three output heads. */
const STRIDES = [8, 16, 32];

/** A point in pixels. */
export type Point = [number, number];

/**
 * A detected face in pixels of the frame it was found in. Landmarks are the
 * right eye, left eye, nose tip, right and left mouth corner, as YuNet orders
 * them (the template in align.ts expects that order).
 */
export interface FaceBox {
	x: number;
	y: number;
	width: number;
	height: number;
	score: number;
	landmarks: Point[];
}

/** A tensor's data as ONNX Runtime hands it back. */
export interface OutputTensor {
	data: ArrayLike<number>;
}

/**
 * How a frame is fitted into the detector's square: scaled by `scale` into
 * the top-left corner, the rest left black.
 */
export interface Letterbox {
	scale: number;
	width: number;
	height: number;
}

/** Fits a frame of the given size into a square of `size`, keeping its aspect. */
export function letterbox(
	frameWidth: number,
	frameHeight: number,
	size = DETECTOR_SIZE
): Letterbox {
	const scale = Math.min(size / frameWidth, size / frameHeight);
	return { scale, width: Math.round(frameWidth * scale), height: Math.round(frameHeight * scale) };
}

/**
 * The detector's input from an RGBA frame: letterboxed into a size×size
 * square (nearest pixel, frames arrive near that size anyway), planar BGR,
 * 0..255, as OpenCV's FaceDetectorYN feeds it.
 */
export function detectorInput(
	rgba: ArrayLike<number>,
	width: number,
	height: number,
	size = DETECTOR_SIZE
): Float32Array {
	const box = letterbox(width, height, size);
	const plane = size * size;
	const input = new Float32Array(3 * plane);
	for (let row = 0; row < box.height; row++) {
		const sourceRow = Math.min(height - 1, Math.floor(row / box.scale));
		for (let column = 0; column < box.width; column++) {
			const sourceColumn = Math.min(width - 1, Math.floor(column / box.scale));
			const source = (sourceRow * width + sourceColumn) * 4;
			const pixel = row * size + column;
			input[pixel] = rgba[source + 2];
			input[plane + pixel] = rgba[source + 1];
			input[2 * plane + pixel] = rgba[source];
		}
	}
	return input;
}

/**
 * Decodes YuNet's outputs (`cls_8`, `obj_8`, `bbox_8`, `kps_8`, … per stride)
 * into faces scoring at least `threshold`, in detector pixels, overlapping
 * ones suppressed.
 */
export function decodeDetections(
	outputs: Record<string, OutputTensor>,
	threshold: number,
	size = DETECTOR_SIZE
): FaceBox[] {
	const faces: FaceBox[] = [];
	for (const stride of STRIDES) {
		faces.push(...decodeStride(outputs, stride, threshold, size));
	}
	return nonMaxSuppression(faces, 0.3);
}

/** Decodes one stride's head: every grid cell is an anchor-free candidate. */
function decodeStride(
	outputs: Record<string, OutputTensor>,
	stride: number,
	threshold: number,
	size: number
): FaceBox[] {
	const classes = outputs[`cls_${stride}`].data;
	const objects = outputs[`obj_${stride}`].data;
	const boxes = outputs[`bbox_${stride}`].data;
	const keypoints = outputs[`kps_${stride}`].data;
	const columns = Math.floor(size / stride);
	const faces: FaceBox[] = [];
	for (let cell = 0; cell < classes.length; cell++) {
		const score = Math.sqrt(clamp01(classes[cell]) * clamp01(objects[cell]));
		if (score < threshold) continue;
		const column = cell % columns;
		const row = Math.floor(cell / columns);
		faces.push(cellFace(boxes, keypoints, cell, column, row, stride, score));
	}
	return faces;
}

/** The face one grid cell predicts. */
function cellFace(
	boxes: ArrayLike<number>,
	keypoints: ArrayLike<number>,
	cell: number,
	column: number,
	row: number,
	stride: number,
	score: number
): FaceBox {
	const centerX = (column + boxes[cell * 4]) * stride;
	const centerY = (row + boxes[cell * 4 + 1]) * stride;
	const width = Math.exp(boxes[cell * 4 + 2]) * stride;
	const height = Math.exp(boxes[cell * 4 + 3]) * stride;
	const landmarks: Point[] = [];
	for (let point = 0; point < 5; point++) {
		landmarks.push([
			(keypoints[cell * 10 + point * 2] + column) * stride,
			(keypoints[cell * 10 + point * 2 + 1] + row) * stride
		]);
	}
	return { x: centerX - width / 2, y: centerY - height / 2, width, height, score, landmarks };
}

/** Clamps a score into 0..1, as OpenCV does before combining them. */
function clamp01(value: number): number {
	return Math.min(1, Math.max(0, value));
}

/** Keeps the best of faces overlapping by more than `maxOverlap` (IoU). */
export function nonMaxSuppression(faces: FaceBox[], maxOverlap: number): FaceBox[] {
	const sorted = [...faces].sort((first, second) => second.score - first.score);
	const kept: FaceBox[] = [];
	for (const face of sorted) {
		if (kept.every((other) => overlap(face, other) <= maxOverlap)) kept.push(face);
	}
	return kept;
}

/** Intersection over union of two boxes. */
export function overlap(first: FaceBox, second: FaceBox): number {
	const width =
		Math.min(first.x + first.width, second.x + second.width) - Math.max(first.x, second.x);
	const height =
		Math.min(first.y + first.height, second.y + second.height) - Math.max(first.y, second.y);
	if (width <= 0 || height <= 0) return 0;
	const intersection = width * height;
	return intersection / (first.width * first.height + second.width * second.height - intersection);
}

/** Maps a face found in the letterboxed square back to frame pixels. */
export function unletterbox(face: FaceBox, box: Letterbox): FaceBox {
	const scale = 1 / box.scale;
	return {
		x: face.x * scale,
		y: face.y * scale,
		width: face.width * scale,
		height: face.height * scale,
		score: face.score,
		landmarks: face.landmarks.map(([x, y]) => [x * scale, y * scale] as Point)
	};
}

/**
 * Whether a face is worth embedding: big enough to carry detail and turned
 * towards the camera (the nose sits between the eyes), since profile and tiny
 * faces embed poorly and would blur a performer's identity.
 */
export function isUsableFace(face: FaceBox, minSide: number): boolean {
	if (Math.min(face.width, face.height) < minSide) return false;
	const [rightEye, leftEye, nose] = face.landmarks;
	const eyeLeftmost = Math.min(rightEye[0], leftEye[0]);
	const eyeRightmost = Math.max(rightEye[0], leftEye[0]);
	const eyeSpan = eyeRightmost - eyeLeftmost;
	if (eyeSpan < minSide * 0.25) return false;
	const margin = eyeSpan * 0.1;
	return nose[0] > eyeLeftmost + margin && nose[0] < eyeRightmost - margin;
}
