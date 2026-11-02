import { describe, expect, test } from 'bun:test';
import * as ort from 'onnxruntime-web';
import {
	alignedInput,
	applySimilarity,
	cosine,
	FACE_TEMPLATE,
	invert,
	normalize,
	similarityTransform
} from '../src/lib/faces/align';
import {
	decodeDetections,
	detectorInput,
	isUsableFace,
	letterbox,
	nonMaxSuppression,
	overlap,
	unletterbox,
	type FaceBox
} from '../src/lib/faces/detect';
import { FaceEngine } from '../src/lib/faces/engine';
import { grayscale, hashDistance, perceptualHash } from '../src/lib/faces/phash';
import {
	compactVector,
	frameSize,
	pickStream,
	qualityHeight,
	regionOf,
	sampleTimes,
	sheetLayout
} from '../src/lib/faces/plan';

/** A face box with landmarks placed like a frontal face. */
function frontalFace(x: number, y: number, side: number, score = 0.9): FaceBox {
	return {
		x,
		y,
		width: side,
		height: side,
		score,
		landmarks: [
			[x + side * 0.3, y + side * 0.4],
			[x + side * 0.7, y + side * 0.4],
			[x + side * 0.5, y + side * 0.6],
			[x + side * 0.35, y + side * 0.8],
			[x + side * 0.65, y + side * 0.8]
		]
	};
}

/** YuNet outputs for a 640 square with one confident cell at stride 32, all else silent. */
function yunetOutputs(cell: number, box: number[], keypoints: number[]) {
	const outputs: Record<string, { data: Float32Array }> = {};
	for (const stride of [8, 16, 32]) {
		const cells = (640 / stride) ** 2;
		outputs[`cls_${stride}`] = { data: new Float32Array(cells) };
		outputs[`obj_${stride}`] = { data: new Float32Array(cells) };
		outputs[`bbox_${stride}`] = { data: new Float32Array(cells * 4) };
		outputs[`kps_${stride}`] = { data: new Float32Array(cells * 10) };
	}
	outputs.cls_32.data[cell] = 0.95;
	outputs.obj_32.data[cell] = 0.9;
	outputs.bbox_32.data.set(box, cell * 4);
	outputs.kps_32.data.set(keypoints, cell * 10);
	return outputs;
}

/** An RGBA image of a horizontal gradient with a bright square at (left, top). */
function gradientImage(
	width: number,
	height: number,
	left: number,
	top: number
): Uint8ClampedArray {
	const rgba = new Uint8ClampedArray(width * height * 4);
	for (let y = 0; y < height; y++) {
		for (let x = 0; x < width; x++) {
			const inSquare = x >= left && x < left + width / 4 && y >= top && y < top + height / 4;
			let value = Math.round((x / width) * 200);
			if (inSquare) value = 255;
			rgba.set([value, value, value, 255], (y * width + x) * 4);
		}
	}
	return rgba;
}

describe('face detection', () => {
	test('a confident cell decodes into a box and landmarks', () => {
		// Cell 21 at stride 32 is column 1, row 1 of the 20-wide grid.
		const faces = decodeDetections(
			yunetOutputs(21, [0.5, 0.5, Math.log(2), Math.log(3)], [0, 0, 1, 0, 0.5, 0.5, 0, 1, 1, 1]),
			0.8
		);
		expect(faces).toHaveLength(1);
		const [face] = faces;
		expect(face.score).toBeCloseTo(Math.sqrt(0.95 * 0.9));
		expect(face.x).toBeCloseTo(48 - 32);
		expect(face.y).toBeCloseTo(48 - 48);
		expect(face.width).toBeCloseTo(64);
		expect(face.height).toBeCloseTo(96);
		expect(face.landmarks[0]).toEqual([32, 32]);
		expect(face.landmarks[4]).toEqual([64, 64]);
		expect(
			decodeDetections(yunetOutputs(21, [0, 0, 0, 0], new Array(10).fill(0)), 0.95)
		).toHaveLength(0);
	});

	test('overlapping boxes keep the best, letterboxing maps back', () => {
		const strong = frontalFace(0, 0, 100, 0.95);
		const weak = frontalFace(10, 10, 100, 0.85);
		const apart = frontalFace(300, 300, 100, 0.8);
		expect(overlap(strong, weak)).toBeGreaterThan(0.6);
		expect(nonMaxSuppression([weak, apart, strong], 0.3)).toEqual([strong, apart]);
		const box = letterbox(1280, 720);
		expect(box).toEqual({ scale: 0.5, width: 640, height: 360 });
		expect(unletterbox(frontalFace(10, 20, 50), box).landmarks[2]).toEqual([70, 100]);
	});

	test('the detector input is planar BGR, letterboxed', () => {
		const rgba = new Uint8ClampedArray([10, 20, 30, 255, 40, 50, 60, 255]);
		// 2×1 scales by 2 into the top half of a 4×4 square.
		const input = detectorInput(rgba, 2, 1, 4);
		expect([input[0], input[16], input[32]]).toEqual([30, 20, 10]);
		expect([input[1], input[2], input[18], input[34]]).toEqual([30, 60, 50, 40]);
		expect(input[5]).toBe(30);
		expect(input[8]).toBe(0);
	});

	test('small and turned-away faces are not embedded', () => {
		expect(isUsableFace(frontalFace(0, 0, 100), 40)).toBe(true);
		expect(isUsableFace(frontalFace(0, 0, 30), 40)).toBe(false);
		const profile = frontalFace(0, 0, 100);
		profile.landmarks[2] = [profile.landmarks[1][0] + 5, profile.landmarks[2][1]];
		expect(isUsableFace(profile, 40)).toBe(false);
	});
});

