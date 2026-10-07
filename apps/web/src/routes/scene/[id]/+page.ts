import { loadSceneDetail } from '$lib/sceneDetail';

/** Loads the scene and its stored playback sources for the watch page. */
export async function load({ params }) {
	const detail = await loadSceneDetail(params.id);
	return { ...detail, id: params.id };
}
