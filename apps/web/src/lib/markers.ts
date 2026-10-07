import { gqlClient } from '$lib/veil';

// A marker describes what happens at a point (or span) in a scene. Personal
// markers are user-created; global ones are plugin-created annotations.
export interface SceneMarker {
	id: string;
	tag: { id: string; name: string } | null;
	label: string | null;
	seconds: number;
	endSeconds: number | null;
	personal: boolean;
}

const MARKER_FIELDS = {
	id: true,
	tag: { id: true, name: true },
	label: true,
	seconds: true,
	endSeconds: true,
	personal: true
} as const;

export function markerTitle(marker: SceneMarker): string {
	if (marker.tag && marker.label) return `${marker.tag.name} — ${marker.label}`;
	if (marker.tag) return marker.tag.name;
	return marker.label ?? '';
}

// Global markers plus the user's personal markers.
export async function fetchSceneMarkers(mediaId: string): Promise<SceneMarker[]> {
	const result = await gqlClient.query({
		sceneMarkers: { __args: { mediaId }, ...MARKER_FIELDS }
	});
	return (result.sceneMarkers ?? []) as unknown as SceneMarker[];
}

// Creates a personal marker. At least one of tagName/label must be set.
export async function createSceneMarker(
	mediaId: string,
	seconds: number,
	options: { tagName?: string; label?: string; endSeconds?: number }
): Promise<SceneMarker> {
	const result = await gqlClient.mutation({
		createSceneMarker: {
			__args: {
				mediaId,
				seconds,
				endSeconds: options.endSeconds,
				tagName: options.tagName || undefined,
				label: options.label || undefined
			},
			...MARKER_FIELDS
		}
	});
	return result.createSceneMarker as unknown as SceneMarker;
}

export async function deleteSceneMarker(markerId: string): Promise<void> {
	await gqlClient.mutation({ deleteSceneMarker: { __args: { markerId } } });
}
