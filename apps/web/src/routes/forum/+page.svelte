<script lang="ts">
	// One thread of an identification forum (a NameThatPorn post face search
	// found), shown in the app instead of on the site: the asked-about picture or
	// clip, then the comments as a chat with the accepted answer marked. Links a
	// scene plugin serves get a Watch button that opens them as a scene; tagged
	// names open the library performer or search the sites.
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import CheckCircleIcon from 'phosphor-svelte/lib/CheckCircleIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import ThumbsUpIcon from 'phosphor-svelte/lib/ThumbsUpIcon';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { cacheUrl } from '$lib/img';
	import { forumThreadUrl, performerUrl, sceneUrl } from '$lib/routes';
	import { ingestPluginResult } from '$lib/search';
	import { searchSitesFor } from '$lib/faceSearch.svelte';
	import {
		fetchForumThread,
		type ForumComment,
		type ForumLink,
		type ForumThread
	} from '$lib/forumThread';
	import { notifications } from '$lib/stores/notifications';
	import { formatRelativeTime } from '$lib/time';
	import { initials } from '$lib/components/scene/format';

	let threadUrl = $derived($page.url.searchParams.get('url') || '');
	let thread = $state<ForumThread | null>(null);
	let loading = $state(true);
	let errorText = $state<string | null>(null);
	// The link being stored as a scene, so its button shows a spinner.
	let watchingUrl = $state<string | null>(null);
	// Comment pictures that failed to load.
	let brokenImages = $state<Set<string>>(new Set());

	// The links the app can play, accepted answer first, each once.
	let watchable = $derived.by(() => {
		if (!thread) return [];
		const ordered = [
			...thread.comments.filter((comment) => comment.accepted),
			...thread.comments.filter((comment) => !comment.accepted)
		];
		const links: ForumLink[] = [];
		for (const comment of ordered) {
			for (const link of comment.links) {
				if (link.scenePlugin && !links.some((known) => known.url === link.url)) links.push(link);
			}
		}
		return links;
	});

	$effect(() => {
		load(threadUrl);
	});

	/** Reads the thread live from its site. */
	async function load(url: string) {
		loading = true;
		thread = null;
		errorText = null;
		if (url === '') {
			errorText = 'No thread to show.';
			loading = false;
			return;
		}
		try {
			thread = await fetchForumThread(url);
		} catch (error) {
			console.error('forumThread failed', url, error);
			errorText = error instanceof Error ? error.message : 'Could not read this thread.';
		} finally {
			loading = false;
		}
	}

	/** Stores a link's video as a scene through its plugin and opens it. */
	async function watch(link: ForumLink) {
		if (!link.scenePlugin || watchingUrl) return;
		watchingUrl = link.url;
		try {
			const sceneId = await ingestPluginResult(link.scenePlugin, link.url, null);
			await goto(sceneUrl(sceneId));
		} catch (error) {
			console.error('watch forum link failed', link.url, error);
			notifications.push('Could not open this video', 'error');
		} finally {
			watchingUrl = null;
		}
	}

	/** "pornhub.com/view_video.php?viewkey=…": a link's site and path, without the scheme. */
	function linkLabel(url: string): string {
		try {
			const parsed = new URL(url);
			return parsed.hostname.replace(/^www\./, '') + parsed.pathname + parsed.search;
		} catch {
			return url;
		}
	}
</script>

<svelte:head>
	<title>{thread ? thread.title : 'Thread'} · Veil</title>
</svelte:head>

