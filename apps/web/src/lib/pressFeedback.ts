/**
 * Touch press feedback: pressable elements shrink slightly while a finger is on
 * them and spring back on release. Driven by the Web Animations API rather than
 * CSS `:active`, so it works regardless of the `transition-*` utilities an
 * element already carries, and releases cleanly when the browser takes the
 * touch over for scrolling (`pointercancel`).
 */

const PRESSABLE = 'button:not(:disabled), a[href], [role="button"], summary';

// The press shrinks every element by about this many pixels on its longest
// side, so small buttons dip visibly and large cards only subtly.
const SHRINK_PIXELS = 5;
const MAX_SHRINK = 0.08;

const PRESS_IN_MS = 90;
const RELEASE_MS = 320;
const RELEASE_EASING = 'cubic-bezier(0.34, 1.56, 0.64, 1)';

/** How far an element scales down while pressed. */
function pressedScale(element: Element): number {
	const bounds = element.getBoundingClientRect();
	const longestSide = Math.max(bounds.width, bounds.height, 1);
	return 1 - Math.min(MAX_SHRINK, SHRINK_PIXELS / longestSide);
}

/** Springs a pressed element back to its resting size. */
function release(element: HTMLElement, scale: number) {
	element.animate([{ scale: String(scale) }, { scale: '1' }], {
		duration: RELEASE_MS,
		easing: RELEASE_EASING
	});
}

/** Shrinks the element until the finger lifts or the browser starts scrolling. */
function press(element: HTMLElement) {
	const scale = pressedScale(element);
	const pressIn = element.animate([{ scale: '1' }, { scale: String(scale) }], {
		duration: PRESS_IN_MS,
		easing: 'ease-out',
		fill: 'forwards'
	});

	/** Ends the press and springs the element back. */
	function onEnd() {
		window.removeEventListener('pointerup', onEnd);
		window.removeEventListener('pointercancel', onEnd);
		pressIn.cancel();
		release(element, scale);
	}
	window.addEventListener('pointerup', onEnd);
	window.addEventListener('pointercancel', onEnd);
}

/** Starts press feedback for touches anywhere in the document. Returns the cleanup. */
export function listenForPressFeedback(): () => void {
	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');

	function onPointerDown(event: PointerEvent) {
		if (event.pointerType !== 'touch' || reducedMotion.matches) return;
		if (!(event.target instanceof Element)) return;
		const element = event.target.closest<HTMLElement>(PRESSABLE);
		if (element) press(element);
	}

	document.addEventListener('pointerdown', onPointerDown, { passive: true });
	return () => document.removeEventListener('pointerdown', onPointerDown);
}
