<!--
	Videos of a performer or studio page as one feed of Home's cards: library
	scenes credited to it first (paged in on scroll), then scenes found live on
	the sites by its name. Chips filter to the library or sort the whole feed.
	One column on phones.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { untrack } from 'svelte';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ExploreSceneCard from './ExploreSceneCard.svelte';
	import {
		entityFilter,
		fetchEntitySceneHits,
		pluginSearchStream,
		streamResultToGridItem,
		type EntityFilter,
		type SceneHit,
		type SearchStream
	} from '$lib/explore';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { sceneUrl } from '$lib/routes';

	interface Props {
		entityKind: 'performer' | 'studio';
		entityId: string;
		// Searched on the sites for more scenes.
		name: string;
	}

	let { entityKind, entityId, name }: Props = $props();

	let filter = $derived(entityFilter(entityKind, entityId));

	type View = 'all' | 'library' | 'newest' | 'longest';

	const VIEWS: { key: View; label: string }[] = [
		{ key: 'all', label: 'All' },
		{ key: 'library', label: 'In library' },
		{ key: 'newest', label: 'Newest' },
		{ key: 'longest', label: 'Longest' }
	];
	const PAGE_SIZE = 40;
	const SKELETON_CARDS = 3;

	let view = $state<View>('all');
	let library = $state.raw<SceneHit[]>([]);
	let web = $state.raw<SceneHit[]>([]);
	let libraryLoading = $state(false);
	let libraryDone = false;
	let webSearching = $state(false);
	let sites = $state<SearchPlugin[]>([]);
	let sentinel = $state<HTMLElement | null>(null);
	let stream: SearchStream | null = null;
	// Stale page loads (after a refresh or another entity) are dropped.
	let requestToken = 0;

	let shown = $derived(scenesForView(view, library, web));

	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	// A new entity (navigating between performers) starts over.
	$effect(() => {
		const current = entityFilter(entityKind, entityId);
		const currentName = name;
		untrack(() => reload(current, currentName));
		return () => stream?.close();
	});

	$effect(() => {
		if (!sentinel) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) loadMoreLibrary();
			},
			{ rootMargin: '800px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	/** Reloads the library scenes and restarts the site search. Resolves once the first page is in. */
	export async function reload(current: EntityFilter = filter, currentName: string = name) {
		const token = ++requestToken;
		stream?.close();
		library = [];
		web = [];
		libraryDone = false;
		libraryLoading = true;
		const firstPage = await fetchEntitySceneHits(current, PAGE_SIZE, 0);
		if (token !== requestToken) return;
		library = firstPage;
		libraryDone = firstPage.length < PAGE_SIZE;
		libraryLoading = false;
		searchSites(currentName, token);
	}

	/** Appends the next page of library scenes. */
	async function loadMoreLibrary() {
		if (libraryLoading || libraryDone) return;
		const token = requestToken;
		libraryLoading = true;
		const next = await fetchEntitySceneHits(filter, PAGE_SIZE, library.length);
		if (token !== requestToken) return;
		library = [...library, ...next];
		libraryDone = next.length < PAGE_SIZE;
		libraryLoading = false;
	}

	/** Streams in scenes the sites list for the name, skipping ones already shown. */
	function searchSites(query: string, token: number) {
		webSearching = true;
		stream = pluginSearchStream(
			query,
			[],
			(result) => {
				if (token !== requestToken || result.mediaType !== 'scene') return;
				const item = streamResultToGridItem(result);
				if (!item || item.kind !== 'scene' || isShown(item.scene)) return;
				web = [...web, item.scene];
			},
			() => {
				if (token === requestToken) webSearching = false;
			}
		);
	}

	/** Whether the scene (or one with the same title) is already in the feed. */
	function isShown(scene: SceneHit): boolean {
		const title = scene.title.toLowerCase();
		for (const shownScene of [...library, ...web]) {
			if (shownScene.id === scene.id || shownScene.title.toLowerCase() === title) return true;
		}
		return false;
	}

	/** The feed for the selected chip. */
	function scenesForView(selected: View, libraryScenes: SceneHit[], webScenes: SceneHit[]): SceneHit[] {
		if (selected === 'library') return libraryScenes;
		const all = [...libraryScenes, ...webScenes];
		if (selected === 'newest') return all.toSorted(byNewest);
		if (selected === 'longest') return all.toSorted(byLongest);
		return all;
	}

	/** Newest release first; undated scenes last. */
	function byNewest(first: SceneHit, second: SceneHit): number {
		const firstDate = first.date ?? first.createdAt ?? '';
		const secondDate = second.date ?? second.createdAt ?? '';
		return secondDate.localeCompare(firstDate);
	}

	/** Longest first; unknown lengths last. */
	function byLongest(first: SceneHit, second: SceneHit): number {
		return (second.durationSeconds ?? -1) - (first.durationSeconds ?? -1);
	}

	/** The site a scene came from, for its caption. */
	function siteOf(scene: SceneHit): SearchPlugin | null {
		return pluginForUrl(scene.sourceUrl, sites);
	}
</script>

<div class="flex flex-col gap-4">
	<div class="no-scrollbar -mx-4 flex gap-2 overflow-x-auto px-4 max-md:-mx-3 max-md:px-3">
		{#each VIEWS as option (option.key)}
			<button
				type="button"
				aria-pressed={view === option.key}
				onclick={() => (view = option.key)}
				class="feed-chip bg-base-200 flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium"
				class:feed-chip-active={view === option.key}
			>
				{option.label}
			</button>
		{/each}
		{#if webSearching}
			<span class="text-base-content/40 flex shrink-0 items-center gap-2 pl-1 text-xs">
				<span class="loading loading-spinner loading-xs"></span>
				Searching sites…
			</span>
		{/if}
	</div>

	{#if shown.length === 0 && libraryLoading}
		<div class="feed-grid grid gap-6">
			{#each { length: SKELETON_CARDS } as _, index (index)}
				<div class="flex flex-col gap-3">
					<div class="skeleton aspect-video w-full rounded-[var(--tv-radius)]"></div>
					<div class="flex gap-3">
						<div class="skeleton h-9 w-9 shrink-0 rounded-full"></div>
						<div class="flex flex-1 flex-col gap-2 pt-1">
							<div class="skeleton h-3.5 w-11/12 rounded"></div>
							<div class="skeleton h-3 w-1/2 rounded"></div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{:else if shown.length === 0 && !webSearching}
		<div class="text-base-content/40 flex flex-col items-center gap-3 py-16 text-center text-sm">
			<MagnifyingGlassIcon size={40} class="opacity-30" />
			{#if view === 'library'}
				Nothing in your library yet.
			{:else}
				No videos found.
			{/if}
		</div>
	{:else}
		<div class="feed-grid grid gap-6">
			{#each shown as scene (scene.id)}
				<ExploreSceneCard {scene} site={siteOf(scene)} onclick={() => goto(sceneUrl(scene.id))} />
			{/each}
		</div>
	{/if}
	<div bind:this={sentinel} class="h-1"></div>
</div>

<style>
	/* One column on phones; wider screens fill the row like Home. */
	.feed-grid {
		grid-template-columns: minmax(0, 1fr);
	}

	@media (width >= 720px) and (not ((height < 540px) and (pointer: coarse))) {
		.feed-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		}
	}

	.feed-chip {
		transition: background-color 150ms;
	}

	.feed-chip-active {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
