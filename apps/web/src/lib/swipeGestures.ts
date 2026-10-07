/**
 * Horizontal swipes on the phone layout. A swipe from the left edge drags the
 * drawer open; any other sideways swipe drags the page towards the neighbouring
 * tab of the tab bar and switches to it on release. Where there is no tab to
 * the left (Home, pages outside the tab bar) a right swipe opens the drawer
 * instead. While the drawer is open, a left swipe drags it closed.
 */
import { get } from 'svelte/store';
import { goto, preloadData } from '$app/navigation';
import { drawerDrag, drawerOpen } from '$lib/stores/viewport';
import { activeTabIndex, tabs } from '$lib/tabs';

// Movement before a touch counts as a swipe or a scroll.
export const SLOP_PIXELS = 10;
// Horizontal movement must dominate vertical by this factor to be a swipe.
export const DIRECTION_RATIO = 1.2;
// Release speed (px/ms) that settles a swipe regardless of distance.
export const FLING_VELOCITY = 0.3;
// Touches starting this close to the left edge always pull the drawer.
export const EDGE_PIXELS = 24;
// Share of the page width a tab swipe must cover to switch without a fling.
const TAB_COMMIT_FRACTION = 0.3;
// How far the page fades out as it follows the finger across its full width.
const TAB_DRAG_FADE = 0.4;

// Touches starting here keep their own horizontal gestures.
const SWIPE_IGNORE = 'video, input, textarea, [contenteditable], [role="dialog"], [data-swipe-ignore]';

type Mode = 'drawer' | 'tab';

type Swipe = {
	target: Element;
	startX: number;
	startY: number;
	startOpen: boolean;
	// Null until the finger leaves the slop radius.
	mode: Mode | null;
	// For tab swipes: where the swipe leads, and which way (1 = finger moving right).
	tabHref: string | null;
	tabDirection: number;
	offset: number;
	lastX: number;
	lastTime: number;
	velocity: number;
};

let swipe: Swipe | null = null;

/** The element holding the page content, which follows tab swipes. */
function pageElement(): HTMLElement | null {
	return document.querySelector<HTMLElement>('[data-swipe-page]');
}

/** Width of the drawer panel, the distance a full drawer swipe covers. */
function drawerWidth(): number {
	const panel = document.querySelector<HTMLElement>('[data-drawer-panel]');
	if (!panel) return 0;
	return panel.offsetWidth;
}

/** Whether an ancestor of the target can still scroll sideways in the swipe's direction. */
export function scrollsSideways(target: Element, deltaX: number): boolean {
	for (let element: Element | null = target; element; element = element.parentElement) {
		if (element.scrollWidth <= element.clientWidth) continue;
		const overflow = getComputedStyle(element).overflowX;
		if (overflow !== 'auto' && overflow !== 'scroll') continue;
		// Finger moving right scrolls towards the start, left towards the end.
		if (deltaX > 0 && element.scrollLeft > 0) return true;
		if (deltaX < 0 && element.scrollLeft + element.clientWidth < element.scrollWidth - 1) return true;
	}
	return false;
}

/** The tab next to the current one in the swipe's direction, if there is one. */
function neighbourTabHref(deltaX: number): string | null {
	const index = activeTabIndex(window.location.pathname);
	if (index === -1) return null;
	// Finger moving right reveals the tab to the left.
	const neighbour = tabs[index - Math.sign(deltaX)];
	if (!neighbour) return null;
	return neighbour.href;
}

/** What a sideways swipe in this direction does, or null to leave it to the page. */
function chooseMode(active: Swipe, deltaX: number): Mode | null {
	if (active.startOpen) {
		if (deltaX < 0) return 'drawer';
		return null;
	}
	if (deltaX > 0 && active.startX <= EDGE_PIXELS) return 'drawer';
	if (active.target.closest(SWIPE_IGNORE) || scrollsSideways(active.target, deltaX)) return null;
	active.tabHref = neighbourTabHref(deltaX);
	active.tabDirection = Math.sign(deltaX);
	if (active.tabHref) {
		// Load the tab while the finger is still moving, so release switches at once.
		preloadData(active.tabHref).catch(() => {});
		return 'tab';
	}
	if (deltaX > 0) return 'drawer';
	return null;
}

/** Decides once the finger leaves the slop radius whether this touch is ours. */
function decide(active: Swipe, deltaX: number, deltaY: number): boolean {
	if (Math.abs(deltaX) < SLOP_PIXELS && Math.abs(deltaY) < SLOP_PIXELS) return false;
	if (Math.abs(deltaX) > Math.abs(deltaY) * DIRECTION_RATIO) {
		active.mode = chooseMode(active, deltaX);
	}
	if (active.mode) return true;
	swipe = null;
	return false;
}

/** Moves the page with the finger, fading it slightly. */
function dragPage(offset: number) {
	const page = pageElement();
	if (!page) return;
	page.style.translate = `${offset}px 0`;
	page.style.opacity = String(1 - (Math.abs(offset) / page.clientWidth) * TAB_DRAG_FADE);
}

