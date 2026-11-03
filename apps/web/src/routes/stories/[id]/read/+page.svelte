<!--
	The story reader: one chapter at a time over the whole screen, in the
	theme, typeface, size, spacing and width set in its settings (kept for
	every story). It opens at ?chapter=<position>, else where reading
	stopped, and keeps that place (chapter and how far into it) as the
	story's progress. A chapter ends with the next one, or in a branching
	story with its choices; reading a chapter finds the choices it offers.
	Concentration mode goes fullscreen, hides the toolbars until the pointer
	nears an edge or the text is tapped, and dims all but the paragraph
	being read. Keys: ←/→ chapters, f concentration, Esc closes.
-->
<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import CornersInIcon from 'phosphor-svelte/lib/CornersInIcon';
	import CornersOutIcon from 'phosphor-svelte/lib/CornersOutIcon';
	import ListBulletsIcon from 'phosphor-svelte/lib/ListBulletsIcon';
	import TextAaIcon from 'phosphor-svelte/lib/TextAaIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ReaderSettings from '$lib/components/ReaderSettings.svelte';
	import StoryContents from '$lib/components/StoryContents.svelte';
	import {
		fetchChapters,
		openStory,
		readChapter,
		type ChapterRow,
		type ChapterText,
		type StoryDetail
	} from '$lib/media';
	import { columnEms, fontStack, readerSettings } from '$lib/stores/readerSettings';
	import { storyReaderUrl, storyUrl } from '$lib/routes';
	import { logEvent } from '$lib/events';
	import {
		followingChapters,
		isBranching,
		minutesLeft,
		paragraphsOf,
		pathTo,
		previousChapter,
		progressOf,
		readingPointOf,
		type ReadingPoint
	} from '$lib/storyReading';

	// How long scrolling rests before the place is saved.
	const SAVE_DELAY_MS = 1500;
	// How long called-back toolbars stay in concentration mode.
	const CHROME_SHOWN_MS = 2500;
	// Pointer this close to the top or bottom edge calls the toolbars back (px).
	const EDGE_PX = 72;
	// The paragraph being read is the one crossing this share of the screen's height.
	const READING_LINE = 0.35;

	let story = $state<StoryDetail | null>(null);
	let chapters = $state<ChapterRow[]>([]);
	let chapter = $state<ChapterText | null>(null);
	let loadingChapter = $state(false);
	let error = $state<string | null>(null);
	let settingsOpen = $state(false);
	let contentsOpen = $state(false);
	let concentrating = $state(false);
	let chromeShown = $state(false);
	let fraction = $state(0);
	let readingIndex = $state(0);
	let scroller = $state<HTMLElement>();
	let article = $state<HTMLElement>();

	// The last place saved, to reopen a chapter where it was left.
	let lastPoint: ReadingPoint | null = null;
	let saveTimer: ReturnType<typeof setTimeout> | null = null;
	let chromeTimer: ReturnType<typeof setTimeout> | null = null;
	let scrollFrame = 0;
	// The chapter being read from the site, so it isn't asked for twice.
	let openingPosition: number | null = null;
	// Whether concentration mode made the page fullscreen, so leaving fullscreen leaves it.
	let wentFullscreen = false;

	let id = $derived(page.params.id ?? '');
	let requestedPosition = $derived(page.url.searchParams.get('chapter'));
	let targetPosition = $derived.by(() => {
		if (requestedPosition !== null) return Number(requestedPosition);
		if (!story) return null;
		const point = readingPointOf(story.userState.progress);
		if (point && chapters.some((entry) => entry.position === point.position)) return point.position;
		return chapters[0]?.position ?? null;
	});
	let current = $derived(chapters.find((entry) => entry.position === chapter?.position) ?? null);
	let branching = $derived(isBranching(chapters));
	let following = $derived(current ? followingChapters(chapters, current) : []);
	let previous = $derived(current ? previousChapter(chapters, current) : null);
	let path = $derived(current ? pathTo(chapters, current) : []);
	let paragraphs = $derived(chapter ? paragraphsOf(chapter.text) : []);
	let chapterNumber = $derived(current ? chapters.indexOf(current) + 1 : 0);
	let chapterLabel = $derived.by(() => {
		if (!current) return '';
		if (branching) return path.length === 0 ? 'Beginning' : `Choice ${path.length}`;
		return `Chapter ${chapterNumber} of ${chapters.length}`;
	});
	let chromeConcealed = $derived(concentrating && !chromeShown && !settingsOpen && !contentsOpen);

	$effect(() => {
		void loadStory(id);
	});

	$effect(() => {
		const position = targetPosition;
		if (
			!story ||
			position === null ||
			position === chapter?.position ||
			position === openingPosition
		)
			return;
		void openChapter(story.id, position);
	});

	onDestroy(() => {
		saveNow();
		if (chromeTimer) clearTimeout(chromeTimer);
		cancelAnimationFrame(scrollFrame);
		if (wentFullscreen && document.fullscreenElement)
			void document.exitFullscreen().catch(() => {});
	});

	/** Opens the story: its chapters and where reading stopped. */
	async function loadStory(storyId: string) {
		story = null;
		chapter = null;
		chapters = [];
		error = null;
		try {
			const found = await openStory(storyId);
			if (storyId !== id) return;
			if (!found) {
				error = "This story isn't in the library.";
				return;
			}
			lastPoint = readingPointOf(found.userState.progress);
			chapters = found.chapters;
			story = found;
			if (found.chapters.length === 0) error = 'The site listed no chapters for this story.';
		} catch (failure) {
			error = String(failure);
		}
	}

	/** Reads a chapter (from the site the first time) and opens it where it was left. */
	async function openChapter(storyId: string, position: number) {
		saveNow();
		openingPosition = position;
		loadingChapter = true;
		error = null;
		try {
			const read = await readChapter(storyId, position);
			if (storyId !== id || targetPosition !== position) return;
			if (!read) {
				error = 'This chapter is gone from the story.';
				return;
			}
			chapter = read;
			// Reading a chapter of a branching story can find the choices it offers.
			if (isBranching(chapters) || read.parent) chapters = await fetchChapters(storyId);
			await restorePlace(position);
		} catch (failure) {
			error = String(failure);
		} finally {
			if (openingPosition === position) openingPosition = null;
			loadingChapter = false;
		}
	}

	/** Scrolls to where the chapter was left, else its top, and saves that place. */
	async function restorePlace(position: number) {
		await tick();
		if (!scroller) return;
		const resumed = lastPoint?.position === position ? lastPoint.fraction : 0;
		scroller.scrollTop = resumed * Math.max(scroller.scrollHeight - scroller.clientHeight, 0);
		scroller.focus({ preventScroll: true });
		measure();
		save({ position, fraction: resumed });
	}

	/** Measures the place while scrolling, once per frame, and saves it once scrolling rests. */
	function onScroll() {
		cancelAnimationFrame(scrollFrame);
		scrollFrame = requestAnimationFrame(measure);
		if (saveTimer) clearTimeout(saveTimer);
		saveTimer = setTimeout(saveNow, SAVE_DELAY_MS);
	}

	/** How far into the chapter the reader is, and which paragraph is being read. */
	function measure() {
		if (!scroller || !article) return;
		const scrollable = scroller.scrollHeight - scroller.clientHeight;
		fraction = scrollable > 0 ? scroller.scrollTop / scrollable : 1;
		const line = scroller.getBoundingClientRect().top + scroller.clientHeight * READING_LINE;
		const elements = article.querySelectorAll('[data-paragraph]');
		const index = Array.from(elements).findIndex(
			(element) => element.getBoundingClientRect().bottom > line
		);
		readingIndex = index === -1 ? elements.length - 1 : index;
	}

	/** Saves the current place right away. */
	function saveNow() {
		if (saveTimer) clearTimeout(saveTimer);
		saveTimer = null;
		if (!chapter || !story) return;
		save({ position: chapter.position, fraction });
	}

	/** Saves a place as the story's progress, unless it's the one saved last. */
	function save(point: ReadingPoint) {
		if (!story) return;
		const same =
			lastPoint?.position === point.position &&
			Math.abs(lastPoint.fraction - point.fraction) < 0.005;
		lastPoint = point;
		if (same) return;
		logEvent({
			kind: 'progress',
			subject: { kind: 'story', id: story.id },
			data: { position: progressOf(point) }
		});
	}

	/** Goes to a chapter; the browser's back button returns to this one. */
	function openPosition(position: number) {
		contentsOpen = false;
		goto(storyReaderUrl(id, position), { noScroll: true, keepFocus: true });
	}

	/** Goes on when exactly one chapter follows. */
	function goNext() {
		if (following.length !== 1) return;
		openPosition(following[0].position);
	}

	/** Goes back one chapter (in a branching story, to the choice before). */
	function goPrevious() {
		if (!previous) return;
		openPosition(previous.position);
	}

	/** Enters or leaves concentration mode, fullscreen with it when the browser lets it. */
	function toggleConcentration() {
		concentrating = !concentrating;
		settingsOpen = false;
		contentsOpen = false;
		if (concentrating) {
			chromeShown = false;
			if (!document.fullscreenElement && document.documentElement.requestFullscreen) {
				document.documentElement
					.requestFullscreen()
					.then(() => (wentFullscreen = true))
					.catch(() => {});
			}
			return;
		}
		if (wentFullscreen && document.fullscreenElement)
			void document.exitFullscreen().catch(() => {});
		wentFullscreen = false;
	}

	/** Leaving fullscreen (the browser takes Esc for that) leaves concentration mode too. */
	function onFullscreenChange() {
		if (document.fullscreenElement || !wentFullscreen) return;
		wentFullscreen = false;
		concentrating = false;
	}

	/** Shows the toolbars for a moment in concentration mode. */
	function revealChrome() {
		chromeShown = true;
		if (chromeTimer) clearTimeout(chromeTimer);
		chromeTimer = setTimeout(() => (chromeShown = false), CHROME_SHOWN_MS);
	}

	/** The pointer near the top or bottom edge calls the toolbars back. */
	function onPointerMove(event: PointerEvent) {
		if (!concentrating || event.pointerType === 'touch') return;
		if (event.clientY < EDGE_PX || event.clientY > window.innerHeight - EDGE_PX) revealChrome();
	}

	/** A tap on the text shows or hides the toolbars in concentration mode. */
	function onTextClick(event: MouseEvent) {
		if (!concentrating || (event.target as HTMLElement).closest('a, button')) return;
		if (chromeShown) {
			chromeShown = false;
			return;
		}
		revealChrome();
	}

	/** Reader keys; not while typing. */
	function onKeydown(event: KeyboardEvent) {
		const target = event.target as HTMLElement | null;
		if (target?.closest('input, textarea, select, [contenteditable]')) return;
		if (event.ctrlKey || event.metaKey || event.altKey) return;
		if (event.key === 'ArrowRight') goNext();
		if (event.key === 'ArrowLeft') goPrevious();
		if (event.key === 'f') toggleConcentration();
		if (event.key === 'Escape') closeSomething();
	}

	/** Esc closes an open panel, else leaves concentration mode. */
	function closeSomething() {
		if (settingsOpen || contentsOpen) {
			settingsOpen = false;
			contentsOpen = false;
			return;
		}
		if (concentrating) toggleConcentration();
	}
