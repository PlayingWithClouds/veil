/**
 * Svelte action: sideways swipes on the element step through a page's own tabs
 * (performer/studio pages). The element is marked data-swipe-ignore so the
 * global tab/drawer swipes leave it alone, except that a swipe from the left
 * edge still opens the drawer. Rows that can scroll sideways keep their swipes.
 */
import {
	DIRECTION_RATIO,
	EDGE_PIXELS,
	FLING_VELOCITY,
	SLOP_PIXELS,
	scrollsSideways
} from '$lib/swipeGestures';

// Finger travel that switches tabs without a fling.
const COMMIT_DISTANCE = 70;

export interface SwipeTabsOptions {
	// Finger moved left: show the next tab.
	onnext: () => void;
	// Finger moved right: show the previous tab.
	onprevious: () => void;
}

export function swipeTabs(element: HTMLElement, options: SwipeTabsOptions) {
	let current = options;
	let startX = 0;
	let startY = 0;
	let lastX = 0;
	let lastTime = 0;
	let velocity = 0;
	// Null until the finger leaves the slop radius, then whether this is our swipe.
	let tracking: boolean | null = null;
	let target: Element | null = null;

	element.dataset.swipeIgnore = '';

	function onTouchStart(event: TouchEvent) {
		tracking = null;
		const touch = event.touches[0];
		if (event.touches.length !== 1 || touch.clientX <= EDGE_PIXELS) {
			tracking = false;
			return;
		}
		startX = touch.clientX;
		startY = touch.clientY;
		lastX = startX;
		lastTime = event.timeStamp;
		velocity = 0;
		target = event.target instanceof Element ? event.target : null;
	}

	/** Decides once the finger leaves the slop radius whether the swipe is ours. */
	function decide(deltaX: number, deltaY: number) {
		if (Math.abs(deltaX) < SLOP_PIXELS && Math.abs(deltaY) < SLOP_PIXELS) return;
		const horizontal = Math.abs(deltaX) > Math.abs(deltaY) * DIRECTION_RATIO;
		tracking = horizontal && !(target && scrollsSideways(target, deltaX));
	}

	function onTouchMove(event: TouchEvent) {
		if (tracking === false) return;
		const touch = event.touches[0];
		if (tracking === null) decide(touch.clientX - startX, touch.clientY - startY);
		if (!tracking) return;
		const elapsed = event.timeStamp - lastTime;
		if (elapsed > 0) velocity = (touch.clientX - lastX) / elapsed;
		lastX = touch.clientX;
		lastTime = event.timeStamp;
	}

	function onTouchEnd() {
		if (!tracking) return;
		tracking = null;
		const deltaX = lastX - startX;
		const far = Math.abs(deltaX) > COMMIT_DISTANCE;
		const flung = Math.abs(velocity) > FLING_VELOCITY && Math.sign(velocity) === Math.sign(deltaX);
		if (!far && !flung) return;
		if (deltaX < 0) {
			current.onnext();
		} else {
			current.onprevious();
		}
	}

	element.addEventListener('touchstart', onTouchStart, { passive: true });
	element.addEventListener('touchmove', onTouchMove, { passive: true });
	element.addEventListener('touchend', onTouchEnd);
	element.addEventListener('touchcancel', onTouchEnd);

	return {
		update(next: SwipeTabsOptions) {
			current = next;
		},
		destroy() {
			element.removeEventListener('touchstart', onTouchStart);
			element.removeEventListener('touchmove', onTouchMove);
			element.removeEventListener('touchend', onTouchEnd);
			element.removeEventListener('touchcancel', onTouchEnd);
		}
	};
}
