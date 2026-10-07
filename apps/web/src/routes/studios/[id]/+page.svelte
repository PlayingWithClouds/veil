<!--
	Studio page: hero with logo, name, Subscribe/⋮ and stat tiles, then sticky
	Videos · Galleries · About tabs (swipe sideways to switch). Videos merges
	library scenes with ones found live on the sites.
-->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import DotsThreeIcon from 'phosphor-svelte/lib/DotsThreeIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import ShareNetworkIcon from 'phosphor-svelte/lib/ShareNetworkIcon';
	import FolderSimplePlusIcon from 'phosphor-svelte/lib/FolderSimplePlusIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import EntityHero, { type HeroStat } from '$lib/components/EntityHero.svelte';
	import EntityTabs from '$lib/components/EntityTabs.svelte';
	import EntitySceneFeed from '$lib/components/EntitySceneFeed.svelte';
	import EntityGalleryGrid from '$lib/components/EntityGalleryGrid.svelte';
	import ActionSheet, { type SheetAction } from '$lib/components/ActionSheet.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { cacheUrl } from '$lib/img';
	import { isCompact } from '$lib/stores/viewport';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import { notifications } from '$lib/stores/notifications';
	import { addBlock } from '$lib/blocklist';
	import { shareLink } from '$lib/share';
	import { swipeTabs } from '$lib/swipeTabs';
	import { studioUrl } from '$lib/routes';
	import type { StudioInfo } from '$lib/studioDetail';

	let { data } = $props();

	const TABS = [
		{ key: 'videos', label: 'Videos' },
		{ key: 'galleries', label: 'Galleries' },
		{ key: 'about', label: 'About' }
	];

	let studio = $derived(data.studio);
	let imageUrl = $derived(cacheUrl(studio?.imagePath));
	let activeTab = $state('videos');
	// Which way the tab content slides in: 1 = from the right.
	let tabDirection = $state(1);
	let actionsOpen = $state(false);
	let feed = $state<{ reload: () => Promise<void> } | null>(null);

	// Another studio (e.g. following the parent link) starts on Videos.
	$effect(() => {
		data.studio;
		activeTab = 'videos';
	});

	/** Switches tabs, sliding the content in from the side of the new tab. */
	function selectTab(key: string) {
		const from = TABS.findIndex((tab) => tab.key === activeTab);
		const to = TABS.findIndex((tab) => tab.key === key);
		if (to === -1 || to === from) return;
		tabDirection = Math.sign(to - from);
		activeTab = key;
	}

	/** Moves one tab left or right, if there is one. */
	function stepTab(step: number) {
		const index = TABS.findIndex((tab) => tab.key === activeTab);
		const next = TABS[index + step];
		if (next) selectTab(next.key);
	}

	/** A website's host for display; the raw text when it isn't a valid URL. */
	function hostOf(url: string): string {
		try {
			return new URL(url).host;
		} catch {
			return url;
		}
	}

	/** The hero's stat tiles; only filled-in ones. */
	function heroStats(info: StudioInfo): HeroStat[] {
		const stats: HeroStat[] = [];
		if (info.sceneCount > 0) stats.push({ label: 'Videos', value: String(info.sceneCount) });
		if (info.tags.length > 0) stats.push({ label: 'Tags', value: String(info.tags.length) });
		return stats;
	}

	/** The ⋮ sheet's actions. */
	function sheetActions(info: StudioInfo): SheetAction[] {
		return [
			{
				label: 'Add to collection…',
				icon: FolderSimplePlusIcon,
				run: () => openMoveDialog('studio', info.id, info.name)
			},
			{ label: 'Share', icon: ShareNetworkIcon, run: () => shareLink(info.name, info.url) },
			{ label: 'Block studio', icon: ProhibitIcon, run: () => block(info), destructive: true }
		];
	}

	/** Hides the studio's scenes everywhere. */
	async function block(info: StudioInfo) {
		try {
			await addBlock('studio', info.id, info.name);
			notifications.push(`Blocked ${info.name}`, 'success');
		} catch {
			notifications.push('Could not block', 'error');
		}
	}
