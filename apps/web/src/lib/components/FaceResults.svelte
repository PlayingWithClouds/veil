<!--
	What the face searches answered for one picture, one view per kind of result,
	switchable when more than one has any: performers who look alike (best
	first, with their similarity); videos with a similar face, which name nobody
	(one a scene plugin serves opens as a scene in the app, others as a link);
	forum posts with a similar picture (NameThatPorn) with the pornstars their
	answers name. Shared by the "Who is this?" dialog, which credits a name to its
	media, and the face search mode, which has no media and offers the
	performer's page or a search for the name instead.
-->
<script lang="ts">
	import UserCircleIcon from 'phosphor-svelte/lib/UserCircleIcon';
	import { goto } from '$app/navigation';
	import { cacheUrl } from '$lib/img';
	import { homeSearchUrl, performerUrl, sceneUrl } from '$lib/routes';
	import { ingestPluginResult } from '$lib/search';
	import { mediaFilter } from '$lib/stores/mediaFilter';
	import { searchQuery } from '$lib/stores/searchQuery';
	import { notifications } from '$lib/stores/notifications';
	import type { FaceIdentification, FacePostMatch, FaceVideoMatch } from '$lib/faceIdentify';

	interface Props {
		identification: FaceIdentification;
		/** Credits a name to the media; without it, names link to their page or a search. */
		onCredit?: (name: string) => void;
		/** The name being credited, so its button shows a spinner. */
		crediting?: string | null;
		/** Called before the app navigates away, e.g. to close a dialog. */
		onLeave?: () => void;
	}

	let { identification, onCredit, crediting = null, onLeave }: Props = $props();

	type ResultView = 'performers' | 'videos' | 'posts';

	// The view the user picked; null shows the first one with results.
	let pickedView = $state<ResultView | null>(null);
	// The similar video being scraped to open, so its row shows a spinner.
	let openingVideoUrl = $state<string | null>(null);

	// The views that have results, in the order they are offered.
	let views = $derived.by(() => {
		const offered: { key: ResultView; label: string; count: number }[] = [];
		if (identification.candidates.length > 0) {
			offered.push({
				key: 'performers',
				label: 'Performers',
				count: identification.candidates.length
			});
		}
		if (identification.videos.length > 0) {
			offered.push({ key: 'videos', label: 'Similar videos', count: identification.videos.length });
		}
		if (identification.posts.length > 0) {
			offered.push({ key: 'posts', label: 'Posts', count: identification.posts.length });
		}
		return offered;
	});

	let activeView = $derived.by(() => {
		if (pickedView && views.some((view) => view.key === pickedView)) return pickedView;
		if (views.length === 0) return null;
		return views[0].key;
	});

	/** Similarity as a whole percentage. */
	function percent(similarity: number): number {
		return Math.round(Math.max(0, Math.min(1, similarity)) * 100);
	}

	/** The site a link points to, for its label. */
	function siteOf(link: string): string {
		try {
			return new URL(link).hostname.replace(/^www\./, '');
		} catch {
			return link;
		}
	}

	/** "Named in 1 post" / "Named in 3 posts". */
	function postCountLabel(posts: number): string {
		if (posts === 1) return 'Named in 1 post';
		return `Named in ${posts} posts`;
	}

	/** Scrapes a similar video through its plugin and opens it as a scene. */
	async function openVideo(video: FaceVideoMatch) {
		if (!video.scenePlugin || openingVideoUrl) return;
		openingVideoUrl = video.url;
		try {
			const sceneId = await ingestPluginResult(video.scenePlugin, video.url, video.thumbnailUrl);
			onLeave?.();
			await goto(sceneUrl(sceneId));
		} catch (error) {
			console.error('open similar video failed', video.url, error);
			notifications.push('Could not open this video', 'error');
		} finally {
			openingVideoUrl = null;
		}
	}

	/** Searches the sites for a name, leaving the face search mode. */
	function searchName(name: string) {
		onLeave?.();
		mediaFilter.set('all');
		searchQuery.set(name);
		goto(homeSearchUrl(name));
	}
</script>

