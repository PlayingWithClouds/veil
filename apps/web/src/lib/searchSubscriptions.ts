import { gqlClient } from '$lib/veil';
import { performerUrl, studioUrl, tagUrl } from '$lib/routes';
import { formatRelativeTime } from '$lib/time';
import {
	browseSceneToResult,
	SCENE_CARD_FIELDS,
	type BrowseSceneData,
	type PluginSearchResult
} from '$lib/search';

// What a subscription follows: a saved search, or a studio (channel),
// performer or tag.
export type SubscriptionKind = 'SEARCH' | 'STUDIO' | 'PERFORMER' | 'TAG';

// Kinds that follow a record rather than a search query.
export type FollowableKind = Exclude<SubscriptionKind, 'SEARCH'>;

// The studio, performer or tag an entity subscription follows.
export interface SubscriptionTarget {
	id: string;
	name: string;
	// Studio logo / performer photo; null for tags.
	imageUrl: string | null;
}

// Subscriptions the backend re-runs on a schedule; each run's scenes collect in
// the subscription's feed.
export interface SearchSubscription {
	id: string;
	kind: SubscriptionKind;
	// Search query, or the followed target's name.
	query: string;
	// Null for searches.
	target: SubscriptionTarget | null;
	// Plugin names searched; empty = every search plugin (searches only).
	sources: string[];
	intervalHours: number;
	enabled: boolean;
	lastRunAt: string | null;
	nextRunAt: string | null;
	lastError: string | null;
	newCount: number;
	totalCount: number;
	createdAt: string;
}

// Re-run intervals offered in the subscription settings: hourly up to weekly.
export const INTERVAL_HOUR_OPTIONS = [1, 3, 6, 12, 24, 48, 168];

// Hours in a day and a week, for interval labels.
const HOURS_PER_DAY = 24;
const HOURS_PER_WEEK = 168;

/** Labels a re-run interval: "Every hour", "Every 6 hours", "Every day", "Every week". */
export function intervalLabel(hours: number): string {
	if (hours === 1) return 'Every hour';
	if (hours === HOURS_PER_DAY) return 'Every day';
	if (hours === HOURS_PER_WEEK) return 'Every week';
	if (hours % HOURS_PER_WEEK === 0) return `Every ${hours / HOURS_PER_WEEK} weeks`;
	if (hours % HOURS_PER_DAY === 0) return `Every ${hours / HOURS_PER_DAY} days`;
	return `Every ${hours} hours`;
}

/** Interval choices, including a non-standard current value. */
export function intervalOptions(current: number): number[] {
	if (INTERVAL_HOUR_OPTIONS.includes(current)) return INTERVAL_HOUR_OPTIONS;
	return [...INTERVAL_HOUR_OPTIONS, current].sort((a, b) => a - b);
}

/** Describes when a subscription last ran and runs next ("Ran 5 min ago · next in 6 hours"). */
export function scheduleLabel(subscription: SearchSubscription): string {
	let lastRun = 'Not run yet';
	if (subscription.lastRunAt) lastRun = `Ran ${formatRelativeTime(subscription.lastRunAt)}`;
	if (!subscription.enabled) return `${lastRun} · paused`;
	if (!subscription.nextRunAt) return lastRun;
	return `${lastRun} · next ${formatRelativeTime(subscription.nextRunAt)}`;
}

// One scene in the combined feed, with the subscription that found it.
export interface SubscriptionFeedEntry {
	scene: PluginSearchResult;
	// When a run first found the scene.
	foundAt: string;
	subscriptionId: string;
}

// Narrows the combined feed; unset fields match everything.
export interface SubscriptionFeedFilter {
	subscriptionId?: string;
	kinds?: SubscriptionKind[];
	newOnly?: boolean;
	unwatchedOnly?: boolean;
}

const SUBSCRIPTION_FIELDS = {
	id: true,
	kind: true,
	query: true,
	target: { id: true, name: true, imageUrl: true },
	sources: true,
	intervalHours: true,
	enabled: true,
	lastRunAt: true,
	nextRunAt: true,
	lastError: true,
	newCount: true,
	totalCount: true,
	createdAt: true
} as const;

/** Normalizes a query the way the backend matches subscriptions (lowercased, trimmed). */
export function normalizeSubscriptionQuery(query: string): string {
	return query.trim().toLowerCase();
}

/** Returns the search subscription whose query matches `query`, if any. */
export function findSubscriptionForQuery(
	subscriptions: SearchSubscription[],
	query: string
): SearchSubscription | undefined {
	const normalized = normalizeSubscriptionQuery(query);
	return subscriptions.find(
		(subscription) =>
			subscription.kind === 'SEARCH' && normalizeSubscriptionQuery(subscription.query) === normalized
	);
}

