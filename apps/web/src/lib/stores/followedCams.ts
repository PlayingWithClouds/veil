// Followed cam models for the sidebar's Subscriptions list and the
// Subscriptions page; refreshed when a follow changes.
import { writable } from 'svelte/store';
import { fetchFollowedCamModels, type CamModel } from '$lib/cams';

export const followedCams = writable<CamModel[]>([]);

/** Reloads the followed cam models, online ones first. */
export async function refreshFollowedCams(): Promise<void> {
	followedCams.set(await fetchFollowedCamModels());
}
