<!--
	One performer: picture, facts, links, tags, and their videos and galleries.
	Opening the page reads their pages on the sites they came from (bio,
	picture, attributes) and asks the other sites to identify them, adding the
	profiles they're sure of; the result shows when it's in. Refresh reads and
	asks everything again. Edit changes any detail (those stay as set; each can be handed
	back to the sites). "Could be the same person" lists performers sharing a
	name or alias to merge with, and anyone else through the last card. A
	video's ⋮ "Not …" takes them off it; when a profile of theirs put them
	there, that profile moves away with its videos.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import dayjs from 'dayjs';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import UsersThreeIcon from 'phosphor-svelte/lib/UsersThreeIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import UserMinusIcon from 'phosphor-svelte/lib/UserMinusIcon';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import type { SheetAction } from '$lib/components/ActionSheet.svelte';
	import { setBlocked, setFavorite, logNotInterested } from '$lib/userState';
	import { showUndo } from '$lib/stores/undoToast';
	import TagChips from '$lib/components/TagChips.svelte';
	import RecordMedia from '$lib/components/RecordMedia.svelte';
	import PerformerEditor from '$lib/components/PerformerEditor.svelte';
	import PerformerMerge from '$lib/components/PerformerMerge.svelte';
	import type { MemberRow, VideoCardRow } from '$lib/library';
	import {
		fetchPerformer,
		unassignPerformer,
		readPerformer,
		fetchPerformerDuplicates,
		fetchPerformerFetching,
		type PerformerFetching,
		unlockPerformerField,
		type PerformerAttributes,
		type PerformerDetail
	} from '$lib/performers';
	import { siteDisplayName } from '$lib/plugins';
	import { capitalize } from '$lib/performerPresets';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl } from '$lib/routes';

	// Labels for the facts a performer page shows, in this order.
	const ATTRIBUTE_LABELS: [keyof PerformerAttributes, string][] = [
		['gender', 'Gender'],
		['birthdate', 'Born'],
		['deathDate', 'Died'],
		['country', 'Country'],
		['ethnicity', 'Ethnicity'],
		['hairColor', 'Hair'],
		['eyeColor', 'Eyes'],
		['heightCm', 'Height (cm)'],
		['weightKg', 'Weight (kg)'],
		['measurements', 'Measurements'],
		['fakeTits', 'Enhanced'],
		['tattoos', 'Tattoos'],
		['piercings', 'Piercings'],
		['careerStart', 'Active since'],
		['careerEnd', 'Active until']
	];

	// How the user's locked fields are named on the page.
	const FIELD_LABELS: Record<string, string> = {
		name: 'Name',
		aliases: 'Aliases',
		image: 'Picture',
		details: 'About',
		attributes: 'Facts',
		links: 'Links',
		tags: 'Tags'
	};

	let performer = $state<PerformerDetail | null>(null);
	let loading = $state(true);
	let reading = $state(false);
	let duplicates = $state<MemberRow[]>([]);
	let editing = $state(false);
	let merging = $state(false);
	let mergeWith = $state<MemberRow | null>(null);
	let menuAnchor = $state<HTMLElement | null>(null);
	let fetching = $state<PerformerFetching | null>(null);
	// Bumped to reload the videos as more come in.
	let mediaKey = $state(0);
	// Counts watches started, so a watch for an earlier page or favourite stops.
	let watchGeneration = 0;

	// Width of a card in the "Could be the same person" row (px).
	const DUPLICATE_WIDTH = 140;
	// How often the page asks what's being fetched while something is.
	const WATCH_INTERVAL_MS = 3000;

	let id = $derived(page.params.id ?? '');
	let facts = $derived(
		ATTRIBUTE_LABELS.flatMap(([key, label]) => {
			const value = performer?.attributes[key];
			if (value === null || value === undefined || value === '') return [];
			return [{ label, value: capitalize(String(value)) }];
		})
	);
	let lockedFields = $derived(
		(performer?.fieldSources ?? []).filter((source) => source.locked).map((source) => source.field)
	);

	$effect(() => {
		void load(id);
	});

	/** Shows the performer as stored, then reads their sites and shows that. */
	async function load(performerId: string) {
		loading = true;
		performer = null;
		duplicates = [];
		try {
			performer = await fetchPerformer(performerId);
		} finally {
			loading = false;
		}
		if (!performer) return;
		void loadDuplicates(performerId);
		await read(performerId, false);
	}

	/** Reads the performer's site pages (all of them with refresh) and shows the result. */
	async function read(performerId: string, refresh: boolean) {
		reading = true;
		try {
			const updated = await readPerformer(performerId, refresh);
			if (updated && updated.id === id) {
				performer = updated;
				void loadDuplicates(performerId);
				// Reading profiles queues listing their videos.
				void watchFetching(performerId);
			}
		} finally {
			reading = false;
		}
	}

	/** Loads the performers that probably are the same person. */
	async function loadDuplicates(performerId: string) {
		const found = await fetchPerformerDuplicates(performerId);
		if (performerId === id) duplicates = found;
	}

	/** Hands a field the user set back to the sites, and reads them again. */
	async function unlock(field: string) {
		if (!performer) return;
		await unlockPerformerField(performer.id, field);
		await read(performer.id, true);
	}

	/** Marks the performer a favourite (their sites then list all their videos), or not. */
	// Optimistic: the button flips at once, the backend follows.
	async function toggleFavorite() {
		if (!performer) return;
		const favorite = !performer.userState.favorite;
		performer.userState.favorite = favorite;
		await setFavorite('performer', performer.id, favorite);
		if (favorite) void watchFetching(performer.id);
	}

	/**
	 * Shows what's being fetched about the performer until it's done, then
	 * reloads their videos when more came in (not while scrolling through them).
	 */
	async function watchFetching(performerId: string) {
		const generation = ++watchGeneration;
		const startCount = performer?.videoCount ?? 0;
		for (;;) {
			const status = await fetchPerformerFetching(performerId);
			if (generation !== watchGeneration || performerId !== id) return;
			fetching = status;
			if (!status) return;
			if (!status.reading && status.listing.length === 0) {
				fetching = null;
				if (status.videoCount !== startCount) mediaKey++;
				return;
			}
			await new Promise((resolve) => setTimeout(resolve, WATCH_INTERVAL_MS));
		}
	}

	/** "xVideos, EPorner and 2 more": the sites still listing videos. */
	function listingSites(plugins: string[]): string {
		const names = plugins.map((plugin) => siteDisplayName(plugin, $sites));
		if (names.length <= 3) return names.join(', ');
		return `${names.slice(0, 3).join(', ')} and ${names.length - 3} more`;
	}

	/** The extra ⋮ entry on each of their videos: they aren't in it. */
	function videoActions(video: VideoCardRow, remove: () => void): SheetAction[] {
		if (!performer) return [];
		const name = performer.name;
		return [{ label: `Not ${name}`, icon: UserMinusIcon, run: () => void notInVideo(video, remove) }];
	}

	/**
	 * Takes the performer off a video. When a profile of theirs credited it,
	 * that profile was someone else's: it moved away with its videos, so the
	 * performer and their videos are shown again.
	 */
	async function notInVideo(video: VideoCardRow, remove: () => void) {
		if (!performer) return;
		const performerId = performer.id;
		const name = performer.name;
		remove();
		const other = await unassignPerformer(video.id, performerId);
		if (!other) {
			showUndo({ message: `Took ${name} off “${video.title}”` });
			return;
		}
		showUndo({
			message: `That site's ${name} is someone else: their profile and its videos moved to a new performer`
		});
		performer = await fetchPerformer(performerId);
		mediaKey++;
		void loadDuplicates(performerId);
	}

	/** The ⋮ menu: not interested, block. */
	function menuActions(): SheetAction[] {
		return [
			{ label: 'Not interested', icon: EyeSlashIcon, run: notInterested },
			{
				label: `Block ${performer?.name ?? ''}`,
				icon: ProhibitIcon,
				run: () => void setBlock(true),
				destructive: true
			}
		];
	}

	/**
	 * Tells the recommendations to leave the performer out; sent once Undo
	 * has passed, since the log can't take it back.
	 */
	function notInterested() {
		if (!performer) return;
		const performerId = performer.id;
		showUndo({
			message: `You won’t be recommended ${performer.name} again.`,
			settle: () => logNotInterested('performer', performerId)
		});
	}

	/** Blocks the performer (hiding them and their videos everywhere) or unblocks them, with Undo. */
	async function setBlock(blocked: boolean) {
		if (!performer) return;
		const target = performer;
		target.userState.blocked = blocked;
		await setBlocked('performer', target.id, blocked);
		if (!blocked) return;
		showUndo({
			message: `Blocked ${target.name}`,
			undo: () => {
				target.userState.blocked = false;
				void setBlocked('performer', target.id, false);
			}
		});
	}

	/** Opens the merge dialog, optionally starting from a suggested duplicate. */
	function startMerge(other: MemberRow | null) {
		mergeWith = other;
		merging = true;
	}

	/** Goes to the performer that stayed after a merge, or reloads this one. */
	function afterMerge(keptId: string) {
		merging = false;
		if (keptId === id) {
			void load(id);
			return;
		}
		goto(performerUrl(keptId));
	}

	/** "read 3 days ago", or that the page hasn't been read. */
	function readLabel(fetchedAt: string | null): string {
		if (!fetchedAt) return 'not read yet';
		return `read ${dayjs(fetchedAt).fromNow()}`;
	}
