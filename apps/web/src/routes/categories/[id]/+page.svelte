<!--
	One category: its group, the words it matches and its subcategories, then
	an endless grid of one kind of its media (?kind=, videos by default): the
	library's first, newest first, then (for a category without
	subcategories) what the sites find for its name, stored as it comes.
	Edit, delete, and add subcategories to a group.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import CategoryEditor from '$lib/components/CategoryEditor.svelte';
	import CategoryKindTabs from '$lib/components/CategoryKindTabs.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import MediaFeed from '$lib/components/MediaFeed.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import StoryCard from '$lib/components/StoryCard.svelte';
	import {
		categoryKindOf,
		deleteCategory,
		fetchCategory,
		type CategoryDetail
	} from '$lib/categories';
	import { endlessFeed, siteSource, type Feed, type Source } from '$lib/feed';
	import { browseGalleries, fetchGalleryListing } from '$lib/galleries';
	import {
		browseVideos,
		fetchVideosMatching,
		type GalleryCardRow,
		type VideoCardRow
	} from '$lib/library';
	import {
		browseAudios,
		browseStories,
		fetchAudios,
		fetchStories,
		type AudioCardRow,
		type StoryCardRow
	} from '$lib/media';
	import { listerNames } from '$lib/stores/sites';
	import {
		audioUrl,
		categoriesUrl,
		categoryUrl,
		galleryUrl,
		storyUrl,
		videoUrl
	} from '$lib/routes';
	import type { ListKind } from '@playingwithclouds/veil-sdk/generated';

	const SURFACE = 'category';

	let category = $state<CategoryDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let editing = $state(false);
	let addingChild = $state(false);

	let id = $derived(page.params.id ?? '');
	let current = $derived(categoryKindOf(page.url.searchParams.get('kind')));
	// Restarts the grid when the category, its name (what the sites are asked) or the kind changes.
	let feedKey = $derived(`${category?.id}:${category?.name}:${current.kind}`);

	$effect(() => {
		void load(id);
	});

	/** Loads the category. */
	async function load(categoryId: string) {
		loading = true;
		error = null;
		try {
			const found = await fetchCategory(categoryId);
			if (categoryId === id) category = found;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/**
	 * The sites asked for the category's name once the library is through; a
	 * group asks none (its name isn't what sites tag).
	 */
	function sitesFor<Item>(
		kind: ListKind,
		browsePage: (plugin: string, limit: number, offset: number, query: string) => Promise<Item[]>
	): Source<Item> | undefined {
		if (!category || category.children.length > 0) return undefined;
		const query = category.name;
		return siteSource(
			() => listerNames(kind),
			(plugin, limit, offset) => browsePage(plugin, limit, offset, query)
		);
	}

	/** The category's videos, then the sites'. */
	function videoFeed(): Feed<VideoCardRow> {
		const filter = { categoryIds: [id] };
		return endlessFeed({
			rest: (limit, offset) => fetchVideosMatching(filter, limit, offset),
			sites: sitesFor('VIDEO', browseVideos)
		});
	}

	/** The category's galleries, then the sites'. */
	function galleryFeed(): Feed<GalleryCardRow> {
		return endlessFeed({
			rest: (limit, offset) => fetchGalleryListing(limit, offset, { categoryId: id }),
			sites: sitesFor('GALLERY', browseGalleries)
		});
	}

	/** The category's audio, then the sites'. */
	function audioFeed(): Feed<AudioCardRow> {
		return endlessFeed({
			rest: (limit, offset) => fetchAudios(limit, offset, { categoryId: id }),
			sites: sitesFor('AUDIO', browseAudios)
		});
	}

	/** The category's stories, then the sites'. */
	function storyFeed(): Feed<StoryCardRow> {
		return endlessFeed({
			rest: (limit, offset) => fetchStories(limit, offset, { categoryId: id }),
			sites: sitesFor('STORY', browseStories)
		});
	}

	/** Deletes the category, after asking, and goes back to the list. */
	async function remove() {
		if (!category) return;
		const question = `Delete “${category.name}”? Its subcategories move up a level. Media isn't touched.`;
		if (!confirm(question)) return;
		try {
			await deleteCategory(category.id);
			goto(categoriesUrl(current.kind));
		} catch (failure) {
			error = String(failure);
		}
	}

	/** Takes the saved category in. */
	function saved(changed: CategoryDetail) {
		editing = false;
		category = changed;
	}

	/** Goes to the new subcategory. */
	function childCreated(child: CategoryDetail) {
		addingChild = false;
		goto(categoryUrl(child.id, current.kind));
	}
</script>

<svelte:head><title>{category?.name ?? 'Category'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6 pt-2">
	{#if loading && !category}
		<p class="text-base-content/50 text-sm">Loading…</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if !loading && !category && !error}
		<p class="text-base-content/50 text-sm">This category doesn't exist (any more).</p>
	{/if}

	{#if category}
		<header class="flex flex-col gap-3">
			<nav class="text-base-content/60 flex items-center gap-1.5 text-sm">
				<a href={categoriesUrl(current.kind)} class="hover:text-base-content">Categories</a>
				{#if category.parent}
					<span aria-hidden="true">›</span>
					<a href={categoryUrl(category.parent.id, current.kind)} class="hover:text-base-content"
						>{category.parent.name}</a
					>
				{/if}
			</nav>
			<div class="flex flex-wrap items-center gap-3">
				<h1 class="text-3xl font-bold tracking-tight">{category.name}</h1>
				<div class="ml-auto flex gap-2">
					{#if !category.parent}
						<button type="button" class="btn btn-sm" onclick={() => (addingChild = true)}>
							<PlusIcon size={16} /> Add a category here
						</button>
					{/if}
					<button type="button" class="btn btn-sm" onclick={() => (editing = true)}>
						<PencilSimpleIcon size={16} /> Edit
					</button>
					<button type="button" class="btn btn-sm btn-ghost text-error" onclick={remove}>
						<TrashIcon size={16} /> Delete
					</button>
				</div>
			</div>
			{#if category.matchTerms.length > 0}
				<p class="text-base-content/60 text-sm">
					Tags it matches: {category.matchTerms.join(', ')}
				</p>
			{/if}
			{#if category.children.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each category.children as child (child.id)}
						<a
							href={categoryUrl(child.id, current.kind)}
							class="bg-base-200 hover:bg-base-300 rounded-full px-3 py-1 text-sm">{child.name}</a
						>
					{/each}
				</div>
			{/if}
		</header>

		<CategoryKindTabs {current} hrefOf={(entry) => categoryUrl(id, entry.kind)} />

		{#key feedKey}
			{#if current.kind === 'video'}
				<MediaFeed
					feed={videoFeed}
					cardWidth={260}
					surface={SURFACE}
					empty="No videos in this category."
				>
					{#snippet card(video, shown)}
						<SceneCard {video} {shown} onclick={() => goto(videoUrl(video.id))} />
					{/snippet}
				</MediaFeed>
			{:else if current.kind === 'gallery'}
				<MediaFeed
					feed={galleryFeed}
					cardWidth={180}
					surface={SURFACE}
					empty="No galleries in this category."
				>
					{#snippet card(gallery, shown)}
						<GalleryCard {gallery} {shown} onclick={() => goto(galleryUrl(gallery.id))} />
					{/snippet}
				</MediaFeed>
			{:else if current.kind === 'audio'}
				<MediaFeed
					feed={audioFeed}
					cardWidth={180}
					surface={SURFACE}
					empty="No audio in this category."
				>
					{#snippet card(audio)}
						<AudioCard {audio} onclick={() => goto(audioUrl(audio.id))} />
					{/snippet}
				</MediaFeed>
			{:else}
				<MediaFeed
					feed={storyFeed}
					cardWidth={150}
					surface={SURFACE}
					empty="No stories in this category."
				>
					{#snippet card(story)}
						<StoryCard {story} onclick={() => goto(storyUrl(story.id))} />
					{/snippet}
				</MediaFeed>
			{/if}
		{/key}
	{/if}
</div>

{#if category && editing}
	<CategoryEditor {category} onsaved={saved} onclose={() => (editing = false)} />
{/if}
{#if category && addingChild}
	<CategoryEditor
		parentId={category.id}
		onsaved={childCreated}
		onclose={() => (addingChild = false)}
	/>
{/if}
