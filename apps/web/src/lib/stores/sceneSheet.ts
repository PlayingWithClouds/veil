// Drives the global SceneActionSheet (mounted once in the layout on phones): a
// long press or the ⋮ button on a scene card opens it for that scene.
import { writable } from 'svelte/store';

export interface SceneSheetTarget {
	sceneId: string;
	title: string;
	imageUrl: string | null;
	// Grey line under the title (who is in it, site).
	subtitle: string | null;
}

export const sceneSheetTarget = writable<SceneSheetTarget | null>(null);

// Scenes marked "Not interested" this session; their cards collapse to an undo row.
export const dismissedScenes = writable<Set<string>>(new Set());

/** Opens the action sheet for a scene. */
export function openSceneSheet(target: SceneSheetTarget): void {
	sceneSheetTarget.set(target);
}

/** Closes the action sheet. */
export function closeSceneSheet(): void {
	sceneSheetTarget.set(null);
}

/** Collapses or restores a scene's card after "Not interested" / undo. */
export function setSceneDismissed(sceneId: string, dismissed: boolean): void {
	dismissedScenes.update((current) => {
		const next = new Set(current);
		if (dismissed) {
			next.add(sceneId);
		} else {
			next.delete(sceneId);
		}
		return next;
	});
}
