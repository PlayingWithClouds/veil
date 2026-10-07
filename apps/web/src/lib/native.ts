import { get } from 'svelte/store';
import { App } from '@capacitor/app';
import { ScreenOrientation } from '@capacitor/screen-orientation';
import { drawerOpen } from '$lib/stores/viewport';
import { closeTopOverlay } from '$lib/stores/overlayStack';
import { isNativeApp } from '$lib/server';

/**
 * Turns the phone to landscape while a landscape video is fullscreen and frees
 * the orientation again afterwards. No-op outside the native app.
 */
export function matchOrientationToVideo(fullscreen: boolean, video: HTMLVideoElement | undefined) {
	if (!isNativeApp()) return;
	if (!fullscreen || !video) {
		ScreenOrientation.unlock();
		return;
	}
	if (video.videoWidth > video.videoHeight) {
		ScreenOrientation.lock({ orientation: 'landscape' });
	}
}

/**
 * Handles the Android back button: leaves fullscreen or closes the topmost
 * sheet/overlay (overlayStack) or the drawer first, then walks back through
 * history, and exits the app at its start.
 * Returns the unsubscribe.
 */
export function listenToBackButton(): () => void {
	const listener = App.addListener('backButton', ({ canGoBack }) => {
		if (document.fullscreenElement) {
			document.exitFullscreen();
			return;
		}
		if (closeTopOverlay()) return;
		if (get(drawerOpen)) {
			drawerOpen.set(false);
			return;
		}
		if (canGoBack) {
			history.back();
			return;
		}
		App.exitApp();
	});
	return () => {
		listener.then((handle) => handle.remove());
	};
}
