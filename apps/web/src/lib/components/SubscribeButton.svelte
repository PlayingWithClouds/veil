<script lang="ts">
	// YouTube-style "Subscribe" / "Subscribed" pill for a studio (channel),
	// performer or tag. Following one collects its new scenes in the
	// subscriptions feed; clicking "Subscribed" unsubscribes after confirmation.
	import BellSimpleRingingIcon from 'phosphor-svelte/lib/BellSimpleRingingIcon';
	import { notifications } from '$lib/stores/notifications';
	import { refreshNewSubscriptionScenes } from '$lib/stores/newSubscriptionScenes';
	import {
		fetchSubscriptionForTarget,
		subscribeTarget,
		unsubscribeSearch,
		type FollowableKind,
		type SearchSubscription
	} from '$lib/searchSubscriptions';

	interface Props {
		kind: FollowableKind;
		targetId: string;
		// Shown in notifications and the unsubscribe confirmation.
		targetName: string;
	}

	let { kind, targetId, targetName }: Props = $props();

	let subscription = $state<SearchSubscription | null>(null);
	let loaded = $state(false);
	let busy = $state(false);

	$effect(() => {
		const id = targetId;
		loaded = false;
		subscription = null;
		fetchSubscriptionForTarget(id)
			.then((found) => {
				if (id !== targetId) return;
				subscription = found;
				loaded = true;
			})
			.catch(() => {
				loaded = true;
			});
	});

	/** Follows the target and refreshes the sidebar list. */
	async function subscribe() {
		busy = true;
		try {
			subscription = await subscribeTarget(kind, targetId);
			notifications.push(`Subscribed to ${targetName}`, 'success');
			refreshNewSubscriptionScenes();
		} catch {
			notifications.push(`Could not subscribe to ${targetName}`, 'error');
		} finally {
			busy = false;
		}
	}

	/** Stops following the target after confirmation. */
	async function unsubscribe(current: SearchSubscription) {
		if (!confirm(`Unsubscribe from ${targetName}?`)) return;
		busy = true;
		try {
			await unsubscribeSearch(current.id);
			subscription = null;
			refreshNewSubscriptionScenes();
		} catch {
			notifications.push('Could not unsubscribe', 'error');
		} finally {
			busy = false;
		}
	}
</script>

{#if subscription}
	{@const current = subscription}
	<button
		type="button"
		class="subscribe-pill bg-base-content/10 hover:bg-base-content/20 text-base-content"
		disabled={busy}
		title="Unsubscribe"
		onclick={() => unsubscribe(current)}
	>
		<BellSimpleRingingIcon size={16} weight="fill" />
		Subscribed
	</button>
{:else}
	<button
		type="button"
		class="subscribe-pill bg-base-content text-base-100 hover:opacity-90"
		disabled={busy || !loaded}
		onclick={subscribe}
	>
		Subscribe
	</button>
{/if}

<style>
	.subscribe-pill {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: 0.375rem;
		height: 2.25rem;
		padding-inline: 1rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 600;
		white-space: nowrap;
		transition:
			background-color 150ms,
			opacity 150ms;
	}

	.subscribe-pill:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
