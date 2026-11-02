import type { Action } from 'svelte/action';
import { logShown, type Subject, type Shown } from '$lib/events';

/** What use:impression logs: the record, and where it's shown (nothing when absent). */
export interface ImpressionOptions {
	subject: Subject;
	shown: Shown | null;
}

// Half a card on screen counts as shown.
const VISIBLE_SHARE = 0.5;

/** Logs the record as shown on its surface once the element is half on screen. */
export const impression: Action<HTMLElement, ImpressionOptions> = (element, initial) => {
	let options = initial;
	const observer = new IntersectionObserver(
		(entries) => {
			if (!options.shown || !entries.some((entry) => entry.isIntersecting)) return;
			logShown(options.subject, options.shown);
		},
		{ threshold: VISIBLE_SHARE }
	);
	observer.observe(element);
	return {
		update(next) {
			options = next;
		},
		destroy() {
			observer.disconnect();
		}
	};
};
