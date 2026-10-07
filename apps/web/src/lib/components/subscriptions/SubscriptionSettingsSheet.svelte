<!--
	Settings for one subscription in an action sheet: run it now, open what it
	follows, unsubscribe, and when it runs on its own (on/off and how often).
-->
<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import { goto } from '$app/navigation';
	import ActionSheet, { type SheetAction } from '$lib/components/ActionSheet.svelte';
	import Toggle from '$lib/components/Toggle.svelte';
	import { cacheUrl } from '$lib/img';
	import {
		intervalLabel,
		intervalOptions,
		scheduleLabel,
		subscriptionName,
		subscriptionTargetUrl,
		type SearchSubscription
	} from '$lib/searchSubscriptions';

	interface Props {
		subscription: SearchSubscription;
		subtitle: string;
		onrun: () => void;
		onremove: () => void;
		onintervalchange: (hours: number) => void;
		onenabledchange: (enabled: boolean) => void;
		onclose: () => void;
	}

	let { subscription, subtitle, onrun, onremove, onintervalchange, onenabledchange, onclose }: Props =
		$props();

	/** Run now, the followed page when there is one, and unsubscribe. */
	function actions(): SheetAction[] {
		const list: SheetAction[] = [{ label: 'Check for new scenes now', icon: ArrowClockwiseIcon, run: onrun }];
		const targetUrl = subscriptionTargetUrl(subscription);
		if (targetUrl) {
			list.push({ label: `Open ${subscriptionName(subscription)}`, icon: ArrowSquareOutIcon, run: () => goto(targetUrl) });
		}
		list.push({ label: 'Unsubscribe', icon: TrashIcon, run: onremove, destructive: true });
		return list;
	}
</script>

<ActionSheet
	title={subscriptionName(subscription)}
	{subtitle}
	imageUrl={cacheUrl(subscription.target?.imageUrl)}
	imageShape="avatar"
	actions={actions()}
	{onclose}
>
	<div class="border-base-content/10 mt-1 flex flex-col gap-1 border-t px-3 pt-3 pb-2">
		<div class="flex items-center gap-3 py-1">
			<div class="flex min-w-0 flex-1 flex-col">
				<span class="text-[15px] font-medium">Check automatically</span>
				<span class="text-base-content/50 text-xs">{scheduleLabel(subscription)}</span>
			</div>
			<Toggle
				checked={subscription.enabled}
				ariaLabel="Check {subscriptionName(subscription)} automatically"
				onchange={onenabledchange}
			/>
		</div>

		{#if subscription.lastError}
			<p class="text-error line-clamp-2 text-xs" title={subscription.lastError}>
				Last check failed: {subscription.lastError}
			</p>
		{/if}

		<p class="text-base-content/50 pt-2 pb-1 text-xs font-semibold">How often</p>
		<div class="flex flex-wrap gap-2" class:opacity-40={!subscription.enabled}>
			{#each intervalOptions(subscription.intervalHours) as hours (hours)}
				{@const selected = hours === subscription.intervalHours}
				<button
					type="button"
					class="interval-chip bg-base-content/8 flex h-9 items-center gap-1.5 rounded-full px-3.5 text-sm font-medium"
					class:interval-chip-selected={selected}
					aria-pressed={selected}
					disabled={!subscription.enabled}
					onclick={() => onintervalchange(hours)}
				>
					{#if selected}
						<CheckIcon size={14} weight="bold" />
					{/if}
					{intervalLabel(hours)}
				</button>
			{/each}
		</div>
	</div>
</ActionSheet>

<style>
	.interval-chip {
		transition:
			background-color 150ms,
			color 150ms;
	}

	.interval-chip-selected {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