/** Link to the followed studio/performer/tag page; null for searches. */
export function subscriptionTargetUrl(subscription: SearchSubscription): string | null {
	if (!subscription.target) return null;
	if (subscription.kind === 'STUDIO') return studioUrl(subscription.target.id);
	if (subscription.kind === 'PERFORMER') return performerUrl(subscription.target.id);
	if (subscription.kind === 'TAG') return tagUrl(subscription.target.id);
	return null;
}

/** Display name: the followed target's current name, else the stored query. */
export function subscriptionName(subscription: SearchSubscription): string {
	if (subscription.target) return subscription.target.name;
	return subscription.query;
}

/** Loads the subscription following a studio, performer or tag; null when there is none. */
export async function fetchSubscriptionForTarget(
	targetId: string
): Promise<SearchSubscription | null> {
	const data = await gqlClient.query({
		subscriptionForTarget: { __args: { targetId }, ...SUBSCRIPTION_FIELDS }
	});
	if (!data.subscriptionForTarget) return null;
	return data.subscriptionForTarget as SearchSubscription;
}

/**
 * Follows a studio, performer or tag. The backend runs the first update in the
 * background and returns the existing subscription when already followed.
 */
export async function subscribeTarget(
	kind: FollowableKind,
	targetId: string
): Promise<SearchSubscription> {
	const data = await gqlClient.mutation({
		subscribe: { __args: { kind, targetId }, ...SUBSCRIPTION_FIELDS }
	});
	return data.subscribe as SearchSubscription;
}

/** Loads every search subscription. */
export async function fetchSearchSubscriptions(): Promise<SearchSubscription[]> {
	try {
		const data = await gqlClient.query({ searchSubscriptions: SUBSCRIPTION_FIELDS });
		return data.searchSubscriptions as SearchSubscription[];
	} catch {
		return [];
	}
}

/**
 * Subscribes to a search query. The backend runs the first search in the
 * background and returns the existing subscription when the query is already
 * subscribed. An empty sources list searches every plugin.
 */
export async function subscribeSearch(
	query: string,
	sources: string[]
): Promise<SearchSubscription> {
	const data = await gqlClient.mutation({
		subscribeSearch: { __args: { query, sources }, ...SUBSCRIPTION_FIELDS }
	});
	return data.subscribeSearch as SearchSubscription;
}

/** Changes a subscription's re-run interval and/or enabled flag. */
export async function updateSearchSubscription(
	id: string,
	changes: { intervalHours?: number; enabled?: boolean }
): Promise<SearchSubscription> {
	const data = await gqlClient.mutation({
		updateSearchSubscription: { __args: { id, ...changes }, ...SUBSCRIPTION_FIELDS }
	});
	return data.updateSearchSubscription as SearchSubscription;
}

/** Deletes a subscription and its feed. */
export async function unsubscribeSearch(id: string): Promise<void> {
	await gqlClient.mutation({ unsubscribeSearch: { __args: { id } } });
}

/** Runs a subscription's search now; resolves when the run finishes (can take a minute). */
export async function runSearchSubscription(id: string): Promise<SearchSubscription> {
	const data = await gqlClient.mutation({
		runSearchSubscription: { __args: { id }, ...SUBSCRIPTION_FIELDS }
	});
	return data.runSearchSubscription as SearchSubscription;
}

/** Clears a subscription's new-scene count. */
export async function markSearchSubscriptionSeen(id: string): Promise<SearchSubscription> {
	const data = await gqlClient.mutation({
		markSearchSubscriptionSeen: { __args: { id }, ...SUBSCRIPTION_FIELDS }
	});
	return data.markSearchSubscriptionSeen as SearchSubscription;
}

/** Loads a page of a subscription's feed as scene cards, newest first. */
export async function fetchSubscriptionScenes(
	id: string,
	limit: number,
	offset: number
): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			searchSubscription: {
				__args: { id },
				scenes: { __args: { limit, offset }, ...SCENE_CARD_FIELDS }
			}
		});
		if (!data.searchSubscription) return [];
		const scenes = data.searchSubscription.scenes as unknown as BrowseSceneData[];
		return scenes.map(browseSceneToResult);
	} catch {
		return [];
	}
}

/**
 * Loads a page of the scenes the subscriptions found, newest find first; a
 * scene several subscriptions found appears once. Scenes new since their
 * subscription was last marked seen carry isNew.
 */
export async function fetchSubscriptionFeed(
	filter: SubscriptionFeedFilter,
	limit: number,
	offset: number
): Promise<SubscriptionFeedEntry[]> {
	try {
		const data = await gqlClient.query({
			subscriptionFeed: {
				__args: { filter, limit, offset },
				foundAt: true,
				subscriptionId: true,
				isNew: true,
				scene: SCENE_CARD_FIELDS
			}
		});
		return data.subscriptionFeed.map((item) => ({
			scene: { ...browseSceneToResult(item.scene as unknown as BrowseSceneData), isNew: item.isNew },
			foundAt: item.foundAt,
			subscriptionId: item.subscriptionId
		}));
	} catch {
		return [];
	}
}

export { formatRelativeTime };
