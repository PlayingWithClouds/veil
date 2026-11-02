/**
 * The main thread's side of the face worker: frames go in, faces come back.
 * The worker starts on the first frame and stops after a minute without
 * work, so its models don't hold memory while nothing is scanned.
 */
import { asset } from '$app/paths';
import type { Frame, FoundFace } from './engine';
import type { FaceRequest, FaceResponse } from './worker';

/** How long the worker lives without a frame to look at. */
const IDLE_STOP_MS = 60_000;

let worker: Worker | null = null;
let nextId = 0;
let stopTimer: ReturnType<typeof setTimeout> | null = null;
const pending = new Map<
	number,
	{ resolve: (faces: FoundFace[]) => void; reject: (error: Error) => void }
>();

/** Finds and embeds the faces of a frame; its pixels are handed to the worker, not copied. */
export function analyzeFrame(frame: Frame): Promise<FoundFace[]> {
	const running = startWorker();
	const id = nextId++;
	const request: FaceRequest = {
		id,
		frame,
		detectorUrl: asset('/models/faces/face_detection_yunet_2023mar.onnx'),
		recognizerUrl: asset('/models/faces/face_recognition_sface_2021dec_int8.onnx')
	};
	return new Promise((resolve, reject) => {
		pending.set(id, { resolve, reject });
		running.postMessage(request, [frame.rgba.buffer]);
	});
}

/** The running worker, started when needed; its idle stop is pushed back. */
function startWorker(): Worker {
	if (stopTimer !== null) clearTimeout(stopTimer);
	stopTimer = setTimeout(stopWhenIdle, IDLE_STOP_MS);
	if (worker !== null) return worker;
	worker = new Worker(new URL('./worker.ts', import.meta.url), { type: 'module' });
	worker.onmessage = (event: MessageEvent<FaceResponse>) => settle(event.data);
	worker.onerror = (event) => failAll(new Error(event.message || 'The face worker stopped'));
	return worker;
}

/** Hands a response to the frame waiting for it. */
function settle(response: FaceResponse) {
	const waiting = pending.get(response.id);
	if (!waiting) return;
	pending.delete(response.id);
	if (response.error !== undefined) {
		waiting.reject(new Error(response.error));
		return;
	}
	waiting.resolve(response.faces ?? []);
}

/** Fails every waiting frame and drops the worker. */
function failAll(error: Error) {
	for (const waiting of pending.values()) waiting.reject(error);
	pending.clear();
	worker?.terminate();
	worker = null;
}

/** Stops the worker unless frames are still waiting. */
function stopWhenIdle() {
	stopTimer = null;
	if (pending.size > 0) {
		stopTimer = setTimeout(stopWhenIdle, IDLE_STOP_MS);
		return;
	}
	worker?.terminate();
	worker = null;
}
