<!--
	Performer page: hero with photo, name, Subscribe/favourite/⋮ and stat tiles,
	then sticky Videos · Galleries · About tabs (swipe sideways to switch). Videos
	merges library scenes with ones found live on the sites. Visiting also asks
	the backend to enrich the profile (photo, measurements, bio).
-->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
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
	import {
		setPerformerFavorite,
		enrichPerformer,
		loadPerformerDetail,
		type PerformerInfo
	} from '$lib/performerDetail';

	let { data } = $props();

	const TABS = [
		{ key: 'videos', label: 'Videos' },
		{ key: 'galleries', label: 'Galleries' },
		{ key: 'about', label: 'About' }
	];

	// A local copy so the enrich poll can fill the profile in after the load.
	let performer = $state<PerformerInfo | null>(data.performer);
	let favorite = $state(false);
	let activeTab = $state('videos');
	// Which way the tab content slides in: 1 = from the right.
	let tabDirection = $state(1);
	let actionsOpen = $state(false);
	let feed = $state<{ reload: () => Promise<void> } | null>(null);

	let imageUrl = $derived(cacheUrl(performer?.imagePath));

	// On visiting a performer: enrich their details. Runs once per performer id;
	// enrichedFor is a plain variable so updating it doesn't re-run the effect.
	let enrichedFor: string | null = null;
	$effect(() => {
		const current = data.performer;
		performer = current;
		favorite = current?.favorite ?? false;
		activeTab = 'videos';
		if (!current || enrichedFor === current.id) return;
		enrichedFor = current.id;
		enrichPerformer(current.id).catch(() => {});
		pollEnriched(current.id);
	});

	/** Reloads the profile until the enrich job lands a photo, or gives up. */
	async function pollEnriched(id: string) {
		for (let attempt = 0; attempt < 8; attempt++) {
			await new Promise((resolve) => setTimeout(resolve, 2500));
			if (enrichedFor !== id) return;
			const detail = await loadPerformerDetail(id);
			if (!detail.performer) continue;
			performer = detail.performer;
			if (detail.performer.imagePath) return;
		}
	}

	async function toggleFavorite() {
		if (!performer) return;
		favorite = !favorite;
		try {
			await setPerformerFavorite(performer.id, favorite);
		} catch {
			favorite = !favorite;
		}
	}

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

	/** Age in years from an ISO birthdate, or null when it can't be read. */
	function ageFrom(birthdate: string | null): number | null {
		if (!birthdate) return null;
		const born = new Date(birthdate);
		if (Number.isNaN(born.getTime())) return null;
		const now = new Date();
		let age = now.getFullYear() - born.getFullYear();
		const birthdayPassed =
			now.getMonth() > born.getMonth() ||
			(now.getMonth() === born.getMonth() && now.getDate() >= born.getDate());
		if (!birthdayPassed) age -= 1;
		return age;
	}

	/** The hero's stat tiles; only filled-in ones. */
	function heroStats(info: PerformerInfo): HeroStat[] {
		const stats: HeroStat[] = [];
		if (info.sceneCount > 0) stats.push({ label: 'Videos', value: String(info.sceneCount) });
		const age = ageFrom(info.birthdate);
		if (age !== null) stats.push({ label: 'Age', value: String(age) });
		if (info.heightCm) stats.push({ label: 'Height', value: `${info.heightCm} cm` });
		if (info.measurements) stats.push({ label: 'Figure', value: info.measurements });
		if (info.country) stats.push({ label: 'From', value: info.country });
		return stats;
	}

	/** The About tab's facts as label/value rows; only filled-in ones. */
	function facts(info: PerformerInfo): { label: string; value: string }[] {
		const rows: { label: string; value: string }[] = [];
		if (info.birthdate) rows.push({ label: 'Born', value: info.birthdate });
		if (info.country) rows.push({ label: 'Country', value: info.country });
		if (info.ethnicity) rows.push({ label: 'Ethnicity', value: info.ethnicity });
		if (info.heightCm) rows.push({ label: 'Height', value: `${info.heightCm} cm` });
		if (info.weightKg) rows.push({ label: 'Weight', value: `${info.weightKg} kg` });
		if (info.measurements) rows.push({ label: 'Measurements', value: info.measurements });
		if (info.hairColor) rows.push({ label: 'Hair', value: info.hairColor });
		if (info.eyeColor) rows.push({ label: 'Eyes', value: info.eyeColor });
		if (info.careerLength) rows.push({ label: 'Career', value: info.careerLength });
		return rows;
	}

	/** The ⋮ sheet's actions. */
	function sheetActions(info: PerformerInfo): SheetAction[] {
		return [
			{
				label: 'Add to collection…',
				icon: FolderSimplePlusIcon,
				run: () => openMoveDialog('performer', info.id, info.name)
			},
			{ label: 'Share', icon: ShareNetworkIcon, run: () => shareLink(info.name, info.url) },
			{ label: 'Block performer', icon: ProhibitIcon, run: () => block(info), destructive: true }
		];
	}

	/** Hides the performer's scenes everywhere. */
	async function block(info: PerformerInfo) {
		try {
			await addBlock('performer', info.id, info.name);
			notifications.push(`Blocked ${info.name}`, 'success');
		} catch {
			notifications.push('Could not block', 'error');
		}
	}
</script>

{#if !performer}
	<div class="flex h-64 items-center justify-center">
		<span class="text-base-content/40">Performer not found</span>
	</div>
{:else}
	{#if $isCompact}
		<PullToRefresh onrefresh={async () => feed?.reload()} />
	{/if}
	<div class="flex flex-col gap-4">
		<EntityHero name={performer.name} {imageUrl} imageShape="portrait" stats={heroStats(performer)}>
			{#snippet subtitle()}
				{#if performer && performer.aliases.length > 0}
					<p class="text-base-content/55 line-clamp-1 text-sm">a.k.a. {performer.aliases.join(', ')}</p>
				{/if}
			{/snippet}
			{#snippet actions()}
				{#if performer}
					<SubscribeButton kind="PERFORMER" targetId={performer.id} targetName={performer.name} />
				{/if}
				<button
					type="button"
					onclick={toggleFavorite}
					aria-label="Favorite"
					aria-pressed={favorite}
					class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
					class:text-error={favorite}
				>
					<HeartIcon size={20} weight={favorite ? 'fill' : 'regular'} />
				</button>
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
						<EntitySceneFeed bind:this={feed} entityKind="performer" entityId={performer.id} name={performer.name} />
					{:else if activeTab === 'galleries'}
						<EntityGalleryGrid entityKind="performer" entityId={performer.id} />
					{:else}
						{@render about(performer)}
					{/if}
				</div>
			{/key}
		</div>
	</div>

	{#if actionsOpen}
		<ActionSheet
			title={performer.name}
			subtitle="Performer"
			{imageUrl}
			imageShape="avatar"
			actions={sheetActions(performer)}
			onclose={() => (actionsOpen = false)}
		/>
	{/if}
{/if}

{#snippet about(info: PerformerInfo)}
	<div class="flex flex-col gap-6 pt-2">
		{#if info.details}
			<p class="text-base-content/80 text-[15px] leading-relaxed">{info.details}</p>
		{/if}
		{#if facts(info).length > 0}
			<dl class="bg-base-200 grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 rounded-2xl p-4 text-sm">
				{#each facts(info) as fact (fact.label)}
					<dt class="text-base-content/55">{fact.label}</dt>
					<dd class="font-medium">{fact.value}</dd>
				{/each}
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
