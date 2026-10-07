<script lang="ts">
	// Subscribe/unsubscribe toggle for the current search query. A subscribed
	// query is re-run by the backend on a schedule and its results collect in the
	// subscriptions feed.
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellSimpleRingingIcon from 'phosphor-svelte/lib/BellSimpleRingingIcon';
	import { subscriptionsUrl } from '$lib/routes';
	import { notifications } from '$lib/stores/notifications';
	import { refreshNewSubscriptionScenes } from '$lib/stores/newSubscriptionScenes';
	import {
		fetchSearchSubscriptions,
		findSubscriptionForQuery,
		subscribeSearch,
		unsubscribeSearch,
		type SearchSubscription
	} from '$lib/searchSubscriptions';

	interface Props {
		query: string;
		// Plugin names to search; empty = every search plugin.
		sources: string[];
	}

	let { query, sources }: Props = $props();

	let subscriptions = $state<SearchSubscription[]>([]);
	let busy = $state(false);

	let currentSubscription = $derived(findSubscriptionForQuery(subscriptions, query));

	$effect(() => {
		fetchSearchSubscriptions().then((found) => {
			subscriptions = found;
		});
	});

	/** Subscribes to the current query. */
	async function subscribe() {
		busy = true;
		try {
			const created = await subscribeSearch(query.trim(), sources);
			subscriptions = [...subscriptions.filter((entry) => entry.id !== created.id), created];
			notifications.push(`Subscribed to "${created.query}"`, 'success');
		} catch {
			notifications.push('Could not subscribe to this search', 'error');
		} finally {
			busy = false;
		}
	}

	/** Removes the subscription for the current query. */
	async function unsubscribe(subscription: SearchSubscription) {
		busy = true;
		try {
			await unsubscribeSearch(subscription.id);
			subscriptions = subscriptions.filter((entry) => entry.id !== subscription.id);
			refreshNewSubscriptionScenes();
		} catch {
			notifications.push('Could not unsubscribe', 'error');
		} finally {
			busy = false;
		}
	}
</script>

{#if currentSubscription}
	{@const subscription = currentSubscription}
	<div class="flex items-center gap-1">
		<a href={subscriptionsUrl(subscription.id)} class="btn btn-sm text-primary">
			<BellSimpleRingingIcon size={14} weight="fill" />
			Subscribed
		</a>
		<button
			type="button"
			class="btn btn-sm btn-ghost"
			disabled={busy}
			onclick={() => unsubscribe(subscription)}
		>
			Unsubscribe
		</button>
	</div>
{:else}
	<button type="button" class="btn btn-sm btn-outline" disabled={busy} onclick={subscribe}>
		<BellIcon size={14} />
		Subscribe
	</button>
{/if}
