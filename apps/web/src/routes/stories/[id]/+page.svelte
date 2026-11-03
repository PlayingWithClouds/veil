<!--
	One story. Opening it reads its page when that wasn't read yet (chapters,
	author, tags). A banner over the blurred cover shows the cover, title,
	author, length, description and the buttons to continue or start over.
	Below, a branching story shows the path read so far and, folded away,
	its map of every choice found; a linear one its numbered chapters. Tags
	are folded away at the end. Beside it on wide screens (below on narrow
	ones): more by the author and stories sharing its tags.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import GitForkIcon from 'phosphor-svelte/lib/GitForkIcon';
	import ArrowCounterClockwiseIcon from 'phosphor-svelte/lib/ArrowCounterClockwiseIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import StoryCard from '$lib/components/StoryCard.svelte';
	import StoryContents from '$lib/components/StoryContents.svelte';
	import StoryPath from '$lib/components/StoryPath.svelte';
	import {
		fetchSimilarStories,
		fetchStoriesBy,
		openStory,
		type StoryCardRow,
		type StoryDetail
	} from '$lib/media';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, storyReaderUrl, storyUrl, studioUrl } from '$lib/routes';
	import { logEvent } from '$lib/events';
	import { isBranching, minutesLeft, pathTo, readingPointOf } from '$lib/storyReading';

	// Stories in each side list.
	const SIDE_STORIES = 8;

	let story = $state<StoryDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let fullDetails = $state(false);
	let byAuthor = $state<StoryCardRow[]>([]);
	let similar = $state<StoryCardRow[]>([]);

	let id = $derived(page.params.id ?? '');
	let cover = $derived(cacheUrl(story?.cover?.uri));
	let branching = $derived(story ? isBranching(story.chapters) : false);
	let author = $derived(story?.performers[0] ?? null);
	let readingPoint = $derived(readingPointOf(story?.userState.progress));
	// The chapter reading stopped in, when it's still there.
	let resumeChapter = $derived.by(() => {
		if (!story || !readingPoint) return null;
		const position = readingPoint.position;
		return story.chapters.find((chapter) => chapter.position === position) ?? null;
	});
	let resumeLabel = $derived.by(() => {
		if (!story || !resumeChapter || !readingPoint) return '';
		if (branching) {
			const choices = pathTo(story.chapters, resumeChapter).length;
			return choices === 0 ? 'At the beginning' : `Choice ${choices} of your path`;
		}
		const number = story.chapters.indexOf(resumeChapter) + 1;
		const percent = Math.round(readingPoint.fraction * 100);
		return `Chapter ${number} of ${story.chapters.length} · ${percent}% read`;
	});
	let words = $derived(
		story?.wordCount ?? story?.chapters.reduce((sum, chapter) => sum + chapter.wordCount, 0) ?? 0
	);

	$effect(() => {
		void load(id);
	});

	/** Opens the story, reading its page when needed, and the stories to show beside it. */
	async function load(storyId: string) {
		story = null;
		byAuthor = [];
		similar = [];
		error = null;
		loading = true;
		fullDetails = false;
		void fetchSimilarStories(storyId, SIDE_STORIES).then((found) => {
			if (storyId === id) similar = found;
		});
		try {
			const found = await openStory(storyId);
			if (storyId !== id) return;
			story = found;
			if (!found) return;
			logEvent({ kind: 'view', subject: { kind: 'story', id: found.id } });
			void loadByAuthor(found);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** The author's other stories. */
	async function loadByAuthor(found: StoryDetail) {
		const authorId = found.performers[0]?.id;
		if (!authorId) return;
		const stories = await fetchStoriesBy(authorId, SIDE_STORIES + 1);
		if (found.id !== id) return;
		byAuthor = stories.filter((other) => other.id !== found.id).slice(0, SIDE_STORIES);
	}
</script>

<svelte:head><title>{story?.title ?? 'Story'} · Veil</title></svelte:head>

<div class="mx-auto grid w-full max-w-[1700px] gap-8 pt-2 xl:grid-cols-[minmax(0,1fr)_380px]">
	<div class="flex min-w-0 flex-col gap-8">
		{#if loading && !story}
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
			<section class="bg-base-200 relative overflow-hidden rounded-3xl">
				{#if cover}
					<img
						src={cover}
						alt=""
						aria-hidden="true"
						class="absolute inset-0 h-full w-full scale-125 object-cover opacity-40 blur-3xl"
					/>
				{/if}
				<div
					class="from-base-200 via-base-200/85 to-base-200/50 absolute inset-0 bg-linear-to-r"
				></div>

				<div class="relative flex gap-8 p-8 max-md:flex-col max-md:gap-5 max-md:p-5">
					<div
						class="bg-base-300 flex aspect-[2/3] w-52 shrink-0 items-center justify-center self-start overflow-hidden rounded-xl shadow-2xl max-md:w-36"
					>
						{#if cover}
							<img src={cover} alt={story.title} class="h-full w-full object-cover" />
						{:else}
							<BookOpenTextIcon size={56} class="text-base-content/20" />
						{/if}
					</div>

					<div class="flex min-w-0 flex-1 flex-col gap-4">
						<div class="flex flex-col gap-1.5">
							<h1 class="text-3xl leading-tight font-bold tracking-tight max-md:text-2xl">
								{story.title}
							</h1>
							{#if author}
								<p class="text-base-content/70">
									by
									<a
										href={performerUrl(author.id)}
										class="text-base-content font-medium hover:underline">{author.name}</a
									>
									{#if story.studio}
										·
										<a href={studioUrl(story.studio.id)} class="hover:underline"
											>{story.studio.name}</a
										>
									{/if}
								</p>
							{/if}
						</div>

						<div class="flex flex-wrap items-center gap-2 text-xs">
							<span class="bg-base-content/10 rounded-full px-3 py-1">
								{story.chapters.length}
								{story.chapters.length === 1 ? 'chapter' : 'chapters'}
								{#if branching}found{/if}
							</span>
							{#if branching}
								<span class="bg-base-content/10 flex items-center gap-1 rounded-full px-3 py-1">
									<GitForkIcon size={13} /> Interactive
								</span>
							{:else if story.complete}
								<span class="bg-base-content/10 rounded-full px-3 py-1">Complete</span>
							{:else}
								<span class="bg-base-content/10 rounded-full px-3 py-1">Ongoing</span>
							{/if}
							<!-- A branching story's length is only what's been read of it. -->
							{#if words > 0 && !branching}
								<span class="bg-base-content/10 rounded-full px-3 py-1">
									{words.toLocaleString()} words · {minutesLeft(words, 0)} min
								</span>
							{/if}
							{#if story.date}
								<span class="bg-base-content/10 rounded-full px-3 py-1">{story.date}</span>
							{/if}
							{#each story.sources as source (source.url)}
								<a
									href={source.url}
									target="_blank"
									rel="noreferrer"
									class="bg-base-content/10 hover:bg-base-content/20 flex items-center gap-1 rounded-full px-3 py-1"
								>
									{siteDisplayName(source.plugin, $sites)}
									<ArrowSquareOutIcon size={12} />
								</a>
							{/each}
						</div>

						{#if story.details}
							<div class="flex max-w-prose flex-col items-start gap-1">
								<p
									class="text-base-content/85 leading-relaxed whitespace-pre-line"
									class:line-clamp-4={!fullDetails}
								>
									{story.details}
								</p>
								{#if story.details.length > 280}
									<button
										type="button"
										class="text-base-content/60 hover:text-base-content text-sm"
										onclick={() => (fullDetails = !fullDetails)}
									>
										{fullDetails ? 'Show less' : 'Show more'}
									</button>
								{/if}
							</div>
						{/if}

						{#if story.chapters.length > 0}
							<div class="mt-auto flex flex-wrap items-center gap-3 pt-2">
								{#if resumeChapter}
									<a
										href={storyReaderUrl(story.id)}
										class="bg-primary text-primary-content flex items-center gap-3 rounded-xl px-5 py-2.5"
									>
										<BookOpenTextIcon size={20} />
										<span class="flex flex-col leading-tight">
											<span class="font-semibold">Continue reading</span>
											<span class="text-xs opacity-70">{resumeLabel}</span>
										</span>
									</a>
									<a
										href={storyReaderUrl(story.id, story.chapters[0].position)}
										class="bg-base-content/10 hover:bg-base-content/20 flex items-center gap-2 rounded-xl px-4 py-3 text-sm font-medium"
									>
										<ArrowCounterClockwiseIcon size={16} /> Start over
									</a>
								{:else}
									<a
										href={storyReaderUrl(story.id, story.chapters[0].position)}
										class="bg-primary text-primary-content flex items-center gap-2 rounded-xl px-5 py-3 font-semibold"
									>
										<BookOpenTextIcon size={20} /> Start reading
									</a>
								{/if}
							</div>
						{/if}
					</div>
				</div>
			</section>

			{#if story.chapters.length === 0 && !loading}
				<p class="text-base-content/50 text-sm">The site listed no chapters for this story.</p>
			{:else if branching}
				<section class="flex flex-col gap-4">
					<h2 class="text-lg font-semibold">Your path</h2>
					<StoryPath storyId={story.id} chapters={story.chapters} current={resumeChapter} />
				</section>

				<details class="group bg-base-200 rounded-2xl">
					<summary
						class="flex cursor-pointer list-none items-center justify-between gap-3 px-5 py-4"
					>
						<span class="flex flex-col">
							<span class="font-semibold">Story map</span>
							<span class="text-base-content/50 text-sm">
								Every choice found so far; reading a chapter finds the ones it offers.
							</span>
						</span>
						<CaretDownIcon size={16} class="shrink-0 transition-transform group-open:rotate-180" />
					</summary>
					<div class="px-3 pb-4">
						<StoryContents
							storyId={story.id}
							chapters={story.chapters}
							currentId={resumeChapter?.id}
						/>
					</div>
				</details>
			{:else}
				<section class="flex flex-col gap-3">
					<h2 class="text-lg font-semibold">Chapters</h2>
					<div class="bg-base-200 rounded-2xl p-2">
						<StoryContents
							storyId={story.id}
							chapters={story.chapters}
							currentId={resumeChapter?.id}
						/>
					</div>
				</section>
			{/if}

			{#if story.tags.length > 0}
				<details class="group">
					<summary class="flex cursor-pointer list-none items-center gap-2 text-lg font-semibold">
						Tags
						<span class="text-base-content/50 text-sm font-normal">{story.tags.length}</span>
						<CaretDownIcon size={14} class="transition-transform group-open:rotate-180" />
					</summary>
					<div class="pt-3">
						<TagChips tags={story.tags} />
					</div>
				</details>
			{/if}
		{/if}
	</div>

	{#if byAuthor.length > 0 || similar.length > 0}
		<aside class="flex flex-col gap-8">
			{#if author && byAuthor.length > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-lg font-semibold">More by {author.name}</h2>
					<div
						class="grid grid-cols-2 gap-x-4 gap-y-6 max-xl:grid-cols-[repeat(auto-fill,minmax(150px,1fr))]"
					>
						{#each byAuthor as other (other.id)}
							<StoryCard story={other} onclick={() => goto(storyUrl(other.id))} />
						{/each}
					</div>
				</section>
			{/if}
			{#if similar.length > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-lg font-semibold">Similar stories</h2>
					<div
						class="grid grid-cols-2 gap-x-4 gap-y-6 max-xl:grid-cols-[repeat(auto-fill,minmax(150px,1fr))]"
					>
						{#each similar as other (other.id)}
							<StoryCard story={other} onclick={() => goto(storyUrl(other.id))} />
						{/each}
					</div>
				</section>
			{/if}
		</aside>
	{/if}
</div>
