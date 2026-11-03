// Logs one playback session of a library video or audio for the recommendation engine
// and the resume point: progress every PROGRESS_EVERY_SECONDS of playing, a
// heatSpan per stretch that actually played (closed by a pause, seek, the end
// or leaving), and finish at the end.
import { logEvent, flushEvents } from '$lib/events';

const PROGRESS_EVERY_SECONDS = 15;
// Stretches shorter than this are a seek landing, not watching.
const MIN_SPAN_SECONDS = 1;
// A larger jump between two time updates is a seek, not playing time.
const MAX_TICK_SECONDS = 2;

/** Starts logging video's playback of the library video (or audio, by kind) videoId. Returns the stop. */
export function trackWatch(
	video: HTMLMediaElement,
	videoId: string,
	kind: 'video' | 'audio' = 'video'
): () => void {
	const subject = { kind, id: videoId } as const;
	const sessionId = crypto.randomUUID();
	let spanStart: number | null = null;
	let lastPosition = video.currentTime;
	let playedSinceProgress = 0;

	/** Logs the stretch played since spanStart, if it's long enough. */
	function closeSpan() {
		if (spanStart === null) return;
		if (lastPosition - spanStart >= MIN_SPAN_SECONDS) {
			logEvent({
				kind: 'heatSpan',
				subject,
				data: { start: spanStart, end: lastPosition },
				sessionId
			});
		}
		spanStart = null;
	}

	/** Logs the resume point. */
	function logProgress() {
		playedSinceProgress = 0;
		logEvent({ kind: 'progress', subject, data: { position: lastPosition }, sessionId });
	}

	/** Counts playing time and logs progress every PROGRESS_EVERY_SECONDS of it. */
	function onTimeUpdate() {
		if (video.seeking) return;
		const tick = video.currentTime - lastPosition;
		lastPosition = video.currentTime;
		if (tick <= 0 || tick > MAX_TICK_SECONDS) return;
		playedSinceProgress += tick;
		if (playedSinceProgress >= PROGRESS_EVERY_SECONDS) logProgress();
	}

	const handlers: Record<string, () => void> = {
		// Also fires after buffering; the stretch then goes on.
		playing: () => {
			if (spanStart === null) spanStart = video.currentTime;
			lastPosition = video.currentTime;
		},
		timeupdate: onTimeUpdate,
		pause: () => {
			closeSpan();
			logProgress();
		},
		// lastPosition still holds where the seek started.
		seeking: closeSpan,
		seeked: () => {
			lastPosition = video.currentTime;
			if (!video.paused) spanStart = video.currentTime;
		},
		ended: () => {
			lastPosition = video.duration;
			closeSpan();
			logEvent({ kind: 'finish', subject, data: { position: video.duration }, sessionId });
		}
	};

	/** Logs what's still open and sends it, when leaving the page or the video. */
	function finishSession() {
		closeSpan();
		if (playedSinceProgress > 0) logProgress();
		void flushEvents();
	}

	for (const [name, handler] of Object.entries(handlers)) {
		video.addEventListener(name, handler);
	}
	window.addEventListener('pagehide', finishSession);
	if (!video.paused) spanStart = video.currentTime;

	return () => {
		for (const [name, handler] of Object.entries(handlers)) {
			video.removeEventListener(name, handler);
		}
		window.removeEventListener('pagehide', finishSession);
		finishSession();
	};
}
