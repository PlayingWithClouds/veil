<!--
	One story. Opening it reads its page when that wasn't read yet (chapters,
	characters, tags). Beside the cover: title, author, length, whether it's
	finished, characters, tags, the description and a button to start or
	continue reading. Below, its chapters: a list, or the tree of choices of
	a branching story.
-->
<script lang="ts">
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
	import GitForkIcon from 'phosphor-svelte/lib/GitForkIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import StoryContents from '$lib/components/StoryContents.svelte';
	import { openStory, type StoryDetail } from '$lib/media';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, storyReaderUrl, studioUrl } from '$lib/routes';
	import { logEvent } from '$lib/events';
	import { isBranching, minutesLeft, readingPointOf } from '$lib/storyReading';

	let story = $state<StoryDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let id = $derived(page.params.id ?? '');
	let cover = $derived(cacheUrl(story?.cover?.uri));
	let branching = $derived(story ? isBranching(story.chapters) : false);
	// The chapter reading stopped in, when it's still there.
	let resumeChapter = $derived.by(() => {
		const point = readingPointOf(story?.userState.progress);
		if (!story || !point) return null;
		return story.chapters.find((chapter) => chapter.position === point.position) ?? null;
	});
	let words = $derived(
		story?.wordCount ?? story?.chapters.reduce((sum, chapter) => sum + chapter.wordCount, 0) ?? 0
	);

	$effect(() => {
		void load(id);
	});

	/** Opens the story, reading its page when needed. */
	async function load(storyId: string) {
		story = null;
		error = null;
		loading = true;
		try {
			const found = await openStory(storyId);
			if (storyId !== id) return;
			story = found;
			if (found) logEvent({ kind: 'view', subject: { kind: 'story', id: found.id } });
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>{story?.title ?? 'Story'} · Veil</title></svelte:head>

<div class="mx-auto flex max-w-4xl flex-col gap-8 pt-2">
	{#if loading}
		<p class="text-base-content/50 flex items-center gap-2 text-sm">
			<span class="loading loading-spinner"></span> Reading the page…
		</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if !loading && !story && !error}
		<p class="text-base-content/50 text-sm">This story isn't in the library.</p>
	{/if}

	{#if story}
		<div class="flex gap-6 max-md:flex-col">
			<div
				class="bg-base-200 flex aspect-[2/3] w-48 shrink-0 items-center justify-center overflow-hidden rounded-2xl max-md:w-36"
			>
				{#if cover}
					<img src={cover} alt={story.title} class="h-full w-full object-cover" />
				{:else}
					<BookOpenTextIcon size={56} class="text-base-content/20" />
				{/if}
			</div>

			<div class="flex min-w-0 flex-1 flex-col gap-3">
				<h1 class="text-2xl font-semibold">{story.title}</h1>
				<div class="text-base-content/60 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
					{#if story.studio}
						<a
							href={studioUrl(story.studio.id)}
							class="text-base-content font-medium hover:underline">{story.studio.name}</a
						>
					{/if}
					{#if story.date}<span>{story.date}</span>{/if}
					<span>
						{story.chapters.length}
						{story.chapters.length === 1 ? 'chapter' : 'chapters'}
					</span>
					{#if words > 0}
						<span>{words.toLocaleString()} words · {minutesLeft(words, 0)} min</span>
					{/if}
					{#if branching}
						<span class="flex items-center gap-1"><GitForkIcon size={14} /> Interactive</span>
					{:else if !story.complete}
						<span>Ongoing</span>
					{/if}
					{#each story.sources as source (source.url)}
						<a
							href={source.url}
							target="_blank"
							rel="noreferrer"
							class="hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={13} />
							{siteDisplayName(source.plugin, $sites)}
						</a>
					{/each}
				</div>
				{#if story.performers.length > 0}
					<div class="flex flex-wrap gap-2 text-sm">
						{#each story.performers as performer (performer.id)}
							<a
								href={performerUrl(performer.id)}
								class="bg-base-200 hover:bg-base-300 rounded-full px-3 py-1">{performer.name}</a
							>
						{/each}
					</div>
				{/if}
				<TagChips tags={story.tags} />
				{#if story.details}
					<p class="text-base-content/80 max-w-prose text-sm leading-relaxed whitespace-pre-line">
						{story.details}
					</p>
				{/if}

				{#if story.chapters.length > 0}
					<div class="mt-2 flex flex-wrap items-center gap-3">
						{#if resumeChapter}
							<a
								href={storyReaderUrl(story.id)}
								class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-lg px-4 text-sm font-medium"
							>
								<BookOpenTextIcon size={18} /> Continue reading
							</a>
							<span class="text-base-content/50 text-sm">
								at {resumeChapter.choice ?? resumeChapter.title ?? 'the chapter you stopped in'}
							</span>
							<a
								href={storyReaderUrl(story.id, story.chapters[0].position)}
								class="text-base-content/70 hover:text-base-content text-sm">Start over</a
							>
						{:else}
							<a
								href={storyReaderUrl(story.id, story.chapters[0].position)}
								class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-lg px-4 text-sm font-medium"
							>
								<BookOpenTextIcon size={18} /> Start reading
							</a>
						{/if}
					</div>
				{/if}
			</div>
		</div>

		{#if story.chapters.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Chapters</h2>
				{#if branching}
					<p class="text-base-content/50 -mt-2 text-sm">
						Each choice opens a branch; reading a chapter finds the choices it offers.
					</p>
				{/if}
				<div class="bg-base-200 rounded-2xl p-2">
					<StoryContents
						storyId={story.id}
						chapters={story.chapters}
						currentId={resumeChapter?.id}
					/>
				</div>
			</section>
		{:else if !loading}
			<p class="text-base-content/50 text-sm">The site listed no chapters for this story.</p>
		{/if}
	{/if}
</div>
