import { writable } from 'svelte/store';
import { fetchSearchSubscriptions, type SearchSubscription } from '$lib/searchSubscriptions';

// Every subscription, for the sidebar list.
export const subscriptionList = writable<SearchSubscription[]>([]);

// Unseen scenes across all search subscriptions, shown as the nav badge.
export const newSubscriptionScenes = writable(0);

/** Reloads the subscriptions and their unseen-scene total from the backend. */
export async function refreshNewSubscriptionScenes(): Promise<void> {
	const subscriptions = await fetchSearchSubscriptions();
	subscriptionList.set(subscriptions);
	let total = 0;
	for (const subscription of subscriptions) total += subscription.newCount;
	newSubscriptionScenes.set(total);
}
