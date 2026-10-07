<!--
	A sideways-scrolling row of subscription avatars above the feed: "All"
	first, then every subscription with its name under it. New scenes show as
	a dot, paused subscriptions fade. Picking one opens its view.
-->
<script lang="ts">
	import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
	import SubscriptionAvatar from '$lib/components/SubscriptionAvatar.svelte';
	import { subscriptionName, type SearchSubscription } from '$lib/searchSubscriptions';

	interface Props {
		subscriptions: SearchSubscription[];
		// null = the combined feed.
		selectedId: string | null;
		onselect: (id: string | null) => void;
	}

	let { subscriptions, selectedId, onselect }: Props = $props();

	const AVATAR_SIZE = 56;
</script>

<ul
	class="no-scrollbar fade-right-edge -mx-3 flex gap-3 overflow-x-auto px-3 py-1 pr-8 md:-mx-4 md:px-4"
	aria-label="Subscriptions"
>
	<li class="shrink-0">
		<button
			type="button"
			class="strip-entry flex w-16 flex-col items-center gap-1.5"
			aria-current={selectedId === null}
			onclick={() => onselect(null)}
		>
			<span
				class="avatar-ring bg-base-200 flex items-center justify-center rounded-full"
				class:avatar-ring-selected={selectedId === null}
				style:width="{AVATAR_SIZE}px"
				style:height="{AVATAR_SIZE}px"
			>
				<SquaresFourIcon size={24} weight="fill" />
			</span>
			<span class="w-full truncate text-center text-xs font-medium">All</span>
		</button>
	</li>
	{#each subscriptions as subscription (subscription.id)}
		{@const selected = subscription.id === selectedId}
		<li class="shrink-0">
			<button
				type="button"
				class="strip-entry flex w-16 flex-col items-center gap-1.5"
				class:opacity-50={!subscription.enabled}
				aria-current={selected}
				onclick={() => onselect(subscription.id)}
			>
				<span class="avatar-ring relative rounded-full" class:avatar-ring-selected={selected}>
					<SubscriptionAvatar {subscription} size={AVATAR_SIZE} />
					{#if subscription.newCount > 0}
						<span
							class="bg-primary ring-base-100 absolute -top-0.5 -right-0.5 h-3.5 w-3.5 rounded-full ring-2"
							aria-label="{subscription.newCount} new"
						></span>
					{/if}
				</span>
				<span
					class="w-full truncate text-center text-xs"
					class:font-semibold={subscription.newCount > 0}
					class:text-base-content={selected || subscription.newCount > 0}
					class:text-base-content-muted={!selected && subscription.newCount === 0}
				>
					{subscriptionName(subscription)}
				</span>
			</button>
		</li>
	{/each}
</ul>

<style>
	.strip-entry {
		transition: transform 150ms;
	}

	.strip-entry:active {
		transform: scale(0.94);
	}

	/* A ring in the page colour, filled in with the text colour once selected. */
	.avatar-ring {
		box-shadow:
			0 0 0 2px var(--color-base-100),
			0 0 0 4px transparent;
		transition: box-shadow 150ms;
	}

	.avatar-ring-selected {
		box-shadow:
			0 0 0 2px var(--color-base-100),
			0 0 0 4px var(--color-base-content);
	}

	.text-base-content-muted {
		color: color-mix(in oklab, var(--color-base-content) 60%, transparent);
	}
</style>
