import { goto } from '$app/navigation';
import {
	identifyPicture,
	type FaceIdentification,
	type FaceSearchPicture
} from '$lib/faceIdentify';
import { homeSearchUrl } from '$lib/routes';
import { mediaFilter } from '$lib/stores/mediaFilter';
import { searchQuery } from '$lib/stores/searchQuery';

/** The longest side a picked picture is scaled down to before it is sent. */
const MAX_PICTURE_SIDE = 1600;
const JPEG_QUALITY = 0.9;

/**
 * The face search mode's one search: the picture asked about and what the face
 * searches answered. Kept for the session, so coming back from a result finds it.
 */
class FaceSearch {
	/** The picture on screen: a JPEG data URL of a picked file, or the web address searched. */
	pictureUrl = $state<string | null>(null);
	identification = $state<FaceIdentification | null>(null);
	loading = $state(false);
	errorText = $state<string | null>(null);
	// Tells searches apart, so a slow answer never replaces a newer one.
	private requestToken = 0;

	/** Searches a picture file the user picked, dropped or pasted. */
	async searchFile(file: File) {
		if (!file.type.startsWith('image/')) {
			this.errorText = 'That is not a picture.';
			return;
		}
		try {
			const dataUrl = await scaledJpeg(file);
			await this.run(dataUrl, { base64: dataUrl.slice(dataUrl.indexOf(',') + 1) });
		} catch (error) {
			console.error('read picture failed', error);
			this.errorText = 'Could not read that picture.';
		}
	}

	/** Searches the picture at a web address; the backend loads it. */
	async searchUrl(address: string) {
		const trimmed = address.trim();
		if (!/^https?:\/\//i.test(trimmed)) {
			this.errorText = 'Paste the address of a picture, starting with http.';
			return;
		}
		await this.run(trimmed, { url: trimmed });
	}

	/** Shows the picture and asks every face search about it. */
	private async run(pictureUrl: string, picture: FaceSearchPicture) {
		const token = ++this.requestToken;
		this.pictureUrl = pictureUrl;
		this.identification = null;
		this.errorText = null;
		this.loading = true;
		try {
			const identification = await identifyPicture(picture);
			if (token !== this.requestToken) return;
			this.identification = identification;
		} catch (error) {
			if (token !== this.requestToken) return;
			console.error('identifyPicture failed', error);
			this.errorText = error instanceof Error ? error.message : 'Could not search this picture.';
		} finally {
			if (token === this.requestToken) this.loading = false;
		}
	}
}

/**
 * The picture as a JPEG data URL, scaled down to MAX_PICTURE_SIDE: keeps
 * uploads small, and turns any format the browser can show into one every
 * face search accepts.
 */
async function scaledJpeg(file: File): Promise<string> {
	const bitmap = await createImageBitmap(file);
	const scale = Math.min(1, MAX_PICTURE_SIDE / Math.max(bitmap.width, bitmap.height));
	const canvas = document.createElement('canvas');
	canvas.width = Math.round(bitmap.width * scale);
	canvas.height = Math.round(bitmap.height * scale);
	const context = canvas.getContext('2d');
	if (!context) throw new Error('no canvas');
	context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
	bitmap.close();
	return canvas.toDataURL('image/jpeg', JPEG_QUALITY);
}

export const faceSearch = new FaceSearch();

/** Searches the sites for a name or title found by a face search, leaving the face search mode. */
export function searchSitesFor(text: string) {
	mediaFilter.set('all');
	searchQuery.set(text);
	goto(homeSearchUrl(text));
}
