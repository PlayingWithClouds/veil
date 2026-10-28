<script lang="ts">
	// Home's recommendation feed: Continue watching on top, then the ranked
	// scenes in an even 16:9 grid broken every few rows by a shelf: the engine's
	// titled rows ("More from <performer>", "Because you watched …") taking
	// turns with rows of portrait gallery covers. Every feed card says why it
	// was picked. On the images filter the galleries are the grid.
	import { goto } from '$app/navigation';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import SceneShelf from './SceneShelf.svelte';
	import { impression, recordClicked, type ShownEvent } from '$lib/recommendations';
	import { galleryUrl, performerUrl, sceneUrl, studioUrl, tagUrl } from '$lib/routes';
	import { pluginForUrl, type SearchPlugin } from '$lib/search';
	import type { GalleryHit, RecommendationShelf, SceneHit } from '$lib/explore';

	interface Props {
		scenes: SceneHit[];
		galleries: GalleryHit[];
		shelves: RecommendationShelf[];
		continueScenes: SceneHit[];
		sites: SearchPlugin[];
		// Whether every page of the feed's scenes is loaded (the leftover shelves then follow).
		scenesDone: boolean;
		// True on the images filter: the galleries are the main grid.
		galleriesOnly: boolean;
		impressionFor: (scene: SceneHit) => ShownEvent;
		badgeFor: (tags: string[]) => string | null;
		onopen: (scene: SceneHit) => void;
	}

	let {
		scenes,
		galleries,
		shelves,
		continueScenes,
		sites,
		scenesDone,
		galleriesOnly,
		impressionFor,
		badgeFor,
		onopen
	}: Props = $props();

	// Grid geometry: a shelf follows every ROWS_PER_CHUNK rows of scenes.
	const CARD_MIN_WIDTH = 280;
	const GRID_GAP = 24;
	const ROWS_PER_CHUNK = 3;
	const GALLERIES_PER_SHELF = 12;

	type Separator =
		| { kind: 'shelf'; shelf: RecommendationShelf }
		| { kind: 'galleries'; key: string; galleries: GalleryHit[] };
	type Segment =
		| { kind: 'grid'; key: string; scenes: SceneHit[] }
		| Separator;

	let gridWidth = $state(0);
	let columns = $derived(Math.max(1, Math.floor((gridWidth + GRID_GAP) / (CARD_MIN_WIDTH + GRID_GAP))));
	let segments = $derived(buildSegments(scenes, separators(shelves, galleries), columns * ROWS_PER_CHUNK));

	/** The shelves between grid chunks: titled rows and gallery rows taking turns. */
	function separators(rows: RecommendationShelf[], covers: GalleryHit[]): Separator[] {
		const galleryShelves: Separator[] = [];
		for (let start = 0; start < covers.length; start += GALLERIES_PER_SHELF) {
			galleryShelves.push({
				kind: 'galleries',
				key: `galleries:${start}`,
				galleries: covers.slice(start, start + GALLERIES_PER_SHELF)
			});
		}
		const rowShelves: Separator[] = rows.map((shelf) => ({ kind: 'shelf', shelf }));
		const mixed: Separator[] = [];
		const longest = Math.max(rowShelves.length, galleryShelves.length);
		for (let index = 0; index < longest; index++) {
			if (index < rowShelves.length) mixed.push(rowShelves[index]);
			if (index < galleryShelves.length) mixed.push(galleryShelves[index]);
		}
		return mixed;
	}

	/**
	 * Cuts the scenes into chunks of chunkSize with a shelf after each full
	 * chunk; leftover shelves follow once the feed has no more scenes.
	 */
	function buildSegments(feed: SceneHit[], between: Separator[], chunkSize: number): Segment[] {
		const built: Segment[] = [];
		let separatorIndex = 0;
		for (let start = 0; start < feed.length; start += chunkSize) {
			const chunk = feed.slice(start, start + chunkSize);
			built.push({ kind: 'grid', key: `grid:${start}`, scenes: chunk });
			const full = chunk.length === chunkSize;
			if (full && separatorIndex < between.length) {
				built.push(between[separatorIndex]);
				separatorIndex++;
			}
		}
		if (scenesDone) built.push(...between.slice(separatorIndex));
		return built;
	}

	/** Stable key of a segment for the keyed each. */
	function segmentKey(segment: Segment): string {
		if (segment.kind === 'shelf') return `shelf:${segment.shelf.key}`;
		return segment.key;
	}

	/** The page a titled row is about, when there is one. */
	function shelfHref(shelf: RecommendationShelf): string | null {
		const entityId = shelf.reason.entityId;
		if (!entityId) return null;
		switch (shelf.reason.kind) {
			case 'performer':
				return performerUrl(entityId);
			case 'studio':
				return studioUrl(entityId);
			case 'tag':
			case 'explore':
				return tagUrl(entityId);
			case 'related':
				return sceneUrl(entityId);
			default:
				return null;
		}
	}

	/** The impression of a titled row's card. */
	function shelfImpression(shelf: RecommendationShelf, position: number): ShownEvent {
		const item = shelf.items[position];
		return { mediaId: item.scene.id, source: item.source, surface: 'row', position };
	}

	/** Opens a titled row's card, logging the click for the engine. */
	function openShelfScene(shelf: RecommendationShelf, position: number) {
		recordClicked(shelfImpression(shelf, position));
		goto(sceneUrl(shelf.items[position].scene.id));
	}