</script>

<svelte:head><title>{chapter?.title ?? story?.title ?? 'Reader'} · Veil</title></svelte:head>
<svelte:window onkeydown={onKeydown} />
<svelte:document onfullscreenchange={onFullscreenChange} />

<div
	class="reader z-modal fixed inset-0 flex flex-col"
	data-reader-theme={$readerSettings.theme}
	onpointermove={onPointerMove}
	role="presentation"
>
	<header
		class="reader-chrome flex h-14 shrink-0 items-center gap-2 border-b px-3"
		class:concealed={chromeConcealed}
		style:border-color="var(--reader-line)"
		style:font-family="var(--font-sans)"
	>
		<a
			href={storyUrl(id)}
			class="flex h-10 w-10 items-center justify-center rounded-lg hover:bg-current/5"
			aria-label="Back to the story"><ArrowLeftIcon size={20} /></a
		>
		<div class="flex min-w-0 flex-1 flex-col leading-tight">
			<span class="truncate text-sm font-medium">{story?.title ?? ''}</span>
			<span class="truncate text-xs" style:color="var(--reader-muted)">{chapterLabel}</span>
		</div>
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-lg hover:bg-current/5"
			onclick={() => {
				contentsOpen = !contentsOpen;
				settingsOpen = false;
			}}
			aria-label="Chapters"
			aria-pressed={contentsOpen}><ListBulletsIcon size={20} /></button
		>
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-lg hover:bg-current/5"
			onclick={() => {
				settingsOpen = !settingsOpen;
				contentsOpen = false;
			}}
			aria-label="Display settings"
			aria-pressed={settingsOpen}><TextAaIcon size={20} /></button
		>
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-lg hover:bg-current/5"
			onclick={toggleConcentration}
			aria-label={concentrating ? 'Leave concentration mode' : 'Concentration mode'}
			aria-pressed={concentrating}
		>
			{#if concentrating}
				<CornersInIcon size={20} />
			{:else}
				<CornersOutIcon size={20} />
			{/if}
		</button>
	</header>

	<div
		bind:this={scroller}
		class="min-h-0 flex-1 overflow-y-auto outline-none"
		tabindex="-1"
		onscroll={onScroll}
		onclick={onTextClick}
		role="presentation"
	>
		<article
			bind:this={article}
			lang="en"
			class="mx-auto px-6 pt-12 pb-32"
			style:max-width="{columnEms($readerSettings.width)}em"
			style:font-family={fontStack($readerSettings.font)}
			style:font-size="{$readerSettings.fontSize}px"
			style:line-height={$readerSettings.lineHeight}
		>
			{#if error}
				<p class="text-center" style:color="var(--reader-muted)">{error}</p>
			{/if}

			{#if chapter}
				<header class="mb-10 flex flex-col gap-2 text-center">
					{#if branching && path.length > 0}
						<p class="text-[0.7em] tracking-wide uppercase" style:color="var(--reader-muted)">
							{path.map((step) => step.choice ?? step.title).join(' › ')}
						</p>
					{:else if !branching}
						<p class="text-[0.7em] tracking-widest uppercase" style:color="var(--reader-muted)">
							Chapter {chapterNumber}
						</p>
					{/if}
					<h1 class="text-[1.5em] leading-snug font-semibold">
						{chapter.title ?? story?.title ?? ''}
					</h1>
				</header>

				<div class="reader-text" class:concentrating>
					{#each paragraphs as paragraph, index (index)}
						<p data-paragraph class:reading={index === readingIndex}>{paragraph}</p>
					{/each}
				</div>

				{#if paragraphs.length === 0 && !loadingChapter}
					<p class="text-center" style:color="var(--reader-muted)">This chapter has no text.</p>
				{/if}

				<footer
					class="mt-16 flex flex-col items-center gap-4 border-t pt-10"
					style:border-color="var(--reader-line)"
					style:font-family="var(--font-sans)"
				>
					{#if branching && following.length > 0}
						<p class="text-sm" style:color="var(--reader-muted)">What happens next?</p>
						<div class="flex w-full max-w-md flex-col gap-2">
							{#each following as choice (choice.id)}
								<button
									type="button"
									class="rounded-xl px-4 py-3 text-left text-[0.85em] leading-snug"
									style:background-color="var(--reader-surface)"
									onclick={() => openPosition(choice.position)}
								>
									{choice.choice ?? choice.title ?? 'Go on'}
								</button>
							{/each}
						</div>
					{:else if following.length === 1}
						<button
							type="button"
							class="flex items-center gap-2 rounded-xl px-5 py-3 text-sm font-medium"
							style:background-color="var(--reader-surface)"
							onclick={goNext}
						>
							Next: {following[0].title ?? `Chapter ${chapterNumber + 1}`}
							<CaretRightIcon size={16} />
						</button>
					{:else if branching}
						<p class="text-sm" style:color="var(--reader-muted)">The end of this path.</p>
					{:else if story?.complete}
						<p class="text-sm" style:color="var(--reader-muted)">The end.</p>
					{:else}
						<p class="text-sm" style:color="var(--reader-muted)">
							The author hasn't written the next chapter yet.
						</p>
					{/if}
					{#if previous}
						<button
							type="button"
							class="flex items-center gap-1 text-sm"
							style:color="var(--reader-muted)"
							onclick={goPrevious}
						>
							<CaretLeftIcon size={14} />
							{branching ? 'Back to the choice before' : 'Previous chapter'}
						</button>
					{/if}
				</footer>
			{:else if !error}
				<p class="text-center text-sm" style:color="var(--reader-muted)">Reading the chapter…</p>
			{/if}
		</article>
	</div>

	<footer
		class="reader-chrome flex h-10 shrink-0 items-center gap-3 px-4 text-xs"
		class:concealed={chromeConcealed}
		style:color="var(--reader-muted)"
		style:font-family="var(--font-sans)"
	>
		<div
			class="h-1 flex-1 overflow-hidden rounded-full"
			style:background-color="var(--reader-line)"
		>
			<div
				class="h-full rounded-full"
				style:width="{Math.round(fraction * 100)}%"
				style:background-color="var(--reader-accent)"
			></div>
		</div>
		{#if chapter}
			<span class="tabular-nums">
				{#if loadingChapter}
					Loading…
				{:else if fraction >= 0.995}
					Chapter done
				{:else}
					{minutesLeft(chapter.wordCount, fraction)} min left in chapter
				{/if}
			</span>
		{/if}
	</footer>

	{#if settingsOpen || contentsOpen}
		<button
			type="button"
			class="absolute inset-0 top-14 cursor-default bg-black/30"
			onclick={closeSomething}
			aria-label="Close"
		></button>
		<aside
			class="absolute top-14 right-0 bottom-0 flex w-96 max-w-full flex-col overflow-hidden border-l shadow-2xl"
			style:background-color="var(--reader-bg)"
			style:border-color="var(--reader-line)"
			style:font-family="var(--font-sans)"
		>
			<div class="flex items-center justify-between px-4 pt-4 pb-2">
				<h2 class="text-sm font-semibold">{settingsOpen ? 'Display' : 'Chapters'}</h2>
				<button
					type="button"
					class="flex h-8 w-8 items-center justify-center rounded-lg hover:bg-current/5"
					onclick={closeSomething}
					aria-label="Close"><XIcon size={16} /></button
				>
			</div>
			<div class="min-h-0 flex-1 overflow-y-auto px-4 pb-6">
				{#if settingsOpen}
					<ReaderSettings />
				{:else}
					<StoryContents
						storyId={id}
						{chapters}
						currentId={current?.id}
						onpick={() => (contentsOpen = false)}
					/>
				{/if}
			</div>
		</aside>
	{/if}
</div>
