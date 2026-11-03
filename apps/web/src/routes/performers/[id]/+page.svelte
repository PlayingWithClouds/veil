<!--
	One performer: a compact profile header (picture lighting the shared
	glow, name, counts and a few facts, actions, two lines of bio, their
	profiles, links and tags as one chip row) so their videos and galleries
	start on the first screen. "All facts" opens every fact as tiles; one bar
	folds what's there to review (merges, face matches, fields the user set).
	Opening the page reads their pages on the sites they came from (bio,
	picture, attributes) and asks the other sites to identify them, adding the
	profiles they're sure of; the result shows when it's in. Refresh reads and
	asks everything again, or only the sites picked from its drop-down. Edit
	changes any detail (those stay as set; each can be handed
	back to the sites). "Could be the same person" lists performers sharing a
	name or alias: Merge folds one in at once (Undo in the toast), ⇄ compares
	first, the last card finds anyone else. "Merged in" lists merges that can
	still be undone. A
	video's ⋮ "Not …" takes them off it; when a profile of theirs put them
	there, that profile moves away with its videos.
-->
<script lang="ts">
	import { tick } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import dayjs from 'dayjs';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import UsersThreeIcon from 'phosphor-svelte/lib/UsersThreeIcon';
	import ArrowsLeftRightIcon from 'phosphor-svelte/lib/ArrowsLeftRightIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import UserMinusIcon from 'phosphor-svelte/lib/UserMinusIcon';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import PictureDialog from '$lib/components/PictureDialog.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import type { SheetAction } from '$lib/components/ActionSheet.svelte';
	import { setBlocked, setFavorite, logNotInterested } from '$lib/userState';
	import { showUndo } from '$lib/stores/undoToast';
	import { tagLabel } from '$lib/components/TagChips.svelte';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import RecordMedia from '$lib/components/RecordMedia.svelte';
	import PerformerEditor from '$lib/components/PerformerEditor.svelte';
	import PerformerMerge from '$lib/components/PerformerMerge.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import {
		fetchVideoDetail,
		setVideoPerformers,
		type MemberRow,
		type VideoCardRow
	} from '$lib/library';
	import {
		fetchPerformer,
		unassignPerformer,
		editPerformer,
		readPerformer,
		fetchPerformerDuplicates,
		fetchPerformerFetching,
		fetchMerges,
		mergePerformers,
		undoMerge,
		type MergeRow,
		type PerformerFetching,
		unlockPerformerField,
		type PerformerAttributes,
		type PerformerDetail
	} from '$lib/performers';
	import { siteDisplayName, type SitePlugin } from '$lib/plugins';
	import RefreshSitesMenu, { type RefreshSite } from '$lib/components/RefreshSitesMenu.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import { capitalize } from '$lib/performerPresets';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, tagFilterUrl, videoUrl } from '$lib/routes';
	import {
		FACE_MODEL,
		fetchFaceMatches,
		learnFaces,
		trackPerformer,
		type FaceMatchRow
	} from '$lib/faces';

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
	let merges = $state<MergeRow[]>([]);
	let menuAnchor = $state<HTMLElement | null>(null);
	let choosingPicture = $state(false);
	let fetching = $state<PerformerFetching | null>(null);
	// Uncredited videos their face was found in, minus the ones waved away here.
	let faceMatches = $state<FaceMatchRow[]>([]);
	let dismissedMatches = $state<string[]>([]);
	// Bumped to reload the videos as more come in.
	let mediaKey = $state(0);
	// Counts watches started, so a watch for an earlier page or favourite stops.
	let watchGeneration = 0;

	// Width of a card in the "Could be the same person" row (px).
	const DUPLICATE_WIDTH = 140;
	// Facts named in the header line; "All facts" shows every one.
	const HEADLINE_FACTS = 5;
	// Width of a card in the "Probably also in" row (px).
	const FACE_MATCH_WIDTH = 260;
	// How often the page asks what's being fetched while something is.
	const WATCH_INTERVAL_MS = 3000;

	let id = $derived(page.params.id ?? '');
	let shownFaceMatches = $derived(
		faceMatches.filter((match) => !dismissedMatches.includes(match.video.id))
	);
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
	let refreshSites = $derived(refreshSitesOf(performer, $sites));
	// The header's folded parts, opened on demand.
	let allFacts = $state(false);
	let fullBio = $state(false);
	let reviewing = $state(false);
	// The few facts the header names in one line.
	let headlineFacts = $derived(facts.slice(0, HEADLINE_FACTS));
	let reviewSummary = $derived(
		reviewLine(duplicates.length, shownFaceMatches.length, merges.length)
	);

	// A link to the face matches (from a video's notice) opens them and scrolls there.
	$effect(() => {
		if (page.url.hash !== '#face-matches' || shownFaceMatches.length === 0) return;
		reviewing = true;
		void tick().then(() =>
			document.getElementById('face-matches')?.scrollIntoView({ behavior: 'smooth' })
		);
	});
	let picture = $derived(cacheUrl(performer?.image?.uri));

	showAmbientGlow(() => picture);

	$effect(() => {
		void load(id);
	});

	/** Shows the performer as stored, then reads their sites and shows that. */
	async function load(performerId: string) {
		loading = true;
		performer = null;
		duplicates = [];
		merges = [];
		faceMatches = [];
		dismissedMatches = [];
		allFacts = false;
		fullBio = false;
		reviewing = false;
		try {
			performer = await fetchPerformer(performerId);
		} finally {
			loading = false;
		}
		if (!performer) return;
		void loadDuplicates(performerId);
		void loadMerges(performerId);
		void loadFaceMatches(performerId);
		void learnFromPictures(performerId);
		await read(performerId, false);
	}

	/**
	 * Learns their face from their unscanned pictures, for the face matches
	 * on videos, and lists the videos it's then found in.
	 */
	async function learnFromPictures(performerId: string) {
		const outcome = await trackPerformer(performerId);
		if (outcome === 'scanned') await loadFaceMatches(performerId);
	}

	/** Loads the uncredited videos their face was probably found in. */
	async function loadFaceMatches(performerId: string) {
		const found = await fetchFaceMatches(performerId);
		if (performerId === id) faceMatches = found;
	}

	/**
	 * Credits them on a video their face was found in, then learns their
	 * face from it, which can find them in more videos.
	 */
	async function creditFaceMatch(match: FaceMatchRow) {
		if (!performer) return;
		const performerId = performer.id;
		dismissedMatches = [...dismissedMatches, match.video.id];
		const shown = await fetchVideoDetail(match.video.id);
		if (!shown) return;
		const ids = [...shown.performers.map((credited) => credited.id), performerId];
		await setVideoPerformers(shown.id, ids);
		mediaKey++;
		const learned = await learnFaces(shown.id, FACE_MODEL);
		if (learned > 0) await loadFaceMatches(performerId);
	}

	/**
	 * Reads the performer's site pages (all of them with refresh, or only
	 * plugins') and shows the result.
	 */
	async function read(performerId: string, refresh: boolean, plugins: string[] = []) {
		reading = true;
		try {
			const updated = await readPerformer(performerId, refresh, plugins);
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

	/** Saves a chosen picture as the user's (the sites then leave it). */
	async function setPicture(uri: string) {
		choosingPicture = false;
		if (!performer) return;
		const saved = await editPerformer(performer.id, { imageUrl: uri });
		if (saved) performer = saved;
	}

	/** The extra ⋮ entry on each of their videos: they aren't in it. */
	function videoActions(video: VideoCardRow, remove: () => void): SheetAction[] {
		if (!performer) return [];
		const name = performer.name;
		return [
			{ label: `Not ${name}`, icon: UserMinusIcon, run: () => void notInVideo(video, remove) }
		];
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

	/** Merges a suggested duplicate into this performer at once, with Undo. */
	async function mergeNow(duplicate: MemberRow) {
		if (!performer) return;
		const kept = performer;
		duplicates = duplicates.filter((candidate) => candidate.id !== duplicate.id);
		const mergeId = await mergePerformers(kept.id, duplicate.id);
		afterMerge(kept, duplicate, mergeId);
	}

	/**
	 * Goes to the performer that stayed after a merge, or reloads this one,
	 * and offers Undo.
	 */
	function afterMerge(kept: MemberRow, merged: MemberRow, mergeId: string | null) {
		merging = false;
		const stayed = kept.id === id;
		if (stayed) {
			void load(id);
			mediaKey++;
		} else {
			goto(performerUrl(kept.id));
		}
		if (!mergeId) return;
		let returnTo: string | null = null;
		if (!stayed) returnTo = merged.id;
		showUndo({
			message: `Merged ${merged.name} into ${kept.name}`,
			undo: () => void unmerge(mergeId, returnTo)
		});
	}

	/**
	 * Undoes a merge, then goes back to the performer whose page it started
	 * on (returnTo), or reloads this one.
	 */
	async function unmerge(mergeId: string, returnTo: string | null) {
		await undoMerge(mergeId);
		if (returnTo) {
			goto(performerUrl(returnTo));
			return;
		}
		void load(id);
		mediaKey++;
	}

	/** Loads the merges into the performer that can still be undone. */
	async function loadMerges(performerId: string) {
		const found = await fetchMerges(performerId);
		if (performerId === id) merges = found;
	}

	/** "read 3 days ago", or that the page hasn't been read. */
	function readLabel(fetchedAt: string | null): string {
		if (!fetchedAt) return 'not read yet';
		return `read ${dayjs(fetchedAt).fromNow()}`;
	}

	/**
	 * The sites a refresh can ask: those with a profile of theirs (read
	 * again), then the enabled ones that look performers up.
	 */
	function refreshSitesOf(current: PerformerDetail | null, installed: SitePlugin[]): RefreshSite[] {
		if (!current) return [];
		const choices: RefreshSite[] = [];
		for (const source of current.sources) {
			if (choices.some((choice) => choice.name === source.plugin)) continue;
			const plugin = installed.find((candidate) => candidate.name === source.plugin);
			choices.push({
				name: source.plugin,
				label: siteDisplayName(source.plugin, installed),
				iconUrl: plugin?.iconUrl ?? null,
				detail: `Profile ${readLabel(source.fetchedAt)}`
			});
		}
		for (const plugin of installed) {
			const identifies = plugin.capabilities.includes('performer:identify');
			if (!identifies || !plugin.enabled || !plugin.available) continue;
			if (choices.some((choice) => choice.name === plugin.name)) continue;
			choices.push({
				name: plugin.name,
				label: siteDisplayName(plugin.name, installed),
				iconUrl: plugin.iconUrl,
				detail: 'Looks them up'
			});
		}
		return choices;
	}

	/** "2 could be the same person · 4 videos probably with them", or null when nothing waits. */
	function reviewLine(
		duplicateCount: number,
		matchCount: number,
		mergeCount: number
	): string | null {
		const parts: string[] = [];
		if (duplicateCount > 0) parts.push(`${duplicateCount} could be the same person`);
		if (matchCount > 0) {
			parts.push(`${matchCount} ${matchCount === 1 ? 'video' : 'videos'} probably with them`);
		}
		if (mergeCount > 0) parts.push(`${mergeCount} merged in`);
		if (parts.length === 0) return null;
		return parts.join(' · ');
	}

	/** A steady colour per name, for the initial of a performer without a picture. */
	function hueOf(name: string): number {
		let hash = 0;
		for (const character of name) hash = (hash * 31 + character.charCodeAt(0)) % 360;
		return hash;
	}

	/** Refreshes from the given sites, or every site when none. */
	function refreshFrom(plugins: string[]) {
		if (!performer) return;
		void read(performer.id, true, plugins);
	}
</script>

<svelte:head><title>{performer?.name ?? 'Performer'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6 pt-2">
	{#if performer}
		<header class="flex gap-6 max-md:flex-col max-md:items-center max-md:text-center">
			<button
				type="button"
				class="group relative flex aspect-[3/4] w-36 shrink-0 items-center justify-center self-start overflow-hidden rounded-3xl shadow-2xl ring-1 shadow-black/50 ring-white/10 max-md:self-center"
				style:background-color={picture ? null : `hsl(${hueOf(performer.name)} 45% 62%)`}
				aria-label="Change the picture"
				onclick={() => (choosingPicture = true)}
			>
				{#if picture}
					<img
						src={picture}
						alt={performer.name}
						class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
					/>
				{:else}
					<span class="text-5xl font-bold text-black/60">
						{performer.name.trim().charAt(0).toUpperCase()}
					</span>
				{/if}
				<span
					class="glass-pill absolute bottom-3 left-1/2 -translate-x-1/2 rounded-full px-3 py-1 text-xs whitespace-nowrap opacity-0 transition-opacity group-hover:opacity-100"
				>
					Change picture
				</span>
			</button>

			<div class="flex min-w-0 flex-1 flex-col gap-3">
				<div class="flex flex-col gap-0.5">
					<h1 class="text-3xl font-bold tracking-tight">{performer.name}</h1>
					{#if performer.aliases.length > 0}
						<p class="text-base-content/50 text-sm">Also known as {performer.aliases.join(', ')}</p>
					{/if}
					<p class="text-base-content/60 text-sm">
						{performer.videoCount}
						{performer.videoCount === 1 ? 'video' : 'videos'}
						{#if performer.galleryCount > 0}
							· {performer.galleryCount}
							{performer.galleryCount === 1 ? 'gallery' : 'galleries'}
						{/if}
						{#each headlineFacts as fact (fact.label)}
							<span class="text-base-content/45"> · {fact.label}</span>
							{fact.value}
						{/each}
						{#if facts.length > 0}
							<button
								type="button"
								class="text-info ml-1 text-xs font-semibold"
								aria-expanded={allFacts}
								onclick={() => (allFacts = !allFacts)}
							>
								{allFacts ? 'Fewer facts' : 'All facts'}
							</button>
						{/if}
					</p>
				</div>

				<div class="flex flex-wrap gap-2 max-md:justify-center">
					<button
						type="button"
						class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
						aria-pressed={performer.userState.favorite}
						title="Favourites get every video from their sites"
						onclick={toggleFavorite}
					>
						<span class="flex" class:text-error={performer.userState.favorite}>
							<HeartIcon size={18} weight={performer.userState.favorite ? 'fill' : 'regular'} />
						</span>
						{performer.userState.favorite ? 'Favourite' : 'Add to favourites'}
					</button>
					<button
						type="button"
						class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
						onclick={() => (editing = true)}
					>
						<PencilSimpleIcon size={16} /> Edit
					</button>
					<RefreshSitesMenu sites={refreshSites} busy={reading} onrefresh={refreshFrom} />
					<button
						type="button"
						class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
						aria-label="More"
						onclick={(event) => (menuAnchor = event.currentTarget)}
					>
						<DotsThreeVerticalIcon size={20} weight="bold" />
					</button>
				</div>

				{#if fetching}
					<span class="chip pointer-events-none self-start" aria-live="polite">
						<span class="loading loading-spinner loading-xs"></span>
						{#if fetching.reading}
							Reading their profiles and asking every site about them…
						{:else}
							Collecting videos from {listingSites(fetching.listing)} · {fetching.videoCount} so far
						{/if}
					</span>
				{/if}

				{#if performer.userState.blocked}
					<div
						class="flex items-center gap-3 rounded-2xl bg-[var(--ctx-red-soft)] px-4 py-2.5 text-sm ring-1 ring-red-400/25"
					>
						<ProhibitIcon size={16} class="text-error" />
						<span class="flex-1">
							Blocked: {performer.name} and their videos are hidden everywhere.
						</span>
						<button type="button" class="chip chip-sm" onclick={() => setBlock(false)}>
							Unblock
						</button>
					</div>
				{/if}

				{#if performer.details}
					<div class="flex max-w-3xl flex-col items-start gap-1">
						<p
							class="text-base-content/75 text-sm leading-relaxed whitespace-pre-line"
							class:line-clamp-2={!fullBio}
						>
							{performer.details}
						</p>
						{#if performer.details.length > 180}
							<button
								type="button"
								class="text-info text-xs font-semibold"
								onclick={() => (fullBio = !fullBio)}
							>
								{fullBio ? 'Show less' : 'Show more'}
							</button>
						{/if}
					</div>
				{/if}

				<div
					class="no-scrollbar fade-right-edge -mx-1 flex items-center gap-2 overflow-x-auto px-1 py-0.5"
				>
					<!-- Keyed by place: two sites, or two ids on one site, can share a URL. -->
					{#each performer.sources as source, position (position)}
						<a
							href={source.url}
							target="_blank"
							rel="noreferrer"
							class="chip chip-sm"
							title="{siteDisplayName(source.plugin, $sites)}, {readLabel(source.fetchedAt)}"
						>
							{siteDisplayName(source.plugin, $sites)}
							<ArrowSquareOutIcon size={11} class="opacity-60" />
						</a>
					{/each}
					<!-- Keyed by place: older records can list a link twice. -->
					{#each performer.links as link, position (position)}
						<a href={link} target="_blank" rel="noreferrer" class="chip chip-sm">
							{new URL(link).hostname.replace(/^www\./, '')}
							<ArrowSquareOutIcon size={11} class="opacity-60" />
						</a>
					{/each}
					{#if performer.tags.length > 0 && (performer.sources.length > 0 || performer.links.length > 0)}
						<span class="h-4 w-px shrink-0 bg-white/15"></span>
					{/if}
					{#each performer.tags as tag (tag.id)}
						<a href={tagFilterUrl(tag.id)} class="chip chip-sm" title={tag.name}>
							{tagLabel(tag.name)}
						</a>
					{/each}
				</div>
			</div>
		</header>

		{#if allFacts && facts.length > 0}
			<dl class="rise-in grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-2">
				{#each facts as fact (fact.label)}
					<div
						class="flex flex-col gap-0.5 rounded-2xl bg-white/[0.04] px-4 py-3 ring-1 ring-white/5"
					>
						<dt class="text-base-content/45 text-xs">{fact.label}</dt>
						<dd class="truncate text-sm font-medium" title={fact.value}>{fact.value}</dd>
					</div>
				{/each}
			</dl>
		{/if}

		<div class="flex flex-col gap-5 rounded-2xl bg-white/[0.03] ring-1 ring-white/5">
			<button
				type="button"
				class="flex w-full items-center gap-3 rounded-2xl px-4 py-3 text-left text-sm transition-colors hover:bg-white/[0.03]"
				aria-expanded={reviewing}
				onclick={() => (reviewing = !reviewing)}
			>
				<UsersThreeIcon size={18} class="shrink-0 opacity-60" />
				<span class="min-w-0 flex-1 truncate">
					{#if reviewSummary}
						<span class="font-medium">To review:</span>
						<span class="text-base-content/65">{reviewSummary}</span>
					{:else}
						<span class="text-base-content/65">Merges, face matches and what you've set</span>
					{/if}
				</span>
				<span class="flex transition-transform" class:rotate-180={reviewing}>
					<CaretDownIcon size={14} weight="bold" />
				</span>
			</button>

			{#if reviewing}
				<div class="flex flex-col gap-6 px-4 pb-4">
					{#if lockedFields.length > 0}
						<div class="flex flex-wrap items-center gap-2 text-xs">
							<span class="text-base-content/45">Set by you, the sites leave these alone:</span>
							{#each lockedFields as field (field)}
								<button
									type="button"
									class="chip chip-sm"
									title="Let the sites update {FIELD_LABELS[field] ?? field} again"
									onclick={() => unlock(field)}
								>
									{FIELD_LABELS[field] ?? field}
									<XIcon size={11} />
								</button>
							{/each}
						</div>
					{/if}

					{#if merges.length > 0}
						<section class="flex flex-col gap-3">
							<h2 class="text-lg font-semibold">Merged in</h2>
							<div class="flex flex-wrap gap-2">
								{#each merges as merge (merge.id)}
									<span class="chip">
										<span class="font-medium">{merge.mergedName ?? 'A performer'}</span>
										<span class="text-xs opacity-50">{dayjs(merge.createdAt).fromNow()}</span>
										<button
											type="button"
											class="text-info text-xs font-semibold"
											onclick={() => void unmerge(merge.id, null)}
										>
											Undo
										</button>
									</span>
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
									<span class="flex gap-1">
										<button
											type="button"
											class="chip chip-sm flex-1 justify-center"
											onclick={() => void mergeNow(duplicate)}
										>
											<UsersThreeIcon size={14} /> Merge
										</button>
										<button
											type="button"
											class="chip chip-sm px-2"
											aria-label="Compare before merging"
											title="Compare before merging"
											onclick={() => startMerge(duplicate)}
										>
											<ArrowsLeftRightIcon size={14} />
										</button>
									</span>
								</div>
							{:else}
								<button
									type="button"
									class="text-base-content/60 hover:text-base-content flex aspect-[3/4] w-full flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-white/15 text-sm transition-colors hover:border-white/35 hover:bg-white/5"
									onclick={() => startMerge(null)}
								>
									<UsersThreeIcon size={22} />
									Someone else…
								</button>
							{/if}
						{/snippet}
					</Shelf>

					<!-- The face matches: suggestions only, nothing is credited without a click. -->
					<div id="face-matches">
						<Shelf
							title="Probably also in"
							items={shownFaceMatches}
							key={(match) => match.video.id}
							itemWidth={FACE_MATCH_WIDTH}
						>
							{#snippet card(match)}
								<div class="flex flex-col gap-2">
									<SceneCard
										video={match.video}
										note={`Face match ${Math.round(match.similarity * 100)}%`}
										onclick={() => goto(videoUrl(match.video.id))}
									/>
									<span class="flex gap-1">
										<button
											type="button"
											class="chip chip-sm flex-1 justify-center"
											onclick={() => void creditFaceMatch(match)}
										>
											Add {performer?.name}
										</button>
										<button
											type="button"
											class="chip chip-sm px-2"
											aria-label="Not {performer?.name}"
											title="Not {performer?.name}"
											onclick={() => (dismissedMatches = [...dismissedMatches, match.video.id])}
										>
											<XIcon size={14} />
										</button>
									</span>
								</div>
							{/snippet}
						</Shelf>
					</div>
				</div>
			{/if}
		</div>

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
		{#if choosingPicture}
			<PictureDialog
				performerId={performer.id}
				onpicked={(uri) => void setPicture(uri)}
				onclose={() => (choosingPicture = false)}
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