</script>

<svelte:head><title>{performer?.name ?? 'Performer'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6 pt-2">
	{#if performer}
		<div class="flex gap-5 max-md:flex-col">
			<div
				class="bg-base-200 flex aspect-[2/3] w-48 shrink-0 items-center justify-center overflow-hidden rounded-xl"
			>
				{#if performer.image}
					<img
						src={cacheUrl(performer.image.uri)}
						alt={performer.name}
						class="h-full w-full object-cover"
					/>
				{:else}
					<UserIcon size={48} class="opacity-15" />
				{/if}
			</div>
			<div class="flex min-w-0 flex-1 flex-col gap-3">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div>
						<h1 class="text-2xl font-bold tracking-tight">{performer.name}</h1>
						{#if performer.aliases.length > 0}
							<p class="text-base-content/50 text-sm">Also: {performer.aliases.join(', ')}</p>
						{/if}
					</div>
					<div class="flex gap-2">
						<button
							type="button"
							class="btn btn-sm"
							aria-pressed={performer.userState.favorite}
							title="Favourites get every video from their sites"
							onclick={toggleFavorite}
						>
							<HeartIcon
								size={16}
								weight={performer.userState.favorite ? 'fill' : 'regular'}
								class={performer.userState.favorite ? 'text-error' : ''}
							/>
							{performer.userState.favorite ? 'Favourite' : 'Add to favourites'}
						</button>
						<button
							type="button"
							class="btn btn-sm btn-square"
							aria-label="More"
							onclick={(event) => (menuAnchor = event.currentTarget)}
						>
							<DotsThreeVerticalIcon size={18} weight="bold" />
						</button>
						<button type="button" class="btn btn-sm" onclick={() => (editing = true)}>
							<PencilSimpleIcon size={16} /> Edit
						</button>
						<button
							type="button"
							class="btn btn-sm"
							disabled={reading}
							onclick={() => performer && read(performer.id, true)}
						>
							<ArrowClockwiseIcon size={16} class={reading ? 'animate-spin' : ''} />
							{reading ? 'Reading sites…' : 'Refresh from sites'}
						</button>
					</div>
				</div>

				{#if fetching}
					<p class="text-base-content/60 flex items-center gap-2 text-sm" aria-live="polite">
						<span class="loading loading-spinner loading-xs"></span>
						{#if fetching.reading}
							Reading their profiles and asking every site about them…
						{:else}
							Collecting videos from {listingSites(fetching.listing)} · {fetching.videoCount} so far
						{/if}
					</p>
				{/if}

				{#if performer.userState.blocked}
					<div class="bg-base-200 flex items-center gap-3 rounded-lg px-3 py-2 text-sm">
						<ProhibitIcon size={16} class="text-error" />
						<span class="flex-1">Blocked: {performer.name} and their videos are hidden everywhere.</span>
						<button type="button" class="btn btn-xs" onclick={() => setBlock(false)}>Unblock</button>
					</div>
				{/if}

				{#if lockedFields.length > 0}
					<div class="flex flex-wrap items-center gap-1.5 text-xs">
						<span class="text-base-content/50">Set by you, sites leave alone:</span>
						{#each lockedFields as field (field)}
							<button
								type="button"
								class="bg-base-200 hover:bg-base-300 flex items-center gap-1 rounded-full px-2 py-0.5"
								title="Let the sites update {FIELD_LABELS[field] ?? field} again"
								onclick={() => unlock(field)}
							>
								{FIELD_LABELS[field] ?? field}
								<XIcon size={11} />
							</button>
						{/each}
					</div>
				{/if}

				{#if facts.length > 0}
					<dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
						{#each facts as fact (fact.label)}
							<dt class="text-base-content/50">{fact.label}</dt>
							<dd>{fact.value}</dd>
						{/each}
					</dl>
				{/if}
				{#if performer.details}
					<p class="text-base-content/80 max-w-3xl text-sm whitespace-pre-line">
						{performer.details}
					</p>
				{/if}
				<TagChips tags={performer.tags} />
				{#if performer.links.length > 0}
					<div class="flex flex-wrap gap-x-4 gap-y-1 text-sm">
						<!-- Keyed by place: older records can list a link twice. -->
						{#each performer.links as link, position (position)}
							<a
								href={link}
								target="_blank"
								rel="noreferrer"
								class="text-base-content/70 hover:text-base-content flex items-center gap-1"
							>
								<ArrowSquareOutIcon size={13} />
								{new URL(link).hostname.replace(/^www\./, '')}
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		{#if performer.sources.length > 0}
			<section class="flex flex-col gap-2">
				<h2 class="text-base font-semibold">On the sites</h2>
				<div class="flex flex-wrap gap-2">
					<!-- Keyed by place: two sites, or two ids on one site, can share a URL. -->
					{#each performer.sources as source, position (position)}
						<a
							href={source.url}
							target="_blank"
							rel="noreferrer"
							class="bg-base-200 hover:bg-base-300 flex items-center gap-2 rounded-lg px-3 py-2 text-sm"
						>
							<ArrowSquareOutIcon size={14} />
							<span class="font-medium">{siteDisplayName(source.plugin, $sites)}</span>
							<span class="text-base-content/50 text-xs">{readLabel(source.fetchedAt)}</span>
						</a>
					{/each}
				</div>
			</section>
		{/if}

		<!-- The suggestions, then a card to merge with anyone else. -->
		<Shelf
			title="Could be the same person"
			items={[...duplicates, null]}
			key={(duplicate) => duplicate?.id ?? 'someone-else'}
			itemWidth={DUPLICATE_WIDTH}
		>
			{#snippet card(duplicate)}
				{#if duplicate}
					<div class="flex flex-col gap-2">
						<MemberCard kind="performer" member={duplicate} />
						<button type="button" class="btn btn-sm" onclick={() => startMerge(duplicate)}>
							<UsersThreeIcon size={16} /> Merge
						</button>
					</div>
				{:else}
					<button
						type="button"
						class="bg-base-200 hover:bg-base-300 text-base-content/60 flex aspect-[2/3] w-full flex-col items-center justify-center gap-2 rounded-xl text-sm"
						onclick={() => startMerge(null)}
					>
						<UsersThreeIcon size={28} />
						Someone else…
					</button>
				{/if}
			{/snippet}
		</Shelf>

		{#key mediaKey}
			<RecordMedia
				filter={{ performerId: performer.id }}
				allVideosHref={`/?${new URLSearchParams({ performer: performer.id })}`}
				{videoActions}
			/>
		{/key}

		{#if editing}
			<PerformerEditor
				{performer}
				onsaved={(saved) => (performer = saved)}
				onclose={() => (editing = false)}
			/>
		{/if}
		{#if menuAnchor}
			<CardMenu
				anchor={menuAnchor}
				title={performer.name}
				imageUrl={cacheUrl(performer.image?.uri)}
				actions={menuActions()}
				onclose={() => (menuAnchor = null)}
			/>
		{/if}
		{#if merging}
			<PerformerMerge
				{performer}
				other={mergeWith}
				onmerged={afterMerge}
				onclose={() => (merging = false)}
			/>
		{/if}
	{:else if !loading}
		<p class="text-base-content/50 text-sm">This performer isn't in the library.</p>
	{/if}
</div>
