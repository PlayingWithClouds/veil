/**
 * Web worker running the face models off the main thread: it loads the
 * weights on the first frame and answers each frame with its faces. One
 * WebAssembly thread keeps the CPU cost of background scans low.
 */
/// <reference lib="webworker" />
import * as ort from 'onnxruntime-web/wasm';
import wasmUrl from 'onnxruntime-web/ort-wasm-simd-threaded.wasm?url';
import { FaceEngine, type Frame, type FoundFace } from './engine';

/** Asks for the faces of a frame; the weights' URLs come along so the worker needs no app paths. */
export interface FaceRequest {
	id: number;
	frame: Frame;
	detectorUrl: string;
	recognizerUrl: string;
}

/** The faces found in a requested frame, or why none could be. */
export interface FaceResponse {
	id: number;
	faces?: FoundFace[];
	error?: string;
}

ort.env.wasm.wasmPaths = { wasm: wasmUrl };
ort.env.wasm.numThreads = 1;
ort.env.logLevel = 'error';

let engine: Promise<FaceEngine> | null = null;

/** Loads both models once. */
function loadEngine(detectorUrl: string, recognizerUrl: string): Promise<FaceEngine> {
	if (engine === null) {
		engine = Promise.all([weights(detectorUrl), weights(recognizerUrl)]).then(
			([detector, recognizer]) => FaceEngine.create(ort, detector, recognizer)
		);
		engine.catch(() => {
			engine = null;
		});
	}
	return engine;
}

/** Fetches a model's weights. */
async function weights(url: string): Promise<Uint8Array> {
	const response = await fetch(url);
	if (!response.ok) throw new Error(`Couldn't load the face model ${url}: ${response.status}`);
	return new Uint8Array(await response.arrayBuffer());
}

self.onmessage = async (event: MessageEvent<FaceRequest>) => {
	const request = event.data;
	try {
		const loaded = await loadEngine(request.detectorUrl, request.recognizerUrl);
		const faces = await loaded.analyze(request.frame);
		self.postMessage({ id: request.id, faces } satisfies FaceResponse);
	} catch (error) {
		self.postMessage({ id: request.id, error: String(error) } satisfies FaceResponse);
	}
};
