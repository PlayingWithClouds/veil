<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import { cacheUrl } from '$lib/img';
	import { galleryUrl } from '$lib/routes';
	import { fetchGalleryCategories, type GalleryCategory } from '$lib/galleries';
	import { pluginSearchStream, type PluginSearchResult, type SearchStream } from '$lib/search';

	const PAGE_SIZE = 24;

	// With ?category=<name> the page shows that category's feed; without it, the
	// category tile index.
	let category = $derived($page.url.searchParams.get('category'));

	let categories = $state<GalleryCategory[]>([]);
	let categoriesLoading = $state(true);

	$effect(() => {
		fetchGalleryCategories().then((list) => {
			categories = list;
			categoriesLoading = false;
		});
	});

	function openCategory(entry: GalleryCategory) {
		goto(`/galleries?category=${encodeURIComponent(entry.name)}`);
	}

	// --- category feed (streamed from the pornpics plugin) --------------------
	let items = $state<PluginSearchResult[]>([]);
	let loading = $state(false);
	let done = $state(false);
	let sentinel = $state<HTMLElement | null>(null);
	let offset = 0;
	let stream: SearchStream | null = null;
	let seen = new Set<string>();

	// Restart the feed whenever the category query param changes.
	let feedFor = $state<string | null>(null);
	$effect(() => {
		const current = category;
		if (feedFor === current) return;
		feedFor = current;
		stream?.close();
		stream = null;
		items = [];
		seen = new Set();
		offset = 0;
		done = false;
		loading = false;
		if (current) loadMore();
	});

	function loadMore() {
		if (loading || done || !category) return;
		loading = true;
		const term = category.toLowerCase();
		const pageOffset = offset;
		let received = 0;
		stream = pluginSearchStream(
			term,
			['pornpics'],
			(result) => {
				if (result.mediaType !== 'gallery') return;
				if (seen.has(result.externalId)) return;
				seen.add(result.externalId);
				received += 1;
				items = [...items, result];
			},
			() => {
				loading = false;
				offset = pageOffset + PAGE_SIZE;
				if (received === 0) done = true;
			},
			{ offset: pageOffset, limit: PAGE_SIZE }
		);
	}

	function openGallery(result: PluginSearchResult) {
		if (result.dbId) goto(galleryUrl(result.dbId));
	}

	$effect(() => {
		if (!sentinel) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) loadMore();
			},
			{ rootMargin: '600px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});
</script>

{#if category}
	<div class="flex w-full flex-col gap-5">
		<div class="flex items-center gap-3">
			<a class="btn btn-square btn-ghost btn-sm" aria-label="All categories" href="/galleries">
				<ArrowLeftIcon size={18} />
			</a>
			<div>
				<h1 class="text-2xl font-bold capitalize tracking-tight">{category}</h1>
				<p class="text-base-content/40 text-sm">Image sets from PornPics.</p>
			</div>
			<span class="badge badge-error ml-auto">18+</span>
		</div>

		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
			{#each items as gallery (gallery.externalId)}
				<button
					type="button"
					class="group flex flex-col gap-2 text-left"
					onclick={() => openGallery(gallery)}
				>
					<div class="bg-base-200 relative aspect-[3/4] overflow-hidden rounded-xl">
						{#if cacheUrl(gallery.posterUrl)}
							<img
								src={cacheUrl(gallery.posterUrl)}
								alt={gallery.title}
								loading="lazy"
								class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
							/>
						{:else}
							<div class="flex h-full w-full items-center justify-center">
								<ImagesIcon size={32} class="opacity-20" />
							</div>
						{/if}
					</div>
					<span class="line-clamp-2 text-xs font-medium">{gallery.title}</span>
				</button>
			{/each}
		</div>

		{#if loading}
			<div class="flex justify-center py-6">
				<span class="loading loading-spinner text-base-content/40"></span>
			</div>
		{/if}
		<div bind:this={sentinel} class="h-1"></div>
	</div>
{:else}
	<div class="flex w-full flex-col gap-5">
		<div class="flex items-center gap-3">
			<span class="text-base-content/70"><ImagesIcon size={26} /></span>
			<div>
				<h1 class="text-2xl font-bold tracking-tight">Galleries</h1>
				<p class="text-base-content/40 text-sm">Browse PornPics galleries by category.</p>
			</div>
			<span class="badge badge-error ml-auto">18+</span>
		</div>

		{#if categoriesLoading}
			<div class="flex justify-center py-16">
				<span class="loading loading-spinner text-base-content/40"></span>
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
				{#each categories as entry (entry.id)}
					<button
						type="button"
						class="group relative aspect-square overflow-hidden rounded-xl bg-base-200 text-left"
						onclick={() => openCategory(entry)}
					>
						{#if cacheUrl(entry.poster)}
							<img
								src={cacheUrl(entry.poster)}
								alt={entry.name}
								loading="lazy"
								class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
							/>
						{/if}
						<div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"></div>
						<span class="absolute inset-x-0 bottom-0 p-2 text-sm font-semibold text-white drop-shadow">
							{entry.name}
						</span>
					</button>
				{/each}
			</div>
		{/if}
	</div>
{/if}
