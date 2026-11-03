import { writable } from 'svelte/store';

/**
 * The picture that lights the app behind the top bar, the sidebar and the
 * page, blurred: what the page features (a room, a top pick). Null is none.
 */
export const ambientGlow = writable<string | null>(null);

/**
 * Lights the app with picture() while the calling component lives; call it
 * during component setup. The glow follows the picture as it changes and
 * goes out when the component does.
 */
export function showAmbientGlow(picture: () => string | null) {
	$effect(() => {
		ambientGlow.set(picture());
	});
	$effect(() => {
		return () => ambientGlow.set(null);
	});
}
