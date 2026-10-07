/**
 * Svelte action: fades an image in once it has loaded instead of letting it pop
 * in. Images already in the cache show at once; later `src` changes (preview
 * frames cycling) swap without fading.
 */
export function fadeInImage(image: HTMLImageElement) {
	if (image.complete) return;
	image.classList.add('fade-in-image');

	function reveal() {
		image.classList.add('loaded');
	}

	image.addEventListener('load', reveal, { once: true });
	image.addEventListener('error', reveal, { once: true });
	return {
		destroy() {
			image.removeEventListener('load', reveal);
			image.removeEventListener('error', reveal);
		}
	};
}