{#snippet watchButton(link: ForumLink)}
	<button
		type="button"
		class="flex shrink-0 items-center gap-1.5 rounded-full bg-white px-3 py-1.5 text-sm font-semibold text-black transition-colors hover:bg-white/80 disabled:opacity-50"
		disabled={watchingUrl !== null}
		onclick={() => watch(link)}
	>
		{#if watchingUrl === link.url}
			<span class="loading loading-xs loading-spinner"></span>
		{:else}
			<PlayIcon size={14} weight="fill" />
		{/if}
		Watch
	</button>
{/snippet}

{#snippet linkRow(link: ForumLink)}
	<li class="flex items-center gap-2 rounded-lg bg-black/20 py-1.5 pr-1.5 pl-3">
		<a
			href={link.url}
			target="_blank"
			rel="noopener noreferrer"
			class="text-base-content/70 hover:text-base-content flex min-w-0 flex-1 items-center gap-1.5 text-sm"
			title={link.url}
		>
			<span class="truncate">{linkLabel(link.url)}</span>
			<ArrowSquareOutIcon size={14} class="shrink-0 opacity-60" />
		</a>
		{#if link.scenePlugin}
			{@render watchButton(link)}
		{/if}
	</li>
{/snippet}

{#snippet commentBubble(comment: ForumComment)}
	<li class="flex gap-3">
		<span
			class="bg-base-300 text-base-content/60 flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
			aria-hidden="true"
		>
			{initials(comment.author || '?')}
		</span>
		<div
			class="bg-base-200 flex min-w-0 flex-1 flex-col gap-2 rounded-2xl rounded-tl-sm p-3"
			class:accepted-answer={comment.accepted}
		>
			<p class="text-base-content/50 flex flex-wrap items-center gap-x-2 text-xs">
				<span class="text-base-content/80 font-semibold">{comment.author || 'Anonymous'}</span>
				{#if comment.date}
					<span>{formatRelativeTime(comment.date)}</span>
				{/if}
				{#if comment.likes}
					<span class="flex items-center gap-1"><ThumbsUpIcon size={12} />{comment.likes}</span>
				{/if}
				{#if comment.accepted}
					<span class="text-primary flex items-center gap-1 font-semibold">
						<CheckCircleIcon size={14} weight="fill" />
						Accepted answer
					</span>
				{/if}
			</p>
			{#if comment.text}
				<p class="text-sm leading-relaxed break-words whitespace-pre-line">{comment.text}</p>
			{/if}
			{#if comment.names.length > 0 || comment.titles.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each comment.names as name (name.name)}
						{#if name.performerId}
							<a href={performerUrl(name.performerId)} class="thread-chip">
								{name.name}
								<span class="opacity-60">In your library</span>
							</a>
						{:else}
							<button type="button" class="thread-chip" onclick={() => searchSitesFor(name.name)}>
								{name.name}
								<span class="opacity-60">Search</span>
							</button>
						{/if}
					{/each}
					{#each comment.titles as title (title)}
						<button type="button" class="thread-chip" onclick={() => searchSitesFor(title)}>
							“{title}”
							<span class="opacity-60">Search</span>
						</button>
					{/each}
				</div>
			{/if}
			{#if comment.links.length > 0}
				<ul class="flex flex-col gap-1.5">
					{#each comment.links as link (link.url)}
						{@render linkRow(link)}
					{/each}
				</ul>
			{/if}
			{#if comment.images.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each comment.images as image (image)}
						<!-- Pictures on dead hosts are common in old threads: hide them instead of an empty box. -->
						<a
							href={image}
							target="_blank"
							rel="noopener noreferrer"
							class:hidden={brokenImages.has(image)}
						>
							<img
								src={cacheUrl(image)}
								alt=""
								loading="lazy"
								class="h-40 rounded-lg bg-white/5 object-cover"
								onerror={() => (brokenImages = new Set([...brokenImages, image]))}
							/>
						</a>
					{/each}
				</div>
			{/if}
			{#if comment.posts.length > 0}
				<ul class="flex flex-col gap-1.5">
					{#each comment.posts as post (post.url)}
						<li>
							<a
								href={forumThreadUrl(post.url)}
								class="flex items-center gap-3 rounded-lg bg-black/20 p-1.5 pr-3 hover:bg-black/30"
							>
								{#if post.thumbnailUrl}
									<img
										src={cacheUrl(post.thumbnailUrl)}
										alt=""
										loading="lazy"
										class="size-12 shrink-0 rounded-md bg-white/5 object-cover"
									/>
								{/if}
								<span class="min-w-0 flex-1 truncate text-sm">{post.title}</span>
								<span class="text-base-content/40 shrink-0 text-xs">Another thread</span>
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	</li>
{/snippet}

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
	{#if loading}
		<div class="flex flex-col items-center gap-3 py-24">
			<span class="loading loading-spinner text-base-content/40"></span>
			<p class="text-base-content/40 text-sm">Reading the thread…</p>
		</div>
	{:else if errorText}
		<p class="text-error py-24 text-center text-sm">{errorText}</p>
	{:else if thread}
		<div class="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
			<!-- The question -->
			<section class="flex flex-col gap-3 lg:sticky lg:top-20 lg:self-start">
				{#if thread.videoUrl}
					<!-- svelte-ignore a11y_media_has_caption -->
					<video
						src={thread.videoUrl}
						poster={thread.imageUrl ? cacheUrl(thread.imageUrl) : undefined}
						class="max-h-[70vh] w-full rounded-xl bg-black object-contain"
						autoplay
						loop
						muted
						playsinline
						controls
					></video>
				{:else if thread.imageUrl}
					<img
						src={cacheUrl(thread.imageUrl)}
						alt={thread.title}
						class="max-h-[70vh] w-full rounded-xl bg-black object-contain"
					/>
				{/if}
				<h1 class="text-xl font-bold">{thread.title}</h1>
				{#if thread.description}
					<p class="text-base-content/70 text-sm whitespace-pre-line">{thread.description}</p>
				{/if}
				<p class="text-base-content/50 flex flex-wrap items-center gap-x-2 text-xs">
					{#if thread.solved}
						<span class="bg-primary rounded px-1.5 py-0.5 font-bold text-black">Answered</span>
					{/if}
					{#if thread.author}<span>asked by {thread.author}</span>{/if}
					{#if thread.date}<span>{formatRelativeTime(thread.date)}</span>{/if}
				</p>
				{#if thread.tags.length > 0}
					<div class="flex flex-wrap gap-1.5">
						{#each thread.tags as tag (tag)}
							<span class="bg-base-200 rounded-md px-2 py-0.5 text-xs">{tag}</span>
						{/each}
					</div>
				{/if}
				<a
					href={thread.url}
					target="_blank"
					rel="noopener noreferrer"
					class="text-base-content/50 hover:text-base-content flex w-fit items-center gap-1.5 text-sm underline"
				>
					View on {linkLabel(thread.url).split('/')[0]}
					<ArrowSquareOutIcon size={14} />
				</a>
			</section>

			<!-- The answers -->
			<section class="flex flex-col gap-4" aria-label="Comments">
				{#if watchable.length > 0}
					<div class="border-base-300 flex flex-col gap-2 rounded-2xl border p-3">
						<h2 class="text-sm font-semibold">Watch in the app</h2>
						<ul class="flex flex-col gap-1.5">
							{#each watchable as link (link.url)}
								{@render linkRow(link)}
							{/each}
						</ul>
					</div>
				{/if}
				<h2 class="flex items-baseline gap-2 text-lg font-semibold">
					Comments
					<span class="text-base-content/40 text-sm font-normal">{thread.comments.length}</span>
				</h2>
				{#if thread.comments.length === 0}
					<p class="text-base-content/40 text-sm">Nobody has answered yet.</p>
				{:else}
					<ul class="flex flex-col gap-4">
						{#each thread.comments as comment (comment.id)}
							{@render commentBubble(comment)}
						{/each}
					</ul>
				{/if}
			</section>
		</div>
	{/if}
</div>

<style>
	.accepted-answer {
		box-shadow: inset 0 0 0 1px var(--color-primary);
	}

	.thread-chip {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.25rem 0.65rem;
		border-radius: 0.5rem;
		background-color: var(--color-base-300);
		font-size: 13px;
		font-weight: 500;
		transition: background-color 150ms;
	}

	.thread-chip:hover {
		background-color: color-mix(in oklab, var(--color-base-content) 20%, transparent);
	}
</style>
