// Live cam rooms: the grid across cam plugins, followed models, and a room's
// live stream. A cam model is a room that comes and goes, not a scene.
import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';

export interface CamModel {
	id: string;
	pluginName: string;
	username: string;
	displayName: string;
	sourceUrl: string;
	thumbnailUrl: string | null;
	tags: string[];
	gender: string | null;
	subject: string | null;
	viewers: number;
	online: boolean;
	lastSeenOnlineAt: string | null;
	followed: boolean;
}

export interface CamSource {
	name: string;
	displayName: string;
	iconUrl: string | null;
}

export interface CamRoomFilter {
	sources: string[];
	tag: string;
	query: string;
	gender: string;
}

export interface CamRoomPage {
	rooms: CamModel[];
	totalCount: number;
}

export interface CamStream {
	url: string;
	mimeType: string;
}

const CAM_MODEL_FIELDS = {
	id: true,
	pluginName: true,
	username: true,
	displayName: true,
	sourceUrl: true,
	thumbnailUrl: true,
	tags: true,
	gender: true,
	subject: true,
	viewers: true,
	online: true,
	lastSeenOnlineAt: true,
	followed: true
} as const;

/** One page of the rooms online now, in the order the sites rank them. */
export async function fetchCamRooms(
	filter: CamRoomFilter,
	limit: number,
	offset: number
): Promise<CamRoomPage> {
	const data = await gqlClient.query({
		camRooms: {
			__args: {
				sources: filter.sources,
				tag: filter.tag || undefined,
				query: filter.query || undefined,
				gender: filter.gender || undefined,
				limit,
				offset
			},
			totalCount: true,
			rooms: CAM_MODEL_FIELDS
		}
	});
	return data.camRooms as unknown as CamRoomPage;
}

/** The cam sites installed, for the site filter. */
export async function fetchCamSources(): Promise<CamSource[]> {
	try {
		const data = await gqlClient.query({
			camSources: { name: true, displayName: true, iconUrl: true }
		});
		return data.camSources as unknown as CamSource[];
	} catch {
		return [];
	}
}

/** Followed models, online ones first. */
export async function fetchFollowedCamModels(): Promise<CamModel[]> {
	try {
		const data = await gqlClient.query({ followedCamModels: CAM_MODEL_FIELDS });
		return data.followedCamModels as unknown as CamModel[];
	} catch {
		return [];
	}
}

/** One model with its online status checked against the site; null when unknown. */
export async function fetchCamModel(id: string): Promise<CamModel | null> {
	try {
		const data = await gqlClient.query({ camModel: { __args: { id }, ...CAM_MODEL_FIELDS } });
		return (data.camModel ?? null) as unknown as CamModel | null;
	} catch {
		return null;
	}
}

/** The room's live HLS stream, through the backend proxy; throws when the room is offline. */
export async function fetchCamStream(id: string): Promise<CamStream> {
	const data = await gqlClient.query({ camStream: { __args: { id }, url: true, mimeType: true } });
	const stream = data.camStream as unknown as CamStream;
	return { url: backendUrl(stream.url), mimeType: stream.mimeType };
}

/** Follows or unfollows a model. */
export async function setCamModelFollowed(id: string, followed: boolean): Promise<CamModel> {
	if (followed) {
		const data = await gqlClient.mutation({
			followCamModel: { __args: { id }, ...CAM_MODEL_FIELDS }
		});
		return data.followCamModel as unknown as CamModel;
	}
	const data = await gqlClient.mutation({
		unfollowCamModel: { __args: { id }, ...CAM_MODEL_FIELDS }
	});
	return data.unfollowCamModel as unknown as CamModel;
}

/** "1.2K" style viewer counts for the card badge. */
export function formatViewers(viewers: number): string {
	if (viewers >= 10000) return `${Math.round(viewers / 1000)}K`;
	if (viewers >= 1000) return `${(viewers / 1000).toFixed(1)}K`;
	return String(viewers);
}
