import { gqlClient } from '$lib/veil';
import type { PerformerRef } from '$lib/sceneDetail';

/** A performer the face-recognition services think the picture shows. */
export interface FaceCandidate {
	/** The performer's stage name. */
	name: string;
	/** 0..1, the best any service gave. */
	similarity: number;
	imageUrl: string | null;
	/** The performer's pages on adult sites. */
	profileUrls: string[];
	/** Plugins that named this candidate. */
	sources: string[];
	/** The library's performer of that name, when there is one. */
	performerId: string | null;
}

export interface FaceIdentification {
	candidates: FaceCandidate[];
	failures: string[];
}

/**
 * Asks the face-recognition plugins who is in the picture of a scene (its
 * thumbnail), a gallery (imageId, else its first image) or an image. Takes a
 * few seconds: the sites are queried live.
 */
export async function identifyFace(
	mediaId: string,
	imageId?: string | null
): Promise<FaceIdentification> {
	const data = await gqlClient.mutation({
		identifyFace: {
			__args: { mediaId, imageId: imageId || null },
			candidates: {
				name: true,
				similarity: true,
				imageUrl: true,
				profileUrls: true,
				sources: true,
				performerId: true
			},
			failures: true
		}
	});
	return data.identifyFace as unknown as FaceIdentification;
}

/** Credits the performer called name to a scene or gallery, creating the performer if the library has none. */
export async function creditPerformerByName(mediaId: string, name: string): Promise<PerformerRef> {
	const data = await gqlClient.mutation({
		creditPerformerByName: { __args: { mediaId, name }, id: true, name: true, imagePath: true }
	});
	return data.creditPerformerByName as unknown as PerformerRef;
}
