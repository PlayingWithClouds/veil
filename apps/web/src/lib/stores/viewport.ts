import { browser } from '$app/environment';
import { derived, readable, writable } from 'svelte/store';

/**
 * When the app uses the phone layout (drawer, bottom nav): below Tailwind's md
 * width, or a short touch screen, i.e. a phone held sideways. Keep in step with
 * the md / max-md variants in routes/layout.css.
 */
export const COMPACT_QUERY = '(width < 720px), ((height < 540px) and (pointer: coarse))';

/** Whether the viewport is phone-sized. Mirrors Tailwind's `md` breakpoint. */
export const isCompact = readable(false, (set) => {
	if (!browser) return;
	const query = window.matchMedia(COMPACT_QUERY);
	set(query.matches);
	const onChange = (event: MediaQueryListEvent) => set(event.matches);
	query.addEventListener('change', onChange);
	return () => query.removeEventListener('change', onChange);
});

/** Whether the sidebar is open as a drawer over the page (phone layout only). */
export const drawerOpen = writable(false);

/**
 * Whether the phone's top bar (and the chips sticking under it) are slid away
 * because the page is being scrolled down. Any scroll up brings them back.
 */
export const chromeHidden = writable(false);

/**
 * Set by pages that start with full-bleed media (the watch page): at the top
 * of the page the top bar stays hidden and the content starts right under the
 * status bar. Scrolling back up mid-page still brings the bar back.
 */
export const immersiveTop = writable(false);

/** Drawer position while a finger drags it (0 = closed, 1 = open); null otherwise. */
export const drawerDrag = writable<number | null>(null);

/** How far the drawer is open, 0 to 1: follows the finger while dragging, else the open state. */
export const drawerProgress = derived([drawerOpen, drawerDrag], ([open, drag]) => {
	if (drag !== null) return drag;
	if (open) return 1;
	return 0;
});