</script>

{#if !studio}
	<div class="flex h-64 items-center justify-center">
		<span class="text-base-content/40">Studio not found</span>
	</div>
{:else}
	{#if $isCompact}
		<PullToRefresh onrefresh={async () => feed?.reload()} />
	{/if}
	<div class="flex flex-col gap-4">
		<EntityHero name={studio.name} {imageUrl} imageShape="logo" stats={heroStats(studio)}>
			{#snippet subtitle()}
				{#if studio?.parent}
					<a href={studioUrl(studio.parent.id)} class="text-base-content/60 text-sm">
						Part of <span class="text-base-content font-medium">{studio.parent.name}</span>
					</a>
				{/if}
			{/snippet}
			{#snippet actions()}
				{#if studio}
					<SubscribeButton kind="STUDIO" targetId={studio.id} targetName={studio.name} />
				{/if}
				<button
					type="button"
					onclick={() => (actionsOpen = true)}
					aria-label="More actions"
					class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
				>
					<DotsThreeIcon size={22} weight="bold" />
				</button>
			{/snippet}
		</EntityHero>

		<EntityTabs tabs={TABS} active={activeTab} onchange={selectTab} />

		<div
			class="grid min-h-[60vh]"
			use:swipeTabs={{ onnext: () => stepTab(1), onprevious: () => stepTab(-1) }}
		>
			{#key activeTab}
				<div
					class="col-start-1 row-start-1 min-w-0"
					in:fly={{ x: tabDirection * 40, duration: 260, easing: cubicOut }}
					out:fly={{ x: tabDirection * -40, duration: 160 }}
				>
					{#if activeTab === 'videos'}
						<EntitySceneFeed bind:this={feed} entityKind="studio" entityId={studio.id} name={studio.name} />
					{:else if activeTab === 'galleries'}
						<EntityGalleryGrid entityKind="studio" entityId={studio.id} />
					{:else}
						{@render about(studio)}
					{/if}
				</div>
			{/key}
		</div>
	</div>

	{#if actionsOpen}
		<ActionSheet
			title={studio.name}
			subtitle="Studio"
			{imageUrl}
			imageShape="avatar"
			actions={sheetActions(studio)}
			onclose={() => (actionsOpen = false)}
		/>
	{/if}
{/if}

{#snippet about(info: StudioInfo)}
	<div class="flex flex-col gap-6 pt-2">
		{#if info.details}
			<p class="text-base-content/80 text-[15px] leading-relaxed">{info.details}</p>
		{/if}
		{#if info.parent || info.url}
			<dl class="bg-base-200 grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 rounded-2xl p-4 text-sm">
				{#if info.parent}
					<dt class="text-base-content/55">Network</dt>
					<dd class="font-medium"><a href={studioUrl(info.parent.id)}>{info.parent.name}</a></dd>
				{/if}
				{#if info.url}
					<dt class="text-base-content/55">Website</dt>
					<dd class="truncate font-medium">
						<a href={info.url} target="_blank" rel="noreferrer">{hostOf(info.url)}</a>
					</dd>
				{/if}
			</dl>
		{/if}
		{#if info.aliases.length > 0}
			<div class="flex flex-col gap-2">
				<h3 class="text-base-content/50 text-xs font-semibold tracking-wider uppercase">Also known as</h3>
				<p class="text-sm">{info.aliases.join(', ')}</p>
			</div>
		{/if}
		{#if info.tags.length > 0}
			<div class="flex flex-col gap-2">
				<h3 class="text-base-content/50 text-xs font-semibold tracking-wider uppercase">Tags</h3>
				<TagChips tags={info.tags} />
			</div>
		{/if}
	</div>
{/snippet}