</script>

<div class="flex flex-col gap-10">
	{#if continueScenes.length > 0 && !galleriesOnly}
		<SceneShelf
			title="Continue watching"
			href="/history"
			scenes={continueScenes}
			{sites}
			onopen={(scene) => goto(sceneUrl(scene.id))}
		/>
	{/if}

	{#if galleriesOnly}
		<div class="gallery-grid grid gap-6">
			{#each galleries as gallery (gallery.id)}
				{@render galleryCard(gallery)}
			{/each}
		</div>
	{:else}
		<div class="flex flex-col gap-10" bind:clientWidth={gridWidth}>
			{#each segments as segment (segmentKey(segment))}
				{#if segment.kind === 'grid'}
					<div class="video-grid grid gap-x-6 gap-y-8">
						{#each segment.scenes as scene (scene.id)}
							<div use:impression={impressionFor(scene)}>
								<ExploreSceneCard
									{scene}
									site={pluginForUrl(scene.sourceUrl, sites)}
									preferredTag={badgeFor(scene.tags)}
									note={scene.reason}
									onclick={() => onopen(scene)}
								/>
							</div>
						{/each}
					</div>
				{:else if segment.kind === 'shelf'}
					<SceneShelf
						title={segment.shelf.title}
						href={shelfHref(segment.shelf)}
						scenes={segment.shelf.items.map((item) => item.scene)}
						{sites}
						impressionAt={(_scene, position) => shelfImpression(segment.shelf, position)}
						onopen={(_scene, position) => openShelfScene(segment.shelf, position)}
					/>
				{:else}
					<section class="flex flex-col gap-3" aria-label="Galleries">
						<h2 class="text-lg font-semibold">Galleries</h2>
						<div class="no-scrollbar -mx-1 flex gap-4 overflow-x-auto px-1 pb-1">
							{#each segment.galleries as gallery (gallery.id)}
								<div class="w-44 shrink-0">
									{@render galleryCard(gallery)}
								</div>
							{/each}
						</div>
					</section>
				{/if}
			{/each}
		</div>
	{/if}
</div>

{#snippet galleryCard(gallery: GalleryHit)}
	<GalleryCard
		{gallery}
		site={pluginForUrl(gallery.sourceUrl, sites)}
		preferredTag={badgeFor(gallery.tags)}
		onclick={() => goto(galleryUrl(gallery.id))}
	/>
{/snippet}

<style>
	.video-grid {
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	}

	.gallery-grid {
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
	}
</style>
