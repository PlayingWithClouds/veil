/**
 * Runs the two face models on a frame: YuNet finds faces, SFace embeds each
 * usable one as a 128-dimensional unit vector. Takes the ONNX Runtime module
 * and the weights as arguments, so the browser worker and `bun test` share it.
 */
import type * as Ort from 'onnxruntime-web';
import { alignedInput, normalize, RECOGNIZER_SIZE } from './align';
import {
	decodeDetections,
	detectorInput,
	DETECTOR_SIZE,
	isUsableFace,
	letterbox,
	unletterbox,
	type FaceBox
} from './detect';

/**
 * Names the detector, recognizer, alignment and thresholds together; stored
 * with every vector so vectors of different pipelines never mix. Bump it when
 * any of them changes.
 */
export const FACE_MODEL = 'yunet-2023mar+sface-2021dec-int8/1';

/** Detections scoring below this are dropped (OpenCV's samples use 0.9; video frames are softer). */
const DETECTION_THRESHOLD = 0.8;

/** Faces with a shorter side (frame pixels) are too small to embed. */
const MIN_FACE_SIDE = 40;

/** At most this many faces of one frame are embedded, largest first. */
const MAX_FACES_PER_FRAME = 4;

/** An RGBA frame, as ImageData holds it. */
export interface Frame {
	rgba: Uint8ClampedArray | Uint8Array;
	width: number;
	height: number;
}

/** A face found in a frame and its embedding. */
export interface FoundFace {
	box: FaceBox;
	vector: number[];
}

/** The two loaded models. */
export class FaceEngine {
	private constructor(
		private readonly ort: typeof Ort,
		private readonly detector: Ort.InferenceSession,
		private readonly recognizer: Ort.InferenceSession
	) {}

	/** Loads both models from their weights. */
	static async create(
		ort: typeof Ort,
		detectorWeights: Uint8Array,
		recognizerWeights: Uint8Array
	): Promise<FaceEngine> {
		// Severity 3 (errors only): SFace's export lists its weights as inputs, which logs a warning per weight.
		const options: Ort.InferenceSession.SessionOptions = {
			executionProviders: ['wasm'],
			graphOptimizationLevel: 'all',
			logSeverityLevel: 3
		};
		const detector = await ort.InferenceSession.create(detectorWeights, options);
		const recognizer = await ort.InferenceSession.create(recognizerWeights, options);
		return new FaceEngine(ort, detector, recognizer);
	}

	/** Finds the usable faces of a frame and embeds them. */
	async analyze(frame: Frame): Promise<FoundFace[]> {
		const faces = await this.detect(frame);
		const usable = faces
			.filter((face) => isUsableFace(face, MIN_FACE_SIDE))
			.sort((first, second) => second.width * second.height - first.width * first.height)
			.slice(0, MAX_FACES_PER_FRAME);
		const found: FoundFace[] = [];
		for (const box of usable) {
			found.push({ box, vector: await this.embed(frame, box) });
		}
		return found;
	}

	/** Every face YuNet finds in a frame, in frame pixels. */
	async detect(frame: Frame): Promise<FaceBox[]> {
		const input = detectorInput(frame.rgba, frame.width, frame.height);
		const tensor = new this.ort.Tensor('float32', input, [1, 3, DETECTOR_SIZE, DETECTOR_SIZE]);
		const outputs = await this.detector.run({ [this.detector.inputNames[0]]: tensor });
		const box = letterbox(frame.width, frame.height);
		return decodeDetections(
			outputs as Record<string, Ort.Tensor & { data: Float32Array }>,
			DETECTION_THRESHOLD
		).map((face) => unletterbox(face, box));
	}

	/** The unit embedding of one face of a frame. */
	async embed(frame: Frame, face: FaceBox): Promise<number[]> {
		const input = alignedInput(frame.rgba, frame.width, frame.height, face.landmarks);
		const tensor = new this.ort.Tensor('float32', input, [1, 3, RECOGNIZER_SIZE, RECOGNIZER_SIZE]);
		const outputs = await this.recognizer.run({ [this.recognizer.inputNames[0]]: tensor });
		return normalize(outputs[this.recognizer.outputNames[0]].data as Float32Array);
	}
}