/** Moves the drawer with the finger. */
function dragDrawer(active: Swipe, deltaX: number) {
	const width = drawerWidth();
	if (width === 0) return;
	let start = 0;
	if (active.startOpen) start = 1;
	drawerDrag.set(Math.min(1, Math.max(0, start + deltaX / width)));
}

function onTouchStart(event: TouchEvent) {
	swipe = null;
	if (event.touches.length !== 1 || !(event.target instanceof Element)) return;
	const touch = event.touches[0];
	swipe = {
		target: event.target,
		startX: touch.clientX,
		startY: touch.clientY,
		startOpen: get(drawerOpen),
		mode: null,
		tabHref: null,
		tabDirection: 0,
		offset: 0,
		lastX: touch.clientX,
		lastTime: event.timeStamp,
		velocity: 0
	};
}

function onTouchMove(event: TouchEvent) {
	const active = swipe;
	if (!active) return;
	const touch = event.touches[0];
	const deltaX = touch.clientX - active.startX;
	if (!active.mode && !decide(active, deltaX, touch.clientY - active.startY)) return;

	if (event.cancelable) event.preventDefault();
	const elapsed = event.timeStamp - active.lastTime;
	if (elapsed > 0) active.velocity = (touch.clientX - active.lastX) / elapsed;
	active.lastX = touch.clientX;
	active.lastTime = event.timeStamp;

	if (active.mode === 'drawer') {
		dragDrawer(active, deltaX);
		return;
	}
	// Only towards the neighbour; dragging back past the start stops there.
	if (Math.sign(deltaX) === active.tabDirection) {
		active.offset = deltaX;
	} else {
		active.offset = 0;
	}
	dragPage(active.offset);
}

/** Opens or closes the drawer on release, by fling or by distance. */
function settleDrawer(active: Swipe) {
	const progress = get(drawerDrag);
	if (progress === null) return;
	if (active.velocity > FLING_VELOCITY) {
		drawerOpen.set(true);
	} else if (active.velocity < -FLING_VELOCITY) {
		drawerOpen.set(false);
	} else {
		drawerOpen.set(progress > 0.5);
	}
	drawerDrag.set(null);
}

/** Whether a released tab swipe went far or fast enough to switch. */
function tabSwipeCommits(active: Swipe, page: HTMLElement): boolean {
	if (active.offset === 0) return false;
	if (Math.abs(active.offset) > page.clientWidth * TAB_COMMIT_FRACTION) return true;
	const flungForward = Math.sign(active.velocity) === Math.sign(active.offset);
	return flungForward && Math.abs(active.velocity) > FLING_VELOCITY;
}

/** Springs the page back to rest after an abandoned tab swipe. */
function springBack(page: HTMLElement, offset: number) {
	const opacity = page.style.opacity;
	page.style.translate = '';
	page.style.opacity = '';
	page.animate(
		[
			{ translate: `${offset}px 0`, opacity },
			{ translate: '0 0', opacity: '1' }
		],
		{ duration: 320, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' }
	);
}

/**
 * Switches to the neighbouring tab and slides the new page in from the side
 * the finger came from. The old page stays where the finger left it until the
 * new one has rendered, so there is never a blank frame in between.
 */
async function switchTab(page: HTMLElement, offset: number, href: string) {
	const direction = Math.sign(offset);
	await goto(href);
	page.style.translate = '';
	page.style.opacity = '';
	page.animate(
		[
			{ translate: `${-direction * page.clientWidth * 0.25}px 0`, opacity: '0.4' },
			{ translate: '0 0', opacity: '1' }
		],
		{ duration: 320, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' }
	);
}

/** Switches tabs or springs the page back on release. */
function settleTab(active: Swipe) {
	const page = pageElement();
	if (!page) return;
	if (active.tabHref && tabSwipeCommits(active, page)) {
		switchTab(page, active.offset, active.tabHref);
		return;
	}
	springBack(page, active.offset);
}

function onTouchEnd() {
	const active = swipe;
	swipe = null;
	if (active?.mode === 'drawer') settleDrawer(active);
	if (active?.mode === 'tab') settleTab(active);
}

/** Starts listening for swipes. Returns the cleanup. */
export function listenForSwipes(): () => void {
	window.addEventListener('touchstart', onTouchStart, { passive: true });
	// Not passive: a sideways drag has to cancel the page's vertical scroll.
	window.addEventListener('touchmove', onTouchMove, { passive: false });
	window.addEventListener('touchend', onTouchEnd);
	window.addEventListener('touchcancel', onTouchEnd);
	return () => {
		window.removeEventListener('touchstart', onTouchStart);
		window.removeEventListener('touchmove', onTouchMove);
		window.removeEventListener('touchend', onTouchEnd);
		window.removeEventListener('touchcancel', onTouchEnd);
	};
}