{#snippet picture(imageUrl: string | null, name: string)}
	{#if imageUrl}
		<img
			src={cacheUrl(imageUrl)}
			alt={name}
			loading="lazy"
			class="h-32 w-24 shrink-0 rounded-lg bg-white/5 object-cover sm:h-40 sm:w-32"
		/>
	{:else}
		<span
			class="text-base-content/20 flex h-32 w-24 shrink-0 items-center justify-center rounded-lg bg-white/5 sm:h-40 sm:w-32"
		>
			<UserCircleIcon size={48} />
		</span>
	{/if}
{/snippet}

{#snippet libraryBadge()}
	<span class="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white/70">In your library</span>
{/snippet}

<!-- Credit with a media to credit to; else the library's performer page, or a search. -->
{#snippet nameAction(name: string, performerId: string | null)}
	{#if onCredit}
		<button
			type="button"
			class="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80 disabled:opacity-50"
			disabled={crediting !== null}
			onclick={() => onCredit(name)}
		>
			{#if crediting === name}
				<span class="loading loading-xs loading-spinner"></span>
			{:else}
				Credit
			{/if}
		</button>
	{:else if performerId}
		<a
			href={performerUrl(performerId)}
			onclick={() => onLeave?.()}
			class="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80"
		>
			Open
		</a>
	{:else}
		<button
			type="button"
			class="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80"
			onclick={() => searchName(name)}
		>
			Search
		</button>
	{/if}
{/snippet}

{#snippet videoRow(video: FaceVideoMatch, opening: boolean)}
	{#if video.thumbnailUrl}
		<img
			src={cacheUrl(video.thumbnailUrl)}
			alt=""
			loading="lazy"
			class="aspect-video w-40 shrink-0 rounded-lg bg-white/5 object-cover sm:w-56"
		/>
	{:else}
		<span class="aspect-video w-40 shrink-0 rounded-lg bg-white/5 sm:w-56"></span>
	{/if}
	<div class="flex min-w-0 flex-1 flex-col gap-1">
		<span class="line-clamp-3 font-semibold">{video.title}</span>
		<div class="flex items-center gap-2 text-xs">
			<span class="bg-primary rounded px-1.5 py-0.5 font-bold text-black">
				{percent(video.similarity)}%
			</span>
			<span class="text-base-content/50">{siteOf(video.url)}</span>
			{#if opening}
				<span class="loading loading-xs loading-spinner"></span>
			{/if}
		</div>
	</div>
{/snippet}

{#snippet postRow(post: FacePostMatch)}
	<a
		href={post.url}
		target="_blank"
		rel="noopener noreferrer"
		class="flex items-center gap-3 rounded-xl bg-white/5 p-2 hover:bg-white/10"
	>
		{#if post.thumbnailUrl}
			<img
				src={cacheUrl(post.thumbnailUrl)}
				alt=""
				loading="lazy"
				class="size-32 shrink-0 rounded-lg bg-white/5 object-cover sm:size-40"
			/>
		{:else}
			<span class="size-32 shrink-0 rounded-lg bg-white/5 sm:size-40"></span>
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<span class="line-clamp-3 font-semibold">{post.title}</span>
			<div class="flex flex-wrap items-center gap-2 text-xs">
				{#if post.solved}
					<span class="bg-primary rounded px-1.5 py-0.5 font-bold text-black">Answered</span>
				{/if}
				{#if post.names.length > 0}
					<span class="text-base-content/70">{post.names.join(', ')}</span>
				{/if}
				<span class="text-base-content/50">{siteOf(post.url)}</span>
			</div>
		</div>
	</a>
{/snippet}

<div class="flex flex-col gap-4">
	{#each identification.failures as failure (failure)}
		<p class="text-warning text-center text-sm">{failure}</p>
	{/each}

	{#if views.length === 0}
		<p class="text-base-content/40 py-10 text-center text-sm">No performer looks like this.</p>
	{/if}

	{#if views.length > 1}
		<div class="flex flex-wrap items-center gap-2" role="tablist">
			{#each views as view (view.key)}
				<button
					type="button"
					role="tab"
					aria-selected={activeView === view.key}
					onclick={() => (pickedView = view.key)}
					class="border-base-300 text-base-content/40 hover:text-base-content/70 rounded-full border px-3 py-1 text-sm font-medium transition-colors"
					class:view-active={activeView === view.key}
				>
					{view.label}
					<span class="opacity-60">{view.count}</span>
				</button>
			{/each}
		</div>
	{/if}

	{#if activeView === 'performers'}
		<ul class="flex flex-col gap-2">
			{#each identification.candidates as candidate (candidate.name)}
				<li class="flex items-center gap-3 rounded-xl bg-white/5 p-2">
					{@render picture(candidate.imageUrl, candidate.name)}
					<div class="flex min-w-0 flex-1 flex-col gap-1">
						<div class="flex flex-wrap items-center gap-2">
							<span class="truncate text-base font-semibold">{candidate.name}</span>
							<span class="bg-primary rounded px-1.5 py-0.5 text-xs font-bold text-black">
								{percent(candidate.similarity)}%
							</span>
							{#if candidate.performerId}
								{@render libraryBadge()}
							{/if}
						</div>
						{#if candidate.profileUrls.length > 0}
							<div class="flex flex-wrap gap-x-3 gap-y-0.5 text-xs">
								{#each candidate.profileUrls as profileUrl (profileUrl)}
									<a
										href={profileUrl}
										target="_blank"
										rel="noopener noreferrer"
										class="text-base-content/50 hover:text-base-content underline"
									>
										{siteOf(profileUrl)}
									</a>
								{/each}
							</div>
						{/if}
					</div>
					{@render nameAction(candidate.name, candidate.performerId)}
				</li>
			{/each}
		</ul>
	{:else if activeView === 'videos'}
		<ul class="flex flex-col gap-2">
			{#each identification.videos as video (video.url)}
				<li>
					{#if video.scenePlugin}
						<button
							type="button"
							class="flex w-full items-center gap-3 rounded-xl bg-white/5 p-2 text-left hover:bg-white/10 disabled:opacity-50"
							disabled={openingVideoUrl !== null}
							onclick={() => openVideo(video)}
						>
							{@render videoRow(video, openingVideoUrl === video.url)}
						</button>
					{:else}
						<a
							href={video.url}
							target="_blank"
							rel="noopener noreferrer"
							class="flex items-center gap-3 rounded-xl bg-white/5 p-2 hover:bg-white/10"
						>
							{@render videoRow(video, false)}
						</a>
					{/if}
				</li>
			{/each}
		</ul>
	{:else if activeView === 'posts'}
		{#if identification.postPerformers.length > 0}
			<section class="flex flex-col gap-2">
				<h3 class="text-base-content/50 text-sm font-semibold">Named in these posts</h3>
				<ul class="flex flex-col gap-2">
					{#each identification.postPerformers as performer (performer.name)}
						<li class="flex items-center gap-3 rounded-xl bg-white/5 p-2">
							{@render picture(performer.imageUrl, performer.name)}
							<div class="flex min-w-0 flex-1 flex-col gap-1">
								<div class="flex flex-wrap items-center gap-2">
									<span class="truncate text-base font-semibold">{performer.name}</span>
									{#if performer.performerId}
										{@render libraryBadge()}
									{/if}
								</div>
								<a
									href={performer.postUrl}
									target="_blank"
									rel="noopener noreferrer"
									class="text-base-content/50 hover:text-base-content text-xs underline"
								>
									{postCountLabel(performer.posts)}
								</a>
							</div>
							{@render nameAction(performer.name, performer.performerId)}
						</li>
					{/each}
				</ul>
			</section>
		{/if}
		<section class="flex flex-col gap-2">
			<h3 class="text-base-content/50 text-sm font-semibold">Posts with a similar picture</h3>
			<ul class="flex flex-col gap-2">
				{#each identification.posts as post (post.url)}
					<li>{@render postRow(post)}</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>

<style>
	.view-active {
		border-color: var(--color-primary);
		background-color: color-mix(in oklab, var(--color-primary) 15%, transparent);
		color: var(--color-primary);
	}
</style>
