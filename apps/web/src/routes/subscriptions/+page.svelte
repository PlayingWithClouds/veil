<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { untrack } from 'svelte';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import SubscriptionAvatar from '$lib/components/SubscriptionAvatar.svelte';
	import SubscriptionStrip from '$lib/components/subscriptions/SubscriptionStrip.svelte';
	import SubscriptionFeedGroups from '$lib/components/subscriptions/SubscriptionFeedGroups.svelte';
	import SubscriptionSettingsSheet from '$lib/components/subscriptions/SubscriptionSettingsSheet.svelte';
	import { subscriptionsUrl } from '$lib/routes';
	import { notifications } from '$lib/stores/notifications';
	import { refreshNewSubscriptionScenes } from '$lib/stores/newSubscriptionScenes';
	import {
		fetchSearchSubscriptions,
		fetchSubscriptionFeed,
		markSearchSubscriptionSeen,
		runSearchSubscription,
		scheduleLabel,
		subscriptionName,
		subscriptionTargetUrl,
		unsubscribeSearch,
		updateSearchSubscription,
		type SearchSubscription,
		type SubscriptionFeedEntry,
		type SubscriptionFeedFilter,
		type SubscriptionKind
	} from '$lib/searchSubscriptions';

	const FEED_PAGE_SIZE = 40;

	const KIND_LABELS: Record<SubscriptionKind, string> = {
		SEARCH: 'Search',
		STUDIO: 'Channel',
		PERFORMER: 'Performer',
		TAG: 'Tag'
	};

	// Kind filter chips, in this order, for the kinds the user follows.
	const KIND_FILTERS: { kind: SubscriptionKind; label: string }[] = [
		{ kind: 'STUDIO', label: 'Channels' },
		{ kind: 'PERFORMER', label: 'Performers' },
		{ kind: 'TAG', label: 'Tags' },
		{ kind: 'SEARCH', label: 'Searches' }
	];

	let subscriptions = $state<SearchSubscription[]>([]);
	let subscriptionsLoaded = $state(false);
	// Subscription ids with a run in flight.
	let runningIds = $state<Set<string>>(new Set());
	let settingsOpen = $state(false);

	// Feed filters; the kind filter applies to the combined feed only.
	let kindFilter = $state<SubscriptionKind | null>(null);
	let newOnly = $state(false);
	let unwatchedOnly = $state(false);

	let selectedId = $derived($page.url.searchParams.get('id'));
	let selectedSubscription = $derived(
		subscriptions.find((subscription) => subscription.id === selectedId)
	);
	let followedKinds = $derived(
		KIND_FILTERS.filter((filter) => subscriptions.some((subscription) => subscription.kind === filter.kind))
	);

	let feedEntries = $state<SubscriptionFeedEntry[]>([]);
	let feedLoading = $state(false);
	let feedDone = $state(false);
	let feedToken = 0;
	let sentinel = $state<HTMLElement | null>(null);

	$effect(() => {
		fetchSearchSubscriptions().then((found) => {
			subscriptions = found;
			subscriptionsLoaded = true;
		});
	});

	// Reload the feed whenever the view or a filter changes.
	$effect(() => {
		const filter = currentFilter();
		untrack(() => reloadFeed(filter));
	});

	// Scenes stay new while they are on screen; leaving the view marks them seen.
	$effect(() => {
		const viewedId = selectedId;
		return () => untrack(() => markViewedSeen(viewedId));
	});

	$effect(() => {
		if (!sentinel) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) loadMoreFeed();
			},
			{ rootMargin: '800px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	/** The feed filter for the open view and the chosen chips. */
	function currentFilter(): SubscriptionFeedFilter {
		const filter: SubscriptionFeedFilter = { newOnly, unwatchedOnly };
		if (selectedId) {
			filter.subscriptionId = selectedId;
			return filter;
		}
		if (kindFilter) filter.kinds = [kindFilter];
		return filter;
	}

	/** Starts the feed over from its first page. */
	function reloadFeed(filter: SubscriptionFeedFilter) {
		feedToken++;
		feedEntries = [];
		feedDone = false;
		feedLoading = false;
		loadMoreFeed(filter);
	}

	/** Appends the next page of the feed. */
	async function loadMoreFeed(filter: SubscriptionFeedFilter = currentFilter()) {
		if (feedLoading || feedDone) return;
		const token = feedToken;
		feedLoading = true;
		const page = await fetchSubscriptionFeed(filter, FEED_PAGE_SIZE, feedEntries.length);
		if (token !== feedToken) return;
		feedEntries = [...feedEntries, ...page];
		if (page.length < FEED_PAGE_SIZE) feedDone = true;
		feedLoading = false;
	}

	/** Clears the new counts of what the view showed: one subscription, or all of them. */
	function markViewedSeen(viewedId: string | null) {
		const viewed = subscriptions.filter((subscription) => {
			if (subscription.newCount === 0) return false;
			return viewedId === null || subscription.id === viewedId;
		});
		if (viewed.length === 0) return;
		Promise.all(viewed.map((subscription) => markSearchSubscriptionSeen(subscription.id)))
			.then((updated) => updated.forEach(replaceSubscription))
			.catch(() => {
				// The badges just stay until the next visit.
			})
			.finally(refreshNewSubscriptionScenes);
	}

	/** Swaps an updated subscription into the list. */
	function replaceSubscription(updated: SearchSubscription) {
		subscriptions = subscriptions.map((subscription) => {
			if (subscription.id === updated.id) return updated;
			return subscription;
		});
	}

	/** Opens a subscription's view, or the combined feed for null. */
	function select(id: string | null) {
		if (id === null) {
			goto(subscriptionsUrl(), { noScroll: true, keepFocus: true });
			return;
		}
		goto(subscriptionsUrl(id), { noScroll: true, keepFocus: true });
	}

	/** Toggles a kind chip; picking the active one shows every kind again. */
	function toggleKind(kind: SubscriptionKind) {
		if (kindFilter === kind) {
			kindFilter = null;
			return;
		}
		kindFilter = kind;
	}

	/** Changes how often a subscription runs. */
	async function changeInterval(subscription: SearchSubscription, intervalHours: number) {
		try {
			replaceSubscription(await updateSearchSubscription(subscription.id, { intervalHours }));
		} catch {
			notifications.push('Could not change the interval', 'error');
		}
	}

	/** Pauses or resumes a subscription's scheduled runs. */
	async function setEnabled(subscription: SearchSubscription, enabled: boolean) {
		try {
			replaceSubscription(await updateSearchSubscription(subscription.id, { enabled }));
		} catch {
			notifications.push('Could not update the subscription', 'error');
		}
	}

	/** Runs subscriptions now, then reloads the feed with what they found. */
	async function runNow(targets: SearchSubscription[]) {
		runningIds = new Set([...runningIds, ...targets.map((subscription) => subscription.id)]);
		await Promise.all(targets.map(runOne));
		refreshNewSubscriptionScenes();
		reloadFeed(currentFilter());
	}

	/** Runs one subscription, reporting a failure by name. */
	async function runOne(subscription: SearchSubscription) {
		try {
			replaceSubscription(await runSearchSubscription(subscription.id));
		} catch {
			notifications.push(`Checking "${subscriptionName(subscription)}" failed`, 'error');
		} finally {
			const remaining = new Set(runningIds);
			remaining.delete(subscription.id);
			runningIds = remaining;
		}
	}

	/** Deletes a subscription after confirmation. */
	async function remove(subscription: SearchSubscription) {
		if (!confirm(`Unsubscribe from "${subscriptionName(subscription)}"?`)) return;
		try {
			await unsubscribeSearch(subscription.id);
			subscriptions = subscriptions.filter((entry) => entry.id !== subscription.id);
			refreshNewSubscriptionScenes();
			if (subscription.id === selectedId) select(null);
		} catch {
			notifications.push('Could not unsubscribe', 'error');
		}
	}

	/** "Performer · 69 scenes", plus the searched sites for searches. */
	function subscriptionSummary(subscription: SearchSubscription): string {
		const summary = `${KIND_LABELS[subscription.kind]} · ${subscription.totalCount} scenes`;
		if (subscription.sources.length === 0) return summary;
		return `${summary} · ${subscription.sources.join(', ')}`;
	}

	/** What the empty feed says, depending on the filters. */
	function emptyFeedMessage(): string {
		if (newOnly) return 'Nothing new since your last visit.';
		if (unwatchedOnly) return 'You have watched everything here.';
		return 'No scenes yet. The first check may still be running.';
	}
</script>

<div class="flex flex-col gap-4">
	<PageHeader title="Subscriptions">
		{#snippet actions()}
			{#if subscriptions.length > 0}
				<button
					type="button"
					class="glass-pill flex h-9 items-center gap-2 rounded-full px-3.5 text-sm font-medium"
					disabled={runningIds.size > 0}
					onclick={() => runNow(subscriptions)}
				>
					<span class="flex" class:animate-spin={runningIds.size > 0}>
						<ArrowClockwiseIcon size={16} weight="bold" />
					</span>
					{#if runningIds.size > 0}Checking…{:else}Check all{/if}
				</button>
			{/if}
		{/snippet}
	</PageHeader>

	{#if subscriptionsLoaded && subscriptions.length === 0}
		<div class="flex flex-col items-center gap-3 py-24 text-center">
			<BellIcon size={48} class="text-base-content/20" />
			<p class="text-base-content/40">No subscriptions yet.</p>
			<p class="text-base-content/30 text-sm">
				Subscribe to a channel, performer, tag or search to collect new scenes here.
			</p>
		</div>
	{:else}
		<SubscriptionStrip {subscriptions} {selectedId} onselect={select} />

		{#if selectedSubscription}
			{@const targetUrl = subscriptionTargetUrl(selectedSubscription)}
			{@const running = runningIds.has(selectedSubscription.id)}
			<div class="flex items-center gap-3">
				<SubscriptionAvatar subscription={selectedSubscription} size={48} />
				<div class="flex min-w-0 flex-1 flex-col">
					{#if targetUrl}
						<a href={targetUrl} class="truncate text-lg font-semibold hover:underline">
							{subscriptionName(selectedSubscription)}
						</a>
					{:else}
						<span class="truncate text-lg font-semibold">{subscriptionName(selectedSubscription)}</span>
					{/if}
					<span class="text-base-content/50 truncate text-xs">{subscriptionSummary(selectedSubscription)}</span>
					<span class="text-base-content/40 truncate text-xs" class:text-error={selectedSubscription.lastError}>
						{#if selectedSubscription.lastError}
							Last check failed
						{:else}
							{scheduleLabel(selectedSubscription)}
						{/if}
					</span>
				</div>
				<button
					type="button"
					class="glass-pill flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
					aria-label="Check {subscriptionName(selectedSubscription)} now"
					disabled={running}
					onclick={() => runNow([selectedSubscription])}
				>
					<span class="flex" class:animate-spin={running}>
						<ArrowClockwiseIcon size={18} weight="bold" />
					</span>
				</button>
				<button
					type="button"
					class="glass-pill flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
					aria-label="Settings for {subscriptionName(selectedSubscription)}"
					onclick={() => (settingsOpen = true)}
				>
					<GearSixIcon size={18} weight="bold" />
				</button>
			</div>
		{/if}

		<div class="no-scrollbar fade-right-edge -mx-3 flex gap-2 overflow-x-auto px-3 pr-8 md:-mx-4 md:px-4">
			{#if !selectedId && followedKinds.length > 1}
				{#each followedKinds as filter (filter.kind)}
					<button
						type="button"
						class="filter-chip bg-base-200 flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium whitespace-nowrap"
						class:filter-chip-active={kindFilter === filter.kind}
						aria-pressed={kindFilter === filter.kind}
						onclick={() => toggleKind(filter.kind)}
					>
						{filter.label}
					</button>
				{/each}
				<span class="bg-base-content/10 my-1.5 w-px shrink-0" aria-hidden="true"></span>
			{/if}
			<button
				type="button"
				class="filter-chip bg-base-200 flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium whitespace-nowrap"
				class:filter-chip-active={newOnly}
				aria-pressed={newOnly}
				onclick={() => (newOnly = !newOnly)}
			>
				New
			</button>
			<button
				type="button"
				class="filter-chip bg-base-200 flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium whitespace-nowrap"
				class:filter-chip-active={unwatchedOnly}
				aria-pressed={unwatchedOnly}
				onclick={() => (unwatchedOnly = !unwatchedOnly)}
			>
				Unwatched
			</button>
		</div>

		{#if feedEntries.length === 0 && feedLoading}
			<div class="flex justify-center py-24">
				<span class="loading loading-spinner text-base-content/40"></span>
			</div>
		{:else if feedEntries.length === 0 && subscriptionsLoaded}
			<p class="text-base-content/40 py-24 text-center">{emptyFeedMessage()}</p>
		{:else}
			<SubscriptionFeedGroups entries={feedEntries} />
			{#if feedLoading}
				<div class="flex justify-center py-6">
					<span class="loading loading-spinner text-base-content/40"></span>
				</div>
			{/if}
			<div bind:this={sentinel} class="h-1"></div>
		{/if}
	{/if}
</div>

{#if settingsOpen && selectedSubscription}
	{@const subscription = selectedSubscription}
	<SubscriptionSettingsSheet
		{subscription}
		subtitle={subscriptionSummary(subscription)}
		onrun={() => runNow([subscription])}
		onremove={() => remove(subscription)}
		onintervalchange={(hours) => changeInterval(subscription, hours)}
		onenabledchange={(enabled) => setEnabled(subscription, enabled)}
		onclose={() => (settingsOpen = false)}
	/>
{/if}

<style>
	.filter-chip {
		transition: background-color 150ms;
	}

	.filter-chip-active {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
