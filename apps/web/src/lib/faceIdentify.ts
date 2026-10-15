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

/** A tube-site video with a similar face. It names nobody and is only linked. */
export interface FaceVideoMatch {
	/** The video's own title. */
	title: string;
	/** 0..1, the best any service gave. */
	similarity: number;
	/** The video's page. */
	url: string;
	thumbnailUrl: string | null;
	/** Plugins that found this video. */
	sources: string[];
}

export interface FaceIdentification {
	candidates: FaceCandidate[];
	videos: FaceVideoMatch[];
	failures: string[];
	/** The stored picture that was sent; null when a frame was. */
	pictureUrl: string | null;
}

/**
 * Asks the face-recognition plugins who is in the picture of a scene (its
 * thumbnail), a gallery (imageId, else its first image) or an image, or in a
 * video frame (base64 JPEG) when one is given. Takes a few seconds: the sites
 * are queried live.
 */
export async function identifyFace(
	mediaId: string,
	imageId?: string | null,
	frameBase64?: string | null
): Promise<FaceIdentification> {
	const data = await gqlClient.mutation({
		identifyFace: {
			__args: { mediaId, imageId: imageId || null, frameBase64: frameBase64 || null },
			candidates: {
				name: true,
				similarity: true,
				imageUrl: true,
				profileUrls: true,
				sources: true,
				performerId: true
			},
			videos: { title: true, similarity: true, url: true, thumbnailUrl: true, sources: true },
			failures: true,
			pictureUrl: true
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
