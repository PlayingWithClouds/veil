/**
 * Touch screens have no hover, so scene cards preview when they rest in the
 * middle of the screen instead: the card crossing a narrow centre band for a
 * moment while scrolling has settled plays its preview, one card at a time.
 * Does nothing on devices that can hover.
 */

// How long a card must stay centred before its preview starts.
const DWELL_MS = 700;
// The observed band: the middle 60% of the viewport's height. Wide enough that
// a card is always in it; the one closest to the middle wins.
const CENTER_BAND = '-20% 0px -20% 0px';
// Scrolling counts as settled after this long without a scroll event.
const SCROLL_SETTLE_MS = 120;

type Listener = (centered: boolean) => void;

const listeners = new Map<Element, Listener>();
const inBand = new Set<Element>();
let centeredElement: Element | null = null;
let previewingElement: Element | null = null;
let dwellTimer: ReturnType<typeof setTimeout> | undefined;
let observer: IntersectionObserver | null = null;

/** The card in the band whose middle is closest to the viewport's middle. */
function closestToCenter(): Element | null {
	const viewportMiddle = window.innerHeight / 2;
	let closest: Element | null = null;
	let closestDistance = Infinity;
	for (const element of inBand) {
		const bounds = element.getBoundingClientRect();
		const distance = Math.abs(bounds.top + bounds.height / 2 - viewportMiddle);
		if (distance < closestDistance) {
			closest = element;
			closestDistance = distance;
		}
	}
	return closest;
}

/** Stops the running preview, if any. */
function stopPreview() {
	if (!previewingElement) return;
	listeners.get(previewingElement)?.(false);
	previewingElement = null;
}

/** Re-picks the centred card and restarts the dwell timer when it changed. */
function updateCentered() {
	const next = closestToCenter();
	if (next === centeredElement) return;
	centeredElement = next;
	clearTimeout(dwellTimer);
	stopPreview();
	if (!next) return;
	dwellTimer = setTimeout(() => {
		previewingElement = next;
		listeners.get(next)?.(true);
	}, DWELL_MS);
}

let scrollSettleTimer: ReturnType<typeof setTimeout> | undefined;

/**
 * Re-picks the centred card once scrolling settles: several cards can sit in
 * the band at once, and which is closest changes without any entering or
 * leaving it.
 */
function onScroll() {
	clearTimeout(scrollSettleTimer);
	scrollSettleTimer = setTimeout(updateCentered, SCROLL_SETTLE_MS);
}

/** The shared observer, created on first use. */
function bandObserver(): IntersectionObserver {
	if (observer) return observer;
	// Capture: the page scrolls inside <main>, and scroll events don't bubble.
	document.addEventListener('scroll', onScroll, { capture: true, passive: true });
	observer = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				if (entry.isIntersecting) {
					inBand.add(entry.target);
				} else {
					inBand.delete(entry.target);
				}
			}
			updateCentered();
		},
		{ rootMargin: CENTER_BAND }
	);
	return observer;
}

/**
 * Svelte action: calls `onchange(true)` once the element has rested in the
 * middle of a touch screen, and `onchange(false)` when it leaves.
 */
export function centerPreview(element: HTMLElement, onchange: Listener) {
	if (window.matchMedia('(hover: hover)').matches) return {};
	listeners.set(element, onchange);
	bandObserver().observe(element);
	return {
		update(next: Listener) {
			listeners.set(element, next);
		},
		destroy() {
			observer?.unobserve(element);
			listeners.delete(element);
			inBand.delete(element);
			if (previewingElement === element) previewingElement = null;
			if (centeredElement === element) updateCentered();
		}
	};
}