describe('face alignment', () => {
	test('the similarity transform recovers rotation, scale and shift', () => {
		const angle = 0.3;
		const truth = { a: 1.5 * Math.cos(angle), b: 1.5 * Math.sin(angle), tx: 12, ty: -7 };
		const moved = FACE_TEMPLATE.map((point) => applySimilarity(truth, point));
		const found = similarityTransform(FACE_TEMPLATE, moved);
		expect(found.a).toBeCloseTo(truth.a);
		expect(found.b).toBeCloseTo(truth.b);
		expect(found.tx).toBeCloseTo(truth.tx);
		const back = applySimilarity(invert(found), moved[2]);
		expect(back[0]).toBeCloseTo(FACE_TEMPLATE[2][0]);
		expect(back[1]).toBeCloseTo(FACE_TEMPLATE[2][1]);
	});

	test('a frame already in template position is copied as planar RGB', () => {
		const rgba = new Uint8ClampedArray(112 * 112 * 4);
		for (let pixel = 0; pixel < 112 * 112; pixel++) rgba.set([pixel % 112, 7, 200, 255], pixel * 4);
		const input = alignedInput(rgba, 112, 112, FACE_TEMPLATE);
		expect(input[5]).toBeCloseTo(5);
		expect(input[112 * 112 + 50]).toBeCloseTo(7);
		expect(input[2 * 112 * 112 + 50]).toBeCloseTo(200);
	});

	test('vectors are normalised and compared by cosine', () => {
		const first = normalize([3, 4]);
		expect(first).toEqual([0.6, 0.8]);
		expect(cosine(first, normalize([6, 8]))).toBeCloseTo(1);
		expect(normalize([0, 0])).toEqual([0, 0]);
	});
});

describe('perceptual hash', () => {
	test('re-encoded copies hash alike, other pictures far apart', () => {
		const original = gradientImage(160, 90, 20, 10);
		const brighter = original.map((value, index) =>
			index % 4 === 3 ? value : Math.min(255, value + 12)
		);
		const other = gradientImage(160, 90, 110, 60).map((value, index) =>
			index % 4 === 3 ? value : 255 - value
		);
		const hash = perceptualHash(grayscale(original, 160, 90), 160, 90);
		expect(hash).toMatch(/^[0-9a-f]{16}$/);
		expect(
			hashDistance(hash, perceptualHash(grayscale(brighter, 160, 90), 160, 90))
		).toBeLessThanOrEqual(4);
		expect(hashDistance(hash, perceptualHash(grayscale(other, 160, 90), 160, 90))).toBeGreaterThan(
			16
		);
	});
});

describe('scan plan', () => {
	test('frames sit in the middle of equal spans and tile a 5x5 sheet', () => {
		expect(sampleTimes(100, 4)).toEqual([12.5, 37.5, 62.5, 87.5]);
		expect(sampleTimes(250)).toHaveLength(25);
		expect(sheetLayout(25, 1920, 1080)).toEqual({
			columns: 5,
			rows: 5,
			tileWidth: 160,
			tileHeight: 90
		});
		expect(frameSize(1920, 1080)).toEqual({ width: 640, height: 360 });
		expect(frameSize(300, 400)).toEqual({ width: 300, height: 400 });
	});

	test('the smallest stream of at least 480p is sampled', () => {
		const streams = [
			{ url: '1080', quality: '1080p' },
			{ url: '480', quality: '480p' },
			{ url: '720', quality: '720p' },
			{ url: '240', quality: '240p' }
		];
		expect(pickStream(streams)?.url).toBe('480');
		expect(
			pickStream([
				{ url: 'low', quality: '240p' },
				{ url: 'lower', quality: '144p' }
			])?.url
		).toBe('low');
		expect(pickStream([{ url: 'unknown', quality: null }])?.url).toBe('unknown');
		expect(pickStream([])).toBeNull();
		expect(qualityHeight('4K')).toBe(2160);
	});

	test('regions are shares of the frame, clamped, and vectors compact', () => {
		expect(regionOf(frontalFace(-10, 20, 110, 0.91234), 200, 100)).toEqual({
			x: 0,
			y: 0.2,
			width: 0.5,
			height: 0.8,
			score: 0.9123
		});
		expect(compactVector([0.123456789, -0.5])).toEqual([0.12346, -0.5]);
	});
});

describe('face models', () => {
	test('both models load and give a 128-dimensional unit vector', async () => {
		ort.env.wasm.numThreads = 1;
		const weights = async (name: string) =>
			new Uint8Array(
				await Bun.file(`${import.meta.dir}/../static/models/faces/${name}`).arrayBuffer()
			);
		const engine = await FaceEngine.create(
			ort,
			await weights('face_detection_yunet_2023mar.onnx'),
			await weights('face_recognition_sface_2021dec_int8.onnx')
		);
		const frame = { rgba: gradientImage(320, 240, 100, 80), width: 320, height: 240 };
		expect(await engine.detect(frame)).toEqual([]);
		const vector = await engine.embed(frame, frontalFace(100, 80, 80));
		expect(vector).toHaveLength(128);
		expect(cosine(vector, vector)).toBeCloseTo(1);
	});
});
