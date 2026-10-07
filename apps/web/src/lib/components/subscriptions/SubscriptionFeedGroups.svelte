<!--
	The subscription feed in sections by when scenes were found: new since the
	last visit first, then today, yesterday, this week, this month, earlier.
-->
<script lang="ts">
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import type { PluginSearchResult } from '$lib/search';
	import type { SubscriptionFeedEntry } from '$lib/searchSubscriptions';

	interface Props {
		entries: SubscriptionFeedEntry[];
	}

	let { entries }: Props = $props();

	const DAY_MS = 24 * 60 * 60 * 1000;
	const GROUP_LABELS = ['New', 'Today', 'Yesterday', 'This week', 'This month', 'Earlier'] as const;
	type GroupLabel = (typeof GROUP_LABELS)[number];

	let groups = $derived(groupEntries(entries));

	/** Sorts the entries into the sections, skipping empty ones. */
	function groupEntries(list: SubscriptionFeedEntry[]): { label: GroupLabel; scenes: PluginSearchResult[] }[] {
		const byLabel = new Map<GroupLabel, PluginSearchResult[]>();
		const todayStart = startOfToday();
		for (const entry of list) {
			const label = groupLabel(entry, todayStart);
			const scenes = byLabel.get(label);
			if (scenes) {
				scenes.push(entry.scene);
			} else {
				byLabel.set(label, [entry.scene]);
			}
		}
		return GROUP_LABELS.filter((label) => byLabel.has(label)).map((label) => ({
			label,
			scenes: byLabel.get(label) as PluginSearchResult[]
		}));
	}

	/** Midnight today, local time, in milliseconds. */
	function startOfToday(): number {
		const now = new Date();
		return new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
	}

	/** The section an entry belongs to by whether it is new and when it was found. */
	function groupLabel(entry: SubscriptionFeedEntry, todayStart: number): GroupLabel {
		if (entry.scene.isNew) return 'New';
		const foundAt = Date.parse(entry.foundAt);
		if (foundAt >= todayStart) return 'Today';
		if (foundAt >= todayStart - DAY_MS) return 'Yesterday';
		if (foundAt >= todayStart - 6 * DAY_MS) return 'This week';
		if (foundAt >= todayStart - 29 * DAY_MS) return 'This month';
		return 'Earlier';
	}
</script>

<div class="flex flex-col gap-8">
	{#each groups as group (group.label)}
		<section class="flex flex-col gap-3" aria-label={group.label}>
			<h2 class="text-base font-semibold">{group.label}</h2>
			<NsfwGrid items={group.scenes} ariaLabel="{group.label} scenes" />
		</section>
	{/each}
</div>
