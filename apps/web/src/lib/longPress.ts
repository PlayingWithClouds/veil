/**
 * Svelte action: calls `onlongpress` when a finger rests on the element for a
 * moment, with a short vibration. The click that would follow the release is
 * swallowed, and the browser's own long-press menu (link/image preview) is
 * suppressed. Mouse and pen are left alone. Moving the finger (scrolling)
 * cancels it.
 */

const HOLD_MS = 450;
const MOVE_TOLERANCE = 10;

export function longPress(element: HTMLElement, onlongpress: () => void) {
	let callback = onlongpress;
	let timer: ReturnType<typeof setTimeout> | undefined;
	let startX = 0;
	let startY = 0;
	// Set when the hold fired, so the click from the release is swallowed.
	let fired = false;

	function cancel() {
		clearTimeout(timer);
		timer = undefined;
	}

	function onPointerDown(event: PointerEvent) {
		if (event.pointerType !== 'touch') return;
		fired = false;
		startX = event.clientX;
		startY = event.clientY;
		timer = setTimeout(() => {
			timer = undefined;
			fired = true;
			navigator.vibrate?.(12);
			callback();
		}, HOLD_MS);
	}

	function onPointerMove(event: PointerEvent) {
		if (timer === undefined) return;
		const moved = Math.hypot(event.clientX - startX, event.clientY - startY);
		if (moved > MOVE_TOLERANCE) cancel();
	}

	function onClick(event: MouseEvent) {
		if (!fired) return;
		fired = false;
		event.preventDefault();
		event.stopPropagation();
	}

	function onContextMenu(event: Event) {
		event.preventDefault();
	}

	element.addEventListener('pointerdown', onPointerDown);
	element.addEventListener('pointermove', onPointerMove);
	element.addEventListener('pointerup', cancel);
	element.addEventListener('pointercancel', cancel);
	element.addEventListener('click', onClick, { capture: true });
	element.addEventListener('contextmenu', onContextMenu);
	element.style.setProperty('-webkit-touch-callout', 'none');

	return {
		update(next: () => void) {
			callback = next;
		},
		destroy() {
			cancel();
			element.removeEventListener('pointerdown', onPointerDown);
			element.removeEventListener('pointermove', onPointerMove);
			element.removeEventListener('pointerup', cancel);
			element.removeEventListener('pointercancel', cancel);
			element.removeEventListener('click', onClick, { capture: true });
			element.removeEventListener('contextmenu', onContextMenu);
		}
	};
}
